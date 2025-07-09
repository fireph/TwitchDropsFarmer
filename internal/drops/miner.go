package drops

import (
	"context"
	"fmt"
	"sync"
	"time"

	"twitchdropsfarmer/internal/config"
	"twitchdropsfarmer/internal/storage"
	"twitchdropsfarmer/internal/twitch"

	"github.com/sirupsen/logrus"
)

type Miner struct {
	twitchClient *twitch.Client
	db           *storage.Database

	// Mining state
	mu              sync.RWMutex
	isRunning       bool
	currentCampaign *twitch.Campaign
	currentStream   *twitch.Stream
	currentSession  *storage.MiningSession
	watchingSession *twitch.WatchingSession

	// Status tracking
	status   *MinerStatus
	statusMu sync.RWMutex

	// Configuration
	config *MinerConfig

	// Channels for coordination
	stopChan   chan struct{}
	statusChan chan *MinerStatus
	configChan chan struct{}
}

type MinerConfig struct {
	CheckInterval   time.Duration
	WatchInterval   time.Duration // How often to send watch requests (like TDM ~20s)
	SwitchThreshold time.Duration
	MinimumPoints   int
	MaximumStreams  int
	PriorityGames   []config.GameConfig
	ExcludeGames    []config.GameConfig
	WatchUnlisted   bool
	ClaimDrops      bool
	WebhookURL      string
}

type MinerStatus struct {
	IsRunning       bool             `json:"is_running"`
	CurrentStream   *twitch.Stream   `json:"current_stream"`
	CurrentCampaign *twitch.Campaign `json:"current_campaign"`
	CurrentProgress int              `json:"current_progress"`
	TotalCampaigns  int              `json:"total_campaigns"`
	ClaimedDrops    int              `json:"claimed_drops"`
	LastUpdate      time.Time        `json:"last_update"`
	NextSwitch      time.Time        `json:"next_switch"`
	ErrorMessage    string           `json:"error_message"`
	ActiveDrops     []ActiveDrop     `json:"active_drops"`
}

type ActiveDrop struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	GameName        string    `json:"game_name"`
	RequiredMinutes int       `json:"required_minutes"`
	CurrentMinutes  int       `json:"current_minutes"`
	Progress        float64   `json:"progress"`
	IsClaimed       bool      `json:"is_claimed"`
	EstimatedTime   time.Time `json:"estimated_time"`
}

func NewMiner(twitchClient *twitch.Client, db *storage.Database) *Miner {
	return &Miner{
		twitchClient: twitchClient,
		db:           db,
		config: &MinerConfig{
			CheckInterval:   60 * time.Second,
			WatchInterval:   20 * time.Second, // Like TDM - every ~20 seconds
			SwitchThreshold: 5 * time.Minute,
			MinimumPoints:   50,
			MaximumStreams:  3,
			PriorityGames:   []config.GameConfig{},
			ExcludeGames:    []config.GameConfig{},
			WatchUnlisted:   true,
			ClaimDrops:      true,
		},
		status: &MinerStatus{
			IsRunning:   false,
			LastUpdate:  time.Now(),
			ActiveDrops: []ActiveDrop{},
		},
		stopChan:   make(chan struct{}),
		statusChan: make(chan *MinerStatus, 100),
		configChan: make(chan struct{}, 1), // Buffered channel to avoid blocking
	}
}

