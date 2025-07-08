package web

import (
	"encoding/json"
	"net/http"
	"time"

	"twitchdropsfarmer/internal/config"
	"twitchdropsfarmer/internal/drops"
	"twitchdropsfarmer/internal/storage"
	"twitchdropsfarmer/internal/twitch"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config       *config.Config
	twitchClient *twitch.Client
	miner        *drops.Miner
	db           *storage.Database
	
	// WebSocket upgrader
	upgrader websocket.Upgrader
	
	// WebSocket connections
	wsConnections map[*websocket.Conn]bool
	wsBroadcast   chan []byte
	wsRegister    chan *websocket.Conn
	wsUnregister  chan *websocket.Conn
	
	// Device code storage (in production use Redis/database)
	deviceCodes map[string]*twitch.DeviceCodeResponse
}

func NewServer(cfg *config.Config, twitchClient *twitch.Client, miner *drops.Miner, db *storage.Database) *Server {
	server := &Server{
		config:       cfg,
		twitchClient: twitchClient,
		miner:        miner,
		db:           db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		wsConnections: make(map[*websocket.Conn]bool),
		wsBroadcast:   make(chan []byte),
		wsRegister:    make(chan *websocket.Conn),
		wsUnregister:  make(chan *websocket.Conn),
		deviceCodes:   make(map[string]*twitch.DeviceCodeResponse),
	}

	// Start WebSocket hub
	go server.runWebSocketHub()

	return server
}

func (s *Server) Router() *gin.Engine {
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(SecurityMiddleware())
	router.Use(ErrorHandlingMiddleware())

	// Serve static files
	router.Static("/static", "./web/static")
	router.Static("/css", "./web/static/css")
	router.Static("/js", "./web/static/js")
	
	// Serve HTML pages
	router.StaticFile("/", "./web/static/html/index.html")
	router.StaticFile("/login", "./web/static/html/login.html")
	
	// Handle favicon
	router.StaticFile("/favicon.ico", "./web/static/favicon.ico")

	// API routes
	api := router.Group("/api")
	{
		// Authentication endpoints
		auth := api.Group("/auth")
		{
			auth.GET("/url", s.getAuthURL)
			auth.POST("/callback", s.handleAuthCallback)
			auth.POST("/logout", s.handleLogout)
			auth.GET("/status", s.getAuthStatus)
		}

		// User endpoints
		user := api.Group("/user")
		{
			user.GET("/profile", s.getUserProfile)
			user.GET("/inventory", s.getUserInventory)
		}

		// Campaigns endpoints
		campaigns := api.Group("/campaigns")
		{
			campaigns.GET("/", s.getCampaigns)
			campaigns.GET("/:id", s.getCampaign)
			campaigns.GET("/:id/drops", s.getCampaignDrops)
		}

		// Miner endpoints
		miner := api.Group("/miner")
		{
			miner.GET("/status", s.getMinerStatus)
			miner.GET("/current-drop", s.getCurrentDrop)
			miner.GET("/progress", s.getDropProgress)
			miner.POST("/start", s.startMiner)
			miner.POST("/stop", s.stopMiner)
		}

		// Settings endpoints
		settings := api.Group("/settings")
		{
			settings.GET("/", s.getSettings)
			settings.PUT("/", s.updateSettings)
		}

		// Streams endpoints
		streams := api.Group("/streams")
		{
			streams.GET("/game/:gameId", s.getStreamsForGame)
			streams.GET("/current", s.getCurrentStream)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", s.handleWebSocket)

	return router
}

func (s *Server) runWebSocketHub() {
	// Listen for status updates from miner
	go func() {
		for status := range s.miner.GetStatusChannel() {
			s.broadcastStatus(status)
		}
	}()

	// Handle WebSocket connections
	for {
		select {
		case conn := <-s.wsRegister:
			s.wsConnections[conn] = true
			logrus.Info("WebSocket client connected")

		case conn := <-s.wsUnregister:
			if _, ok := s.wsConnections[conn]; ok {
				delete(s.wsConnections, conn)
				conn.Close()
				logrus.Info("WebSocket client disconnected")
			}

		case message := <-s.wsBroadcast:
			for conn := range s.wsConnections {
				select {
				case <-time.After(time.Second):
					// Write timeout, close connection
					delete(s.wsConnections, conn)
					conn.Close()
				default:
					if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
						delete(s.wsConnections, conn)
						conn.Close()
					}
				}
			}
		}
	}
}

func (s *Server) broadcastStatus(status *drops.MinerStatus) {
	data, err := json.Marshal(map[string]interface{}{
		"type": "status_update",
		"data": status,
	})
	if err != nil {
		logrus.Errorf("Failed to marshal status: %v", err)
		return
	}

	select {
	case s.wsBroadcast <- data:
	default:
		// Channel is full, skip this update
	}
}

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade WebSocket: %v", err)
		return
	}

	s.wsRegister <- conn

	// Handle incoming messages
	go func() {
		defer func() {
			s.wsUnregister <- conn
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("WebSocket error: %v", err)
				}
				break
			}
		}
	}()

	// Send initial status
	status := s.miner.GetStatus()
	s.broadcastStatus(status)
}