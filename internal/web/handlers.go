package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"twitchdropsfarmer/internal/config"
	"twitchdropsfarmer/internal/drops"
	"twitchdropsfarmer/internal/twitch"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Authentication handlers - Device Code Flow (like TDM)
func (s *Server) getAuthURL(c *gin.Context) {
	deviceResp, err := s.twitchClient.StartDeviceFlow(c.Request.Context())
	if err != nil {
		logrus.Errorf("Failed to start device flow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start device flow"})
		return
	}

	// Store device code in memory (in production, use Redis or database)
	s.storeDeviceCode(deviceResp.DeviceCode, deviceResp)

	c.JSON(http.StatusOK, gin.H{
		"device_code":      deviceResp.DeviceCode,
		"user_code":        deviceResp.UserCode,
		"verification_uri": deviceResp.VerificationURI,
		"expires_in":       deviceResp.ExpiresIn,
		"interval":         deviceResp.Interval,
	})
}

func (s *Server) handleAuthCallback(c *gin.Context) {
	var req struct {
		DeviceCode string `json:"device_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("Auth callback binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	logrus.Infof("Received device code: %s", req.DeviceCode)

	deviceResp := s.getDeviceCode(req.DeviceCode)
	if deviceResp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device code"})
		return
	}

	// Start polling in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if err := s.twitchClient.PollForToken(ctx, req.DeviceCode, deviceResp.Interval); err != nil {
			logrus.Errorf("Failed to poll for token: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Polling for authorization...",
	})
}

func (s *Server) handleLogout(c *gin.Context) {
	if err := s.twitchClient.Logout(c.Request.Context()); err != nil {
		logrus.Errorf("Failed to logout: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) getAuthStatus(c *gin.Context) {
	isLoggedIn := s.twitchClient.IsLoggedIn()
	var user interface{} = nil

	if isLoggedIn {
		user = s.twitchClient.GetUser()
	}

	c.JSON(http.StatusOK, gin.H{
		"is_logged_in": isLoggedIn,
		"user":         user,
	})
}

// User handlers
func (s *Server) getUserProfile(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	user := s.twitchClient.GetUser()
	c.JSON(http.StatusOK, user)
}

func (s *Server) getUserInventory(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	inventory, err := s.twitchClient.GetInventory(c.Request.Context())
	if err != nil {
		logrus.Errorf("Failed to get inventory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

// Campaign handlers
func (s *Server) getCampaigns(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	campaigns, err := s.twitchClient.GetDropCampaigns(c.Request.Context())
	if err != nil {
		logrus.Errorf("Failed to get campaigns: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaigns"})
		return
	}

	c.JSON(http.StatusOK, campaigns)
}

func (s *Server) getCampaign(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	campaignID := c.Param("id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign ID is required"})
		return
	}

	campaigns, err := s.twitchClient.GetDropCampaigns(c.Request.Context())
	if err != nil {
		logrus.Errorf("Failed to get campaigns: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get campaigns"})
		return
	}

	for _, campaign := range campaigns {
		if campaign.ID == campaignID {
			c.JSON(http.StatusOK, campaign)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
}

func (s *Server) getCampaignDrops(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	campaignID := c.Param("id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign ID is required"})
		return
	}

	// Get drops from current miner status
	status := s.miner.GetStatus()
	var drops []drops.ActiveDrop
	
	// Find drops for the requested campaign
	for _, drop := range status.ActiveDrops {
		if status.CurrentCampaign != nil && status.CurrentCampaign.ID == campaignID {
			drops = append(drops, drop)
		}
	}

	c.JSON(http.StatusOK, drops)
}

// Miner handlers
func (s *Server) getMinerStatus(c *gin.Context) {
	status := s.miner.GetStatus()
	c.JSON(http.StatusOK, status)
}

func (s *Server) getCurrentDrop(c *gin.Context) {
	status := s.miner.GetStatus()

	if !status.IsRunning {
		c.JSON(http.StatusOK, gin.H{
			"is_running":   false,
			"current_drop": nil,
			"message":      "Miner is not running",
		})
		return
	}

	// Find the active drop being farmed (first unclaimed drop from current campaign)
	var currentDrop *drops.ActiveDrop
	if status.CurrentCampaign != nil {
		for _, drop := range status.ActiveDrops {
			if drop.GameName == status.CurrentCampaign.Game.Name && !drop.IsClaimed {
				currentDrop = &drop
				break
			}
		}

		// Use basic campaign data for now (no CampaignDetails call)
		if currentDrop == nil && status.CurrentCampaign != nil {
			if len(status.CurrentCampaign.TimeBasedDrops) > 0 {
				for _, drop := range status.CurrentCampaign.TimeBasedDrops {
					if !drop.Self.IsClaimed {
						// Create an ActiveDrop from the TimeBasedDrop
						activeDrop := drops.ActiveDrop{
							ID:              drop.ID,
							Name:            drop.Name,
							GameName:        status.CurrentCampaign.Game.Name,
							RequiredMinutes: drop.RequiredMinutesWatched,
							CurrentMinutes:  drop.Self.CurrentMinutesWatched,
							Progress:        float64(drop.Self.CurrentMinutesWatched) / float64(drop.RequiredMinutesWatched),
							IsClaimed:       drop.Self.IsClaimed,
						}
						currentDrop = &activeDrop
						break
					}
				}
			}
		}
	}

	response := gin.H{
		"is_running":       true,
		"current_campaign": status.CurrentCampaign,
		"current_stream":   status.CurrentStream,
		"current_drop":     currentDrop,
		"last_update":      status.LastUpdate,
		"next_switch":      status.NextSwitch,
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) getDropProgress(c *gin.Context) {
	logrus.Infof("=== getDropProgress HANDLER CALLED ===")
	status := s.miner.GetStatus()

	if !status.IsRunning {
		c.JSON(http.StatusOK, gin.H{
			"is_running":   false,
			"active_drops": []drops.ActiveDrop{},
			"total_progress": gin.H{
				"claimed_drops":         0,
				"total_drops":           0,
				"completion_percentage": 0,
			},
		})
		return
	}

	// Use TDM's DropCurrentSessionContext to get real progress
	var activeDrops []drops.ActiveDrop
	logrus.Infof("=== Progress Handler Debug ===")
	logrus.Infof("CurrentCampaign is nil: %v", status.CurrentCampaign == nil)
	logrus.Infof("CurrentStream is nil: %v", status.CurrentStream == nil)
	if status.CurrentCampaign != nil && status.CurrentStream != nil {
		logrus.Infof("=== Using DropCurrentSessionContext for Real Progress ===")
		logrus.Infof("Channel ID (UserID): %s, Stream ID: %s", status.CurrentStream.UserID, status.CurrentStream.ID)

		// Use DropCurrentSessionContext with correct parameters - use UserID as channelID
		logrus.Infof("About to call GetCurrentDropProgress with channelID: %s", status.CurrentStream.UserID)
		currentDropInfo, err := s.twitchClient.GetCurrentDropProgress(c.Request.Context(), status.CurrentStream.UserID)
		if err != nil {
			logrus.Errorf("Failed to get DropCurrentSessionContext progress: %v", err)
		} else {
			logrus.Infof("GetCurrentDropProgress completed successfully - got real progress!")
		}

		// Use ONLY DropCurrentSessionContext for real progress, infer other drops
		sortedDrops := make([]twitch.TimeBased, len(status.CurrentCampaign.TimeBasedDrops))
		copy(sortedDrops, status.CurrentCampaign.TimeBasedDrops)

		// Sort drops by required minutes (30, 90, 180)
		for i := 0; i < len(sortedDrops)-1; i++ {
			for j := i + 1; j < len(sortedDrops); j++ {
				if sortedDrops[i].RequiredMinutesWatched > sortedDrops[j].RequiredMinutesWatched {
					sortedDrops[i], sortedDrops[j] = sortedDrops[j], sortedDrops[i]
				}
			}
		}

		logrus.Infof("🔄 Processing drops in order by required minutes...")

		for i, drop := range sortedDrops {
			currentMinutes := 0
			isClaimed := false

			if currentDropInfo != nil && currentDropInfo.DropID == drop.ID {
				// This is the currently active drop - use real progress
				currentMinutes = currentDropInfo.CurrentMinutesWatched
				isClaimed = currentMinutes >= drop.RequiredMinutesWatched
				logrus.Infof("🎯 ACTIVE drop '%s': %d/%d minutes (real-time)", drop.Name, currentMinutes, drop.RequiredMinutesWatched)
			} else {
				// This is not the active drop - infer status
				if currentDropInfo != nil {
					// Find which drop is currently active
					for j, checkDrop := range sortedDrops {
						if checkDrop.ID == currentDropInfo.DropID {
							if j > i {
								// Active drop is after this one, so this one must be completed
								currentMinutes = drop.RequiredMinutesWatched
								isClaimed = true
								logrus.Infof("✅ COMPLETED drop '%s': %d/%d minutes (inferred)", drop.Name, currentMinutes, drop.RequiredMinutesWatched)
							} else {
								// Active drop is this one or before, so this one is not started
								currentMinutes = 0
								isClaimed = false
								logrus.Infof("⏳ NOT STARTED drop '%s': %d/%d minutes (inferred)", drop.Name, currentMinutes, drop.RequiredMinutesWatched)
							}
							break
						}
					}
				}
			}

			activeDrop := drops.ActiveDrop{
				ID:              drop.ID,
				Name:            drop.Name,
				GameName:        status.CurrentCampaign.Game.Name,
				RequiredMinutes: drop.RequiredMinutesWatched,
				CurrentMinutes:  currentMinutes,
				Progress:        float64(currentMinutes) / float64(drop.RequiredMinutesWatched),
				IsClaimed:       isClaimed,
			}
			activeDrops = append(activeDrops, activeDrop)
		}
	}

	// Calculate total progress statistics
	totalDrops := len(activeDrops)
	claimedDrops := 0
	for _, drop := range activeDrops {
		if drop.IsClaimed {
			claimedDrops++
		}
	}

	completionPercentage := 0.0
	if totalDrops > 0 {
		completionPercentage = (float64(claimedDrops) / float64(totalDrops)) * 100
	}

	response := gin.H{
		"is_running":   true,
		"active_drops": activeDrops,
		"total_progress": gin.H{
			"claimed_drops":         claimedDrops,
			"total_drops":           totalDrops,
			"completion_percentage": completionPercentage,
		},
		"last_update": status.LastUpdate,
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) startMiner(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	if s.miner.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Miner is already running"})
		return
	}

	// Start miner in background with a fresh context
	go func() {
		ctx := context.Background()
		if err := s.miner.Start(ctx); err != nil {
			logrus.Errorf("Miner start error: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) stopMiner(c *gin.Context) {
	if !s.miner.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Miner is not running"})
		return
	}

	if err := s.miner.Stop(); err != nil {
		logrus.Errorf("Failed to stop miner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop miner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Settings handlers
func (s *Server) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, s.config)
}

func (s *Server) updateSettings(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update configuration
	if priorityGames, ok := updates["priority_games"].([]interface{}); ok {
		var games []config.GameConfig
		for _, game := range priorityGames {
			if gameMap, ok := game.(map[string]interface{}); ok {
				gameConfig := config.GameConfig{
					Name: getString(gameMap, "name"),
					Slug: getString(gameMap, "slug"),
					ID:   getString(gameMap, "id"),
				}
				games = append(games, gameConfig)
			} else if gameStr, ok := game.(string); ok {
				// Handle legacy string format - convert to GameConfig
				gameConfig := config.GameConfig{
					Name: gameStr,
					Slug: "", // Will be populated when used
					ID:   "", // Will be populated when used
				}
				games = append(games, gameConfig)
			}
		}
		s.config.PriorityGames = games
	}

	if excludeGames, ok := updates["exclude_games"].([]interface{}); ok {
		var games []config.GameConfig
		for _, game := range excludeGames {
			if gameMap, ok := game.(map[string]interface{}); ok {
				gameConfig := config.GameConfig{
					Name: getString(gameMap, "name"),
					Slug: getString(gameMap, "slug"),
					ID:   getString(gameMap, "id"),
				}
				games = append(games, gameConfig)
			} else if gameStr, ok := game.(string); ok {
				// Handle legacy string format - convert to GameConfig
				gameConfig := config.GameConfig{
					Name: gameStr,
					Slug: "", // Will be populated when used
					ID:   "", // Will be populated when used
				}
				games = append(games, gameConfig)
			}
		}
		s.config.ExcludeGames = games
	}

	if watchUnlisted, ok := updates["watch_unlisted"].(bool); ok {
		s.config.WatchUnlisted = watchUnlisted
	}

	if claimDrops, ok := updates["claim_drops"].(bool); ok {
		s.config.ClaimDrops = claimDrops
	}

	if webhookURL, ok := updates["webhook_url"].(string); ok {
		s.config.WebhookURL = webhookURL
	}

	if checkInterval, ok := updates["check_interval"].(float64); ok {
		s.config.CheckInterval = int(checkInterval)
	}

	if switchThreshold, ok := updates["switch_threshold"].(float64); ok {
		s.config.SwitchThreshold = int(switchThreshold)
	}

	if minimumPoints, ok := updates["minimum_points"].(float64); ok {
		s.config.MinimumPoints = int(minimumPoints)
	}

	if maximumStreams, ok := updates["maximum_streams"].(float64); ok {
		s.config.MaximumStreams = int(maximumStreams)
	}

	if theme, ok := updates["theme"].(string); ok {
		s.config.Theme = theme
	}

	if language, ok := updates["language"].(string); ok {
		s.config.Language = language
	}

	if showTray, ok := updates["show_tray"].(bool); ok {
		s.config.ShowTray = showTray
	}

	if startMinimized, ok := updates["start_minimized"].(bool); ok {
		s.config.StartMinimized = startMinimized
	}

	// Update miner configuration
	minerConfig := &drops.MinerConfig{
		CheckInterval:   time.Duration(s.config.CheckInterval) * time.Second,
		SwitchThreshold: time.Duration(s.config.SwitchThreshold) * time.Minute,
		MinimumPoints:   s.config.MinimumPoints,
		MaximumStreams:  s.config.MaximumStreams,
		PriorityGames:   s.config.PriorityGames,
		ExcludeGames:    s.config.ExcludeGames,
		WatchUnlisted:   s.config.WatchUnlisted,
		ClaimDrops:      s.config.ClaimDrops,
		WebhookURL:      s.config.WebhookURL,
	}
	s.miner.SetConfig(minerConfig)

	// Save configuration
	if err := s.config.Save(); err != nil {
		logrus.Errorf("Failed to save configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Game management handlers
func (s *Server) addGameWithSlug(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	var req struct {
		GameName   string `json:"game_name" binding:"required"`
		ToPriority bool   `json:"to_priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get the slug and ID from Twitch
	slugInfo, err := s.twitchClient.GetGameSlug(c.Request.Context(), req.GameName)
	if err != nil {
		logrus.Errorf("Failed to get slug for game '%s': %v", req.GameName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve game slug"})
		return
	}

	// Add the game to config with the resolved slug and ID
	logrus.Infof("Adding game '%s' with slug '%s' and ID '%s' to config (priority: %v)", req.GameName, slugInfo.Slug, slugInfo.ID, req.ToPriority)
	err = s.config.AddGameToConfig(req.GameName, slugInfo.Slug, slugInfo.ID, req.ToPriority)
	if err != nil {
		logrus.Errorf("Failed to add game to config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add game to config"})
		return
	}

	logrus.Infof("Successfully added game '%s' with slug '%s' and ID '%s' to config", req.GameName, slugInfo.Slug, slugInfo.ID)

	// Update miner configuration with the new game list
	minerConfig := &drops.MinerConfig{
		CheckInterval:   time.Duration(s.config.CheckInterval) * time.Second,
		SwitchThreshold: time.Duration(s.config.SwitchThreshold) * time.Minute,
		MinimumPoints:   s.config.MinimumPoints,
		MaximumStreams:  s.config.MaximumStreams,
		PriorityGames:   s.config.PriorityGames,
		ExcludeGames:    s.config.ExcludeGames,
		WatchUnlisted:   s.config.WatchUnlisted,
		ClaimDrops:      s.config.ClaimDrops,
		WebhookURL:      s.config.WebhookURL,
	}
	s.miner.SetConfig(minerConfig)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"game": config.GameConfig{
			Name: req.GameName,
			Slug: slugInfo.Slug,
			ID:   slugInfo.ID,
		},
	})
}

// Stream handlers
func (s *Server) getStreamsForGame(c *gin.Context) {
	if !s.twitchClient.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	streams, err := s.twitchClient.GetStreamsForGameName(c.Request.Context(), gameID, limit)
	if err != nil {
		logrus.Errorf("Failed to get streams for game: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get streams"})
		return
	}

	c.JSON(http.StatusOK, streams)
}

func (s *Server) getCurrentStream(c *gin.Context) {
	status := s.miner.GetStatus()
	if status.CurrentStream == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No current stream"})
		return
	}

	c.JSON(http.StatusOK, status.CurrentStream)
}

// Device code storage methods (in production, use Redis or database)
func (s *Server) storeDeviceCode(deviceCode string, response *twitch.DeviceCodeResponse) {
	s.deviceCodes[deviceCode] = response

	// Clean up expired codes after their expiry time
	go func() {
		time.Sleep(time.Duration(response.ExpiresIn) * time.Second)
		delete(s.deviceCodes, deviceCode)
	}()
}

func (s *Server) getDeviceCode(deviceCode string) *twitch.DeviceCodeResponse {
	return s.deviceCodes[deviceCode]
}

// Helper function for safe string extraction
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