func (m *Miner) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return fmt.Errorf("miner is already running")
	}
	m.isRunning = true
	m.mu.Unlock()

	logrus.Info("Starting drop miner...")

	// Update status
	m.updateStatus(func(s *MinerStatus) {
		s.IsRunning = true
		s.LastUpdate = time.Now()
		s.ErrorMessage = ""
	})

	// Start mining loop (campaign selection, stream switching)
	checkTicker := time.NewTicker(m.config.CheckInterval)
	defer checkTicker.Stop()

	// Start watch loop (periodic HEAD requests to maintain viewing)
	watchTicker := time.NewTicker(m.config.WatchInterval)
	defer watchTicker.Stop()

	// Initial check
	if err := m.checkAndUpdate(ctx); err != nil {
		logrus.Errorf("Initial check failed: %v", err)
		m.updateStatus(func(s *MinerStatus) {
			s.ErrorMessage = fmt.Sprintf("Initial check failed: %v", err)
		})
	}

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Drop miner context cancelled")
			return m.stop()
		case <-m.stopChan:
			logrus.Info("Drop miner stop requested")
			return m.stop()
		case <-checkTicker.C:
			if err := m.checkAndUpdate(ctx); err != nil {
				logrus.Errorf("Mining check failed: %v", err)
				m.updateStatus(func(s *MinerStatus) {
					s.ErrorMessage = fmt.Sprintf("Mining check failed: %v", err)
				})
			}
		case <-m.configChan:
			// Configuration changed, trigger immediate re-evaluation
			logrus.Info("Configuration updated, re-evaluating campaigns...")
			if err := m.checkAndUpdate(ctx); err != nil {
				logrus.Errorf("Config-triggered mining check failed: %v", err)
				m.updateStatus(func(s *MinerStatus) {
					s.ErrorMessage = fmt.Sprintf("Config-triggered mining check failed: %v", err)
				})
			}
		case <-watchTicker.C:
			// Send periodic watch request to maintain viewing (like TDM)
			if err := m.sendWatchRequest(ctx); err != nil {
				logrus.Debugf("Watch request failed: %v", err)
			}
		}
	}
}

func (m *Miner) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return fmt.Errorf("miner is not running")
	}

	close(m.stopChan)
	return nil
}

func (m *Miner) stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isRunning = false

	// End current session if active
	if m.currentSession != nil {
		minutesWatched := int(time.Since(m.currentSession.StartedAt).Minutes())
		if err := m.db.EndMiningSession(m.currentSession.ID, minutesWatched); err != nil {
			logrus.Errorf("Failed to end mining session: %v", err)
		}
		m.currentSession = nil
	}

	// Clear watching session
	m.watchingSession = nil

	// Update status
	m.updateStatus(func(s *MinerStatus) {
		s.IsRunning = false
		s.LastUpdate = time.Now()
		s.CurrentStream = nil
		s.CurrentCampaign = nil
	})

	logrus.Info("Drop miner stopped")
	return nil
}

