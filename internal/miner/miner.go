package miner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"twitchdropsfarmer/internal/storage"
	"twitchdropsfarmer/internal/twitch"
)

// Miner handles the drop farming logic
type Miner struct {
	storage      *storage.Storage
	twitchClient *twitch.Client
	
	// State management
	isRunning       bool
	currentGame     *twitch.Game
	currentStream   *twitch.Stream
	watchStartTime  time.Time
	
	// Control channels
	ctx        context.Context
	cancel     context.CancelFunc
	
	// Logs
	logs    []LogEntry
	logsMux sync.RWMutex
	
	// Status
	statusMux sync.RWMutex
	status    MinerStatus
	
	// WebSocket clients for real-time updates
	wsClients map[chan []byte]bool
	wsMux     sync.RWMutex
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	GameID    string    `json:"gameId,omitempty"`
	StreamID  string    `json:"streamId,omitempty"`
}

type MinerStatus struct {
	IsRunning     bool               `json:"isRunning"`
	CurrentGame   *twitch.Game       `json:"currentGame"`
	CurrentStream *twitch.Stream     `json:"currentStream"`
	WatchDuration time.Duration      `json:"watchDuration"`
	TotalWatched  time.Duration      `json:"totalWatched"`
	GamesQueue    []twitch.Game      `json:"gamesQueue"`
	LastUpdate    time.Time          `json:"lastUpdate"`
}

// New creates a new miner instance
func New(storage *storage.Storage, twitchClient *twitch.Client) *Miner {
	return &Miner{
		storage:      storage,
		twitchClient: twitchClient,
		logs:         make([]LogEntry, 0),
		wsClients:    make(map[chan []byte]bool),
		status: MinerStatus{
			IsRunning:    false,
			GamesQueue:   make([]twitch.Game, 0),
			LastUpdate:   time.Now(),
		},
	}
}

// Start begins the mining process
func (m *Miner) Start() {
	m.statusMux.Lock()
	if m.isRunning {
		m.statusMux.Unlock()
		return
	}
	
	m.isRunning = true
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.status.IsRunning = true
	m.status.LastUpdate = time.Now()
	m.statusMux.Unlock()
	
	m.log("INFO", "Miner started", "", "")
	
	go m.runMiningLoop()
}

// Stop stops the mining process
func (m *Miner) Stop() {
	m.statusMux.Lock()
	if !m.isRunning {
		m.statusMux.Unlock()
		return
	}
	
	m.isRunning = false
	if m.cancel != nil {
		m.cancel()
	}
	m.status.IsRunning = false
	m.status.CurrentGame = nil
	m.status.CurrentStream = nil
	m.status.LastUpdate = time.Now()
	m.statusMux.Unlock()
	
	m.log("INFO", "Miner stopped", "", "")
}

// runMiningLoop is the main mining loop
func (m *Miner) runMiningLoop() {
	ticker := time.NewTicker(20 * time.Second) // Watch interval like TwitchDropsMiner
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.processNextGame()
		}
	}
}

// processNextGame processes the next game in the queue
func (m *Miner) processNextGame() {
	settings := m.storage.GetSettings()
	if len(settings.Games) == 0 {
		m.log("DEBUG", "No games in watch list", "", "")
		return
	}
	
	// Get the next game to watch (priority order)
	var nextGame *twitch.Game
	for _, gameID := range settings.Games {
		if game := m.storage.GetGame(gameID); game != nil {
			nextGame = game
			break
		}
	}
	
	if nextGame == nil {
		m.log("WARNING", "No valid games found in watch list", "", "")
		return
	}
	
	// Check if we need to switch games
	if m.currentGame == nil || m.currentGame.ID != nextGame.ID {
		m.switchToGame(nextGame)
	}
	
	// Continue watching current stream or find a new one
	if m.currentStream == nil || !m.isStreamLive(m.currentStream) {
		m.findAndWatchStream(nextGame)
	} else {
		m.continueWatching()
	}
}

// switchToGame switches to watching a different game
func (m *Miner) switchToGame(game *twitch.Game) {
	m.statusMux.Lock()
	m.currentGame = game
	m.currentStream = nil
	m.status.CurrentGame = game
	m.status.CurrentStream = nil
	m.status.LastUpdate = time.Now()
	m.statusMux.Unlock()
	
	m.log("INFO", fmt.Sprintf("Switching to game: %s", game.DisplayName), game.ID, "")
}

// findAndWatchStream finds a stream for the current game and starts watching
func (m *Miner) findAndWatchStream(game *twitch.Game) {
	streams, err := m.twitchClient.GetStreamsForGame(game.ID, 20)
	if err != nil {
		m.log("ERROR", fmt.Sprintf("Failed to get streams for game %s: %v", game.DisplayName, err), game.ID, "")
		return
	}
	
	if len(streams) == 0 {
		m.log("WARNING", fmt.Sprintf("No live streams found for game: %s", game.DisplayName), game.ID, "")
		return
	}
	
	// Select the first available stream
	selectedStream := &streams[0]
	
	m.statusMux.Lock()
	m.currentStream = selectedStream
	m.watchStartTime = time.Now()
	m.status.CurrentStream = selectedStream
	m.status.LastUpdate = time.Now()
	m.statusMux.Unlock()
	
	m.log("INFO", fmt.Sprintf("Started watching stream: %s (%s)", selectedStream.UserName, selectedStream.Title), game.ID, selectedStream.ID)
	
	// Start watching
	m.watchStream(selectedStream)
}

