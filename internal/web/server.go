package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"twitchdropsfarmer/internal/config"
	"twitchdropsfarmer/internal/drops"
	"twitchdropsfarmer/internal/twitch"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config       *config.Config
	twitchClient *twitch.Client
	miner        *drops.Miner

	// WebSocket upgrader
	upgrader websocket.Upgrader

	// WebSocket connections
	wsConnections map[*websocket.Conn]bool
	wsBroadcast   chan []byte
	wsRegister    chan *websocket.Conn
	wsUnregister  chan *websocket.Conn

	// Device code storage (in production use Redis/database)
	deviceCodes map[string]*twitch.DeviceCodeResponse

	// Miner context management
	minerCtx    context.Context
	minerCancel context.CancelFunc
}

func NewServer(cfg *config.Config, twitchClient *twitch.Client, miner *drops.Miner) *Server {
	server := &Server{
		config:       cfg,
		twitchClient: twitchClient,
		miner:        miner,
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
	router.Static("/css", "./web/static/css")
	router.Static("/js", "./web/static/js")
	router.Static("/assets", "./web/static/assets")

	// Handle favicon
	router.StaticFile("/favicon.ico", "./web/static/favicon.ico")

	// Serve the Vue.js SPA for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// Check if it's an API route or WebSocket
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		if len(path) >= 3 && path[:3] == "/ws" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		
		// Serve the Vue.js index.html for all other routes (SPA routing)
		c.File("./web/static/index.html")
	})

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

		// Config endpoints (renamed from settings for consistency with Vue frontend)
		config := api.Group("/config")
		{
			config.GET("/", s.getSettings)
			config.POST("/", s.updateSettings)
			config.POST("/game", s.addGameWithSlug)
		}

		// Settings endpoints (keep for backward compatibility)
		settings := api.Group("/settings")
		{
			settings.GET("/", s.getSettings)
			settings.PUT("/", s.updateSettings)
		}

		// Games endpoints (keep for backward compatibility)
		games := api.Group("/games")
		{
			games.POST("/add", s.addGameWithSlug)
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

// Cleanup properly cancels the miner context and closes connections
func (s *Server) Cleanup() {
	// Cancel miner context if it exists
	if s.minerCancel != nil {
		s.minerCancel()
	}
	
	// Close all WebSocket connections
	for conn := range s.wsConnections {
		conn.Close()
	}
}