func (m *Miner) checkAndUpdate(ctx context.Context) error {
	// Check if user is logged in
	if !m.twitchClient.IsLoggedIn() {
		return fmt.Errorf("user is not logged in")
	}

	// Debug: Check user info first
	user := m.twitchClient.GetUser()
	if user != nil {
		logrus.Debugf("Authenticated user: %s (ID: %s)", user.DisplayName, user.ID)
	} else {
		logrus.Debug("No user info available")
	}

	// Fetch active campaigns
	campaigns, err := m.twitchClient.GetDropCampaigns(ctx)
	if err != nil {
		logrus.Debugf("Failed to fetch campaigns (this is expected during development): %v", err)
		// Return nil to prevent error spam during development
		return nil
	}

	if len(campaigns) == 0 {
		logrus.Info("No active campaigns found")
		return nil
	}

	// Find best campaign to watch
	bestCampaign := m.selectBestCampaign(campaigns)
	if bestCampaign == nil {
		logrus.Info("No suitable campaign found")
		return nil
	}

	// Always try to get detailed campaign information for priority games since ViewerDropsDashboard doesn't include timeBasedDrops
	if m.isGamePriority(bestCampaign.Game.Name) {
		logrus.Debugf("Fetching detailed campaign info for priority game '%s' campaign '%s'", bestCampaign.Game.Name, bestCampaign.Name)
		logrus.Debugf("Basic campaign has %d timeBasedDrops and %d eventBasedDrops", len(bestCampaign.TimeBasedDrops), len(bestCampaign.EventBasedDrops))

		detailedCampaign, err := m.twitchClient.GetCampaignDetails(ctx, bestCampaign.ID)
		if err != nil {
			logrus.Errorf("Failed to get detailed campaign info: %v", err)
		} else {
			logrus.Infof("SUCCESS: Detailed campaign '%s' has %d timeBasedDrops and %d eventBasedDrops",
				detailedCampaign.Name, len(detailedCampaign.TimeBasedDrops), len(detailedCampaign.EventBasedDrops))

			// Log details about the drops
			for i, drop := range detailedCampaign.TimeBasedDrops {
				logrus.Infof("Drop %d: %s - %d minutes required, %d minutes watched, claimed: %v",
					i+1, drop.Name, drop.RequiredMinutesWatched, drop.Self.CurrentMinutesWatched, drop.Self.IsClaimed)
			}

			// Always replace with detailed campaign since it has proper drop data
			bestCampaign = detailedCampaign
			logrus.Infof("✅ UPDATED campaign with detailed drop information: %d timeBasedDrops, %d eventBasedDrops",
				len(bestCampaign.TimeBasedDrops), len(bestCampaign.EventBasedDrops))
		}
	}

	// Check if we need to switch campaigns
	if m.shouldSwitchCampaign(bestCampaign) {
		if err := m.switchToCampaign(ctx, bestCampaign); err != nil {
			return fmt.Errorf("failed to switch campaign: %w", err)
		}
	}

	// Update progress for current drops
	if err := m.updateDropProgress(ctx); err != nil {
		logrus.Errorf("Failed to update drop progress: %v", err)
	}

	// Check for completed drops to claim
	if m.config.ClaimDrops {
		if err := m.checkAndClaimDrops(ctx); err != nil {
			logrus.Errorf("Failed to check and claim drops: %v", err)
		}
	}

	// Update status
	m.updateMinerStatus(campaigns)
	return nil
}

func (m *Miner) selectBestCampaign(campaigns []twitch.Campaign) *twitch.Campaign {
	var bestCampaign *twitch.Campaign
	var bestScore int

	logrus.Debugf("Selecting from %d campaigns", len(campaigns))
	logrus.Debugf("Priority games configured: %v", m.config.PriorityGames)
	logrus.Debugf("Watch unlisted: %v", m.config.WatchUnlisted)

	// Debug: Count campaigns by game to see what's available
	gameCount := make(map[string]int)
	dropCount := make(map[string]int)
	for _, campaign := range campaigns {
		gameCount[campaign.Game.Name]++
		dropCount[campaign.Game.Name] += len(campaign.TimeBasedDrops)
	}
	logrus.Debugf("Available campaigns by game: %+v", gameCount)
	logrus.Debugf("Available drops by game: %+v", dropCount)

	for _, campaign := range campaigns {
		logrus.Debugf("Evaluating campaign: %s (Game: %s, Status: %s, Connected: %v)",
			campaign.Name, campaign.Game.Name, campaign.Status, campaign.Self.IsAccountConnected)

		// Skip expired campaigns first
		if campaign.Status != "ACTIVE" {
			logrus.Debugf("Skipping %s - campaign status is %s (not ACTIVE)", campaign.Game.Name, campaign.Status)
			continue
		}

		// Skip if not account connected and game is not in priority list
		if !campaign.Self.IsAccountConnected && !m.isGamePriority(campaign.Game.Name) {
			logrus.Debugf("Skipping %s - not connected and not priority", campaign.Game.Name)
			continue
		}

		// Skip if game is excluded
		if m.isGameExcluded(campaign.Game.Name) {
			logrus.Debugf("Skipping %s - game is excluded", campaign.Game.Name)
			continue
		}

		// Calculate score
		score := m.calculateCampaignScore(&campaign)
		logrus.Debugf("Campaign %s score: %d", campaign.Game.Name, score)
		if score > bestScore {
			logrus.Debugf("New best campaign: %s (score %d beats previous %d)", campaign.Game.Name, score, bestScore)
			bestScore = score
			// Create a copy to avoid Go loop variable reference issue
			campaignCopy := campaign
			bestCampaign = &campaignCopy
		}
	}

	if bestCampaign != nil {
		logrus.Infof("Selected campaign: %s (Game: %s, Score: %d)",
			bestCampaign.Name, bestCampaign.Game.Name, bestScore)
	} else {
		logrus.Info("No suitable campaign found")
	}

	return bestCampaign
}

func (m *Miner) calculateCampaignScore(campaign *twitch.Campaign) int {
	score := 0

	// Priority games get higher score
	isPriority := m.isGamePriority(campaign.Game.Name)
	logrus.Debugf("Game '%s' priority check: %v (priority games: %v)", campaign.Game.Name, isPriority, m.config.PriorityGames)
	if isPriority {
		score += 100
		logrus.Debugf("Added 100 points for priority game '%s'", campaign.Game.Name)
	} else if m.config.WatchUnlisted {
		// Non-priority games get base score if watching unlisted is enabled
		score += 10
		logrus.Debugf("Added 10 points for unlisted game '%s' (watch_unlisted=true)", campaign.Game.Name)
	} else {
		logrus.Debugf("No points for game '%s' (not priority and watch_unlisted=false)", campaign.Game.Name)
	}

	// Account connected campaigns get bonus
	if campaign.Self.IsAccountConnected {
		score += 50
	}

	// Active campaigns get bonus
	if campaign.Status == "ACTIVE" {
		score += 25
	}

	// Active time-based drops get score
	for _, drop := range campaign.TimeBasedDrops {
		if !drop.Self.IsClaimed {
			score += 10
			// Partially completed drops get extra score
			if drop.Self.CurrentMinutesWatched > 0 {
				score += 5
			}
		}
	}

	// Event-based drops get score
	score += len(campaign.EventBasedDrops) * 5

	return score
}

func (m *Miner) shouldSwitchCampaign(newCampaign *twitch.Campaign) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Switch if no current campaign
	if m.currentCampaign == nil {
		return true
	}

	// Switch if campaign changed
	if m.currentCampaign.ID != newCampaign.ID {
		return true
	}

	// Switch if current stream is offline or has issues
	if m.currentStream == nil {
		return true
	}

	// Switch if we've been watching for the threshold time
	if m.currentSession != nil &&
		time.Since(m.currentSession.StartedAt) > m.config.SwitchThreshold {
		return true
	}

	return false
}

func (m *Miner) switchToCampaign(ctx context.Context, campaign *twitch.Campaign) error {
	logrus.Infof("Switching to campaign: %s", campaign.Name)

	// End current session
	if m.currentSession != nil {
		minutesWatched := int(time.Since(m.currentSession.StartedAt).Minutes())
		if err := m.db.EndMiningSession(m.currentSession.ID, minutesWatched); err != nil {
			logrus.Errorf("Failed to end current session: %v", err)
		}
	}

	// Find best stream for this campaign
	streams, err := m.twitchClient.GetStreamsForGameName(ctx, campaign.Game.Name, m.config.MaximumStreams)
	if err != nil {
		return fmt.Errorf("failed to get streams for game: %w", err)
	}

	if len(streams) == 0 {
		return fmt.Errorf("no streams found for game: %s", campaign.Game.Name)
	}

	// Select best stream
	bestStream := m.selectBestStream(streams)
	if bestStream == nil {
		return fmt.Errorf("no suitable stream found for game: %s", campaign.Game.Name)
	}

	// Start new session
	user := m.twitchClient.GetUser()
	if user == nil {
		return fmt.Errorf("user not available")
	}

	sessionID, err := m.db.StartMiningSession(user.ID, campaign.ID, bestStream.ID)
	if err != nil {
		return fmt.Errorf("failed to start mining session: %w", err)
	}

	// Start watching session like TDM
	watchingSession, err := m.twitchClient.StartWatching(ctx, bestStream.UserLogin)
	if err != nil {
		return fmt.Errorf("failed to start watching session: %w", err)
	}

	// Update current state
	m.mu.Lock()
	m.currentCampaign = campaign
	m.currentStream = bestStream
	m.currentSession = &storage.MiningSession{
		ID:         sessionID,
		UserID:     user.ID,
		CampaignID: campaign.ID,
		StreamID:   bestStream.ID,
		StartedAt:  time.Now(),
		Status:     "active",
	}
	m.watchingSession = watchingSession
	m.mu.Unlock()

	// Save to database
	if err := m.db.SetCurrentWatchingStream(bestStream.ID); err != nil {
		logrus.Errorf("Failed to save current watching stream: %v", err)
	}

	logrus.Infof("Now watching: %s playing %s", bestStream.UserName, bestStream.GameName)
	return nil
}

func (m *Miner) selectBestStream(streams []twitch.Stream) *twitch.Stream {
	if len(streams) == 0 {
		return nil
	}

	// For now, select the stream with highest viewer count
	// TODO: Add more sophisticated selection logic
	bestStream := &streams[0]
	for _, stream := range streams {
		if stream.ViewerCount > bestStream.ViewerCount {
			bestStream = &stream
		}
	}

	return bestStream
}

func (m *Miner) updateDropProgress(ctx context.Context) error {
	m.mu.RLock()
	campaign := m.currentCampaign
	session := m.currentSession
	m.mu.RUnlock()

	if campaign == nil || session == nil {
		return nil
	}

	// Calculate minutes watched in current session
	minutesWatched := int(time.Since(session.StartedAt).Minutes())

	// Update progress for time-based drops
	for _, drop := range campaign.TimeBasedDrops {
		if !drop.Self.IsClaimed {
			newProgress := drop.Self.CurrentMinutesWatched + minutesWatched
			if newProgress > drop.RequiredMinutesWatched {
				newProgress = drop.RequiredMinutesWatched
			}

			// Update in database
			if err := m.db.UpdateDropProgress(drop.ID, newProgress); err != nil {
				logrus.Errorf("Failed to update drop progress: %v", err)
			}
		}
	}

	return nil
}

func (m *Miner) checkAndClaimDrops(ctx context.Context) error {
	m.mu.RLock()
	campaign := m.currentCampaign
	m.mu.RUnlock()

	if campaign == nil {
		return nil
	}

	for _, drop := range campaign.TimeBasedDrops {
		if !drop.Self.IsClaimed &&
			drop.Self.CurrentMinutesWatched >= drop.RequiredMinutesWatched &&
			drop.Self.DropInstanceID != "" {

			logrus.Infof("Claiming drop: %s", drop.Name)
			if err := m.twitchClient.ClaimDrop(ctx, drop.Self.DropInstanceID); err != nil {
				logrus.Errorf("Failed to claim drop %s: %v", drop.Name, err)
				continue
			}

			// Mark as claimed in database
			if err := m.db.MarkDropClaimed(drop.ID); err != nil {
				logrus.Errorf("Failed to mark drop as claimed: %v", err)
			}

			logrus.Infof("Successfully claimed drop: %s", drop.Name)
		}
	}

	return nil
}

func (m *Miner) updateMinerStatus(campaigns []twitch.Campaign) {
	m.mu.RLock()
	currentCampaign := m.currentCampaign
	currentStream := m.currentStream
	currentSession := m.currentSession
	m.mu.RUnlock()

	// Calculate current session minutes for progress tracking
	var currentSessionMinutes int
	if currentCampaign != nil && currentSession != nil {
		currentSessionMinutes = int(time.Since(currentSession.StartedAt).Minutes())
	}

	// Calculate active drops
	var activeDrops []ActiveDrop
	var claimedDrops int

	for _, campaign := range campaigns {
		for _, drop := range campaign.TimeBasedDrops {
			if drop.Self.IsClaimed {
				claimedDrops++
			} else {
				// Add current session progress for the current campaign's drops
				var currentMinutes int = drop.Self.CurrentMinutesWatched
				if currentCampaign != nil && campaign.ID == currentCampaign.ID && currentSessionMinutes > 0 {
					currentMinutes += currentSessionMinutes
				}

				// Ensure we don't exceed the required minutes
				if currentMinutes > drop.RequiredMinutesWatched {
					currentMinutes = drop.RequiredMinutesWatched
				}

				progress := float64(currentMinutes) / float64(drop.RequiredMinutesWatched)
				if progress > 1.0 {
					progress = 1.0
				}

				remainingMinutes := drop.RequiredMinutesWatched - currentMinutes
				if remainingMinutes < 0 {
					remainingMinutes = 0
				}
				estimatedTime := time.Now().Add(time.Duration(remainingMinutes) * time.Minute)

				activeDrops = append(activeDrops, ActiveDrop{
					ID:              drop.ID,
					Name:            drop.Name,
					GameName:        campaign.Game.Name,
					RequiredMinutes: drop.RequiredMinutesWatched,
					CurrentMinutes:  currentMinutes, // Now includes current session progress
					Progress:        progress,
					IsClaimed:       drop.Self.IsClaimed,
					EstimatedTime:   estimatedTime,
				})
			}
		}
	}

	// Use the same current session minutes for consistency
	currentProgress := currentSessionMinutes

	// Calculate next switch time
	var nextSwitch time.Time
	if currentSession != nil {
		nextSwitch = currentSession.StartedAt.Add(m.config.SwitchThreshold)
	}

	// Debug: Log active drops information
	logrus.Debugf("Active drops summary: %d total active drops found", len(activeDrops))
	if currentCampaign != nil {
		campaignDrops := 0
		for _, drop := range activeDrops {
			if drop.GameName == currentCampaign.Game.Name {
				campaignDrops++
			}
		}
		logrus.Debugf("Current campaign '%s' (%s) has %d active drops", currentCampaign.Name, currentCampaign.Game.Name, campaignDrops)
	}

	m.updateStatus(func(s *MinerStatus) {
		s.CurrentStream = currentStream
		s.CurrentCampaign = currentCampaign
		s.CurrentProgress = currentProgress
		s.TotalCampaigns = len(campaigns)
		s.ClaimedDrops = claimedDrops
		s.LastUpdate = time.Now()
		s.NextSwitch = nextSwitch
		s.ActiveDrops = activeDrops
	})
}

func (m *Miner) updateStatus(updateFunc func(*MinerStatus)) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	updateFunc(m.status)

	// Send status update to channel (non-blocking)
	select {
	case m.statusChan <- m.status:
	default:
		// Channel is full, skip this update
	}
}

func (m *Miner) GetStatus() *MinerStatus {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()

	// Return a copy to prevent external modification
	statusCopy := *m.status
	return &statusCopy
}

func (m *Miner) GetStatusChannel() <-chan *MinerStatus {
	return m.statusChan
}

func (m *Miner) isGamePriority(gameName string) bool {
	for _, game := range m.config.PriorityGames {
		if game.Name == gameName {
			return true
		}
	}
	return false
}

func (m *Miner) isGameExcluded(gameName string) bool {
	for _, game := range m.config.ExcludeGames {
		if game.Name == gameName {
			return true
		}
	}
	return false
}

func (m *Miner) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

func (m *Miner) SetConfig(config *MinerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config

	// Trigger immediate re-evaluation if miner is running
	if m.isRunning {
		select {
		case m.configChan <- struct{}{}:
			// Successfully sent notification
		default:
			// Channel is already full, no need to send another notification
		}
	}
}

func (m *Miner) sendWatchRequest(ctx context.Context) error {
	m.mu.RLock()
	watchingSession := m.watchingSession
	m.mu.RUnlock()

	if watchingSession == nil {
		return nil // No active watching session
	}

	return m.twitchClient.SendWatchRequest(ctx, watchingSession)
}