// continueWatching continues watching the current stream
func (m *Miner) continueWatching() {
	if m.currentStream == nil {
		return
	}
	
	m.watchStream(m.currentStream)
}

// watchStream makes the GraphQL request to simulate watching
func (m *Miner) watchStream(stream *twitch.Stream) {
	err := m.twitchClient.WatchStream(stream.UserLogin)
	if err != nil {
		m.log("ERROR", fmt.Sprintf("Failed to watch stream %s: %v", stream.UserName, err), m.currentGame.ID, stream.ID)
		
		// Mark stream as offline and find a new one
		m.statusMux.Lock()
		m.currentStream = nil
		m.status.CurrentStream = nil
		m.status.LastUpdate = time.Now()
		m.statusMux.Unlock()
		
		return
	}
	
	// Update watch duration
	m.statusMux.Lock()
	if !m.watchStartTime.IsZero() {
		m.status.WatchDuration = time.Since(m.watchStartTime)
		m.status.TotalWatched += 20 * time.Second // Add the watch interval
	}
	m.status.LastUpdate = time.Now()
	m.statusMux.Unlock()
	
	m.log("SUCCESS", fmt.Sprintf("Watching %s - %s", stream.UserName, stream.Title), m.currentGame.ID, stream.ID)
	
	// Broadcast status update to WebSocket clients
	m.broadcastStatusUpdate()
}

// isStreamLive checks if a stream is still live
func (m *Miner) isStreamLive(stream *twitch.Stream) bool {
	// This would typically make a GraphQL request to check stream status
	// For now, assume streams stay live for the duration of our watch interval
	return true
}

// log adds a log entry and broadcasts it
func (m *Miner) log(level, message, gameID, streamID string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		GameID:    gameID,
		StreamID:  streamID,
	}
	
	m.logsMux.Lock()
	m.logs = append(m.logs, entry)
	
	// Keep only the last 1000 log entries
	if len(m.logs) > 1000 {
		m.logs = m.logs[len(m.logs)-1000:]
	}
	m.logsMux.Unlock()
	
	// Also log to console
	log.Printf("[%s] %s", level, message)
	
	// Broadcast log update to WebSocket clients
	m.broadcastLogUpdate(entry)
}

// GetStatus returns the current miner status
func (m *Miner) GetStatus() MinerStatus {
	m.statusMux.RLock()
	defer m.statusMux.RUnlock()
	
	status := m.status
	
	// Update current watch duration if watching
	if m.isRunning && !m.watchStartTime.IsZero() {
		status.WatchDuration = time.Since(m.watchStartTime)
	}
	
	return status
}

// GetLogs returns recent log entries
func (m *Miner) GetLogs(limit int) []LogEntry {
	m.logsMux.RLock()
	defer m.logsMux.RUnlock()
	
	if limit <= 0 || limit > len(m.logs) {
		limit = len(m.logs)
	}
	
	start := len(m.logs) - limit
	if start < 0 {
		start = 0
	}
	
	logs := make([]LogEntry, limit)
	copy(logs, m.logs[start:])
	
	return logs
}

// AddWebSocketClient adds a WebSocket client for real-time updates
func (m *Miner) AddWebSocketClient(client chan []byte) {
	m.wsMux.Lock()
	m.wsClients[client] = true
	m.wsMux.Unlock()
}

// RemoveWebSocketClient removes a WebSocket client
func (m *Miner) RemoveWebSocketClient(client chan []byte) {
	m.wsMux.Lock()
	delete(m.wsClients, client)
	close(client)
	m.wsMux.Unlock()
}

// broadcastStatusUpdate broadcasts status updates to all WebSocket clients
func (m *Miner) broadcastStatusUpdate() {
	status := m.GetStatus()
	
	message := map[string]interface{}{
		"type": "status",
		"data": status,
	}
	
	m.broadcast(message)
}

// broadcastLogUpdate broadcasts log updates to all WebSocket clients
func (m *Miner) broadcastLogUpdate(entry LogEntry) {
	message := map[string]interface{}{
		"type": "log",
		"data": entry,
	}
	
	m.broadcast(message)
}

// broadcast sends a message to all connected WebSocket clients
func (m *Miner) broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}
	
	m.wsMux.RLock()
	for client := range m.wsClients {
		select {
		case client <- data:
		default:
			// Client is not receiving, remove it
			delete(m.wsClients, client)
			close(client)
		}
	}
	m.wsMux.RUnlock()
}