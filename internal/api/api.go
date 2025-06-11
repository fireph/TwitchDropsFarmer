package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"twitchdropsfarmer/internal/miner"
	"twitchdropsfarmer/internal/storage"
	"twitchdropsfarmer/internal/twitch"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Handler handles HTTP API requests
type Handler struct {
	storage      *storage.Storage
	twitchClient *twitch.Client
	miner        *miner.Miner
	upgrader     websocket.Upgrader
}

// New creates a new API handler
func New(storage *storage.Storage, twitchClient *twitch.Client, miner *miner.Miner) *Handler {
	return &Handler{
		storage:      storage,
		twitchClient: twitchClient,
		miner:        miner,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// Authentication endpoints

func (h *Handler) GetAuthURL(c *gin.Context) {
	deviceFlow, err := h.twitchClient.StartDeviceFlow()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verificationURI": deviceFlow.VerificationURI,
		"userCode":       deviceFlow.UserCode,
		"deviceCode":     deviceFlow.DeviceCode,
		"expiresIn":      deviceFlow.ExpiresIn,
		"interval":       deviceFlow.Interval,
	})
}

func (h *Handler) HandleAuthCallback(c *gin.Context) {
	var req struct {
		DeviceCode string `json:"deviceCode"`
		Interval   int    `json:"interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Poll for token in a separate goroutine to avoid blocking
	go func() {
		token, err := h.twitchClient.PollForToken(req.DeviceCode, req.Interval)
		if err != nil {
			return
		}

		// Save authentication data
		authData := &storage.AuthData{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
			UpdatedAt:    time.Now(),
		}

		h.storage.SaveAuth(authData)
		h.twitchClient.SetAccessToken(token.AccessToken)
	}()

	c.JSON(http.StatusOK, gin.H{"status": "polling"})
}

func (h *Handler) GetAuthStatus(c *gin.Context) {
	auth := h.storage.GetAuth()
	
	if auth == nil {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
			"debug": "No auth data found",
		})
		return
	}

	isAuthenticated := h.storage.IsAuthenticated()
	
	c.JSON(http.StatusOK, gin.H{
		"authenticated": isAuthenticated,
		"username":      auth.Username,
		"expiresAt":     auth.ExpiresAt,
		"hasToken":      auth.AccessToken != "",
		"debug":         fmt.Sprintf("Auth exists: %v, Expires: %v, Now: %v", isAuthenticated, auth.ExpiresAt, time.Now()),
	})
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.storage.ClearAuth(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.twitchClient.SetAccessToken("")
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// Games management endpoints

func (h *Handler) GetGames(c *gin.Context) {
	games := h.storage.GetGames()
	settings := h.storage.GetSettings()

	// Return games in priority order
	orderedGames := make([]*twitch.Game, 0)
	for _, gameID := range settings.Games {
		if game, exists := games[gameID]; exists {
			orderedGames = append(orderedGames, game)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"games": orderedGames,
	})
}

func (h *Handler) AddGame(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game name is required"})
		return
	}

	// Use the simplified AddGame method that doesn't require search validation
	game, err := h.twitchClient.AddGame(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add game: " + err.Error()})
		return
	}

	// Save the game to storage
	if err := h.storage.AddGame(game); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save game: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game": game,
		"message": fmt.Sprintf("Game '%s' added successfully", game.DisplayName),
	})
}

func (h *Handler) RemoveGame(c *gin.Context) {
	gameID := c.Param("id")

	if err := h.storage.RemoveGame(gameID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (h *Handler) ReorderGames(c *gin.Context) {
	var req struct {
		GameIDs []string `json:"gameIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.ReorderGames(req.GameIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "reordered"})
}

func (h *Handler) GetGameStreams(c *gin.Context) {
	gameNameOrID := c.Param("game")
	limitStr := c.DefaultQuery("limit", "20")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}
	
	if limit > 50 {
		limit = 50
	}

	// The updated GetStreamsForGame now handles slug resolution automatically
	streams, err := h.twitchClient.GetStreamsForGame(gameNameOrID, limit)
	if err != nil {
		// Check if it's a slug resolution issue
		if strings.Contains(err.Error(), "game not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Game not found: " + gameNameOrID,
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get streams: " + err.Error(),
			"game": gameNameOrID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game": gameNameOrID,
		"streamCount": len(streams),
		"streams": streams,
	})
}

// Drops endpoints

func (h *Handler) GetDrops(c *gin.Context) {
	if !h.storage.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	campaigns, err := h.twitchClient.GetDropCampaigns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter for uncompleted drops from games in watch list
	settings := h.storage.GetSettings()
	gameSet := make(map[string]bool)
	for _, gameID := range settings.Games {
		gameSet[gameID] = true
	}

	var availableDrops []twitch.Drop
	for _, campaign := range campaigns {
		if gameSet[campaign.GameID] {
			for _, drop := range campaign.Drops {
				if !drop.IsCompleted {
					availableDrops = append(availableDrops, drop)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"drops": availableDrops,
	})
}

func (h *Handler) ClaimDrop(c *gin.Context) {
	dropID := c.Param("id")

	// Use the new ClaimDrop method instead of TODO
	err := h.twitchClient.ClaimDrop(dropID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to claim drop: " + err.Error(),
			"dropId": dropID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "claimed",
		"dropId": dropID,
	})
}

func (h *Handler) GetCurrentDrop(c *gin.Context) {
	channelID := c.Query("channelId")
	channelLogin := c.Query("channelLogin")
	
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channelId is required"})
		return
	}

	err := h.twitchClient.GetCurrentDrop(channelID, channelLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get current drop: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Check logs for drop progress details",
	})
}

func (h *Handler) ClaimCommunityPoints(c *gin.Context) {
	var req struct {
		ClaimID   string `json:"claimId"`
		ChannelID string `json:"channelId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.twitchClient.ClaimCommunityPoints(req.ClaimID, req.ChannelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to claim community points: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "claimed",
		"claimId": req.ClaimID,
	})
}

// Miner control endpoints

func (h *Handler) StartMiner(c *gin.Context) {
	if !h.storage.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	h.miner.Start()
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

func (h *Handler) StopMiner(c *gin.Context) {
	h.miner.Stop()
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

func (h *Handler) GetMinerStatus(c *gin.Context) {
	status := h.miner.GetStatus()
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetMinerLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	logs := h.miner.GetLogs(limit)
	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}

// Settings endpoints

func (h *Handler) GetSettings(c *gin.Context) {
	settings := h.storage.GetSettings()
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var settings storage.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.SaveSettings(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// WebSocket endpoint for real-time updates

func (h *Handler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Create a channel for this client
	client := make(chan []byte, 256)
	h.miner.AddWebSocketClient(client)
	defer h.miner.RemoveWebSocketClient(client)

	// Send initial status
	status := h.miner.GetStatus()
	initialMessage := map[string]interface{}{
		"type": "status",
		"data": status,
	}
	if data, err := json.Marshal(initialMessage); err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Handle incoming messages and outgoing broadcasts
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
			// Handle incoming WebSocket messages if needed
		}
	}()

	// Send broadcasts to client
	for {
		select {
		case message, ok := <-client:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}