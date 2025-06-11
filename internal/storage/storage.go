package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"twitchdropsfarmer/internal/twitch"
)

// Storage handles persistent data storage
type Storage struct {
	dataDir string
	mutex   sync.RWMutex
	
	// In-memory cache
	settings *Settings
	games    map[string]*twitch.Game
	auth     *AuthData
}

// Settings represents app settings
type Settings struct {
	Games               []string          `json:"games"`               // Ordered list of game IDs
	WatchInterval       int               `json:"watchInterval"`       // Seconds between watch requests
	AutoClaimDrops      bool              `json:"autoClaimDrops"`      // Auto-claim completed drops
	NotificationsEnabled bool             `json:"notificationsEnabled"` // Enable notifications
	Theme               string            `json:"theme"`               // UI theme (light/dark)
	Language            string            `json:"language"`            // UI language
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// AuthData stores authentication information
type AuthData struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	UserID       string    `json:"userID"`
	Username     string    `json:"username"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// New creates a new storage instance
func New(dataDir string) *Storage {
	// Ensure data directory exists
	os.MkdirAll(dataDir, 0755)
	
	s := &Storage{
		dataDir: dataDir,
		games:   make(map[string]*twitch.Game),
	}
	
	// Load existing data
	s.loadSettings()
	s.loadGames()
	s.loadAuth()
	
	return s
}

// Settings management
func (s *Storage) GetSettings() *Settings {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if s.settings == nil {
		return &Settings{
			Games:               make([]string, 0),
			WatchInterval:       20,
			AutoClaimDrops:      true,
			NotificationsEnabled: true,
			Theme:               "dark",
			Language:            "en",
			UpdatedAt:           time.Now(),
		}
	}
	
	return s.settings
}

func (s *Storage) SaveSettings(settings *Settings) error {
	settings.UpdatedAt = time.Now()
	
	s.mutex.Lock()
	s.settings = settings
	s.mutex.Unlock()
	
	return s.saveSettings()
}

func (s *Storage) loadSettings() error {
	path := filepath.Join(s.dataDir, "settings.json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}
	
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}
	
	s.settings = &settings
	return nil
}

func (s *Storage) saveSettings() error {
	path := filepath.Join(s.dataDir, "settings.json")
	
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// Games management
func (s *Storage) GetGames() map[string]*twitch.Game {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	games := make(map[string]*twitch.Game)
	for id, game := range s.games {
		games[id] = game
	}
	
	return games
}

func (s *Storage) GetGame(id string) *twitch.Game {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if game, exists := s.games[id]; exists {
		return game
	}
	
	return nil
}

func (s *Storage) AddGame(game *twitch.Game) error {
	s.mutex.Lock()
	s.games[game.ID] = game
	s.mutex.Unlock()
	
	// Add to settings games list if not already there
	settings := s.GetSettings()
	for _, gameID := range settings.Games {
		if gameID == game.ID {
			return s.saveGames() // Already in list, just save games data
		}
	}
	
	// Add to games list
	settings.Games = append(settings.Games, game.ID)
	if err := s.SaveSettings(settings); err != nil {
		return err
	}
	
	return s.saveGames()
}

func (s *Storage) RemoveGame(id string) error {
	s.mutex.Lock()
	delete(s.games, id)
	s.mutex.Unlock()
	
	// Remove from settings games list
	settings := s.GetSettings()
	newGames := make([]string, 0)
	for _, gameID := range settings.Games {
		if gameID != id {
			newGames = append(newGames, gameID)
		}
	}
	settings.Games = newGames
	
	if err := s.SaveSettings(settings); err != nil {
		return err
	}
	
	return s.saveGames()
}

func (s *Storage) ReorderGames(gameIDs []string) error {
	settings := s.GetSettings()
	settings.Games = gameIDs
	return s.SaveSettings(settings)
}

func (s *Storage) loadGames() error {
	path := filepath.Join(s.dataDir, "games.json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}
	
	var games map[string]*twitch.Game
	if err := json.Unmarshal(data, &games); err != nil {
		return err
	}
	
	s.games = games
	return nil
}

func (s *Storage) saveGames() error {
	path := filepath.Join(s.dataDir, "games.json")
	
	data, err := json.MarshalIndent(s.games, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// Authentication management
func (s *Storage) GetAuth() *AuthData {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.auth
}

func (s *Storage) SaveAuth(auth *AuthData) error {
	auth.UpdatedAt = time.Now()
	
	fmt.Printf("Storage: Saving auth data to %s: %+v\n", s.dataDir, auth)
	
	s.mutex.Lock()
	s.auth = auth
	s.mutex.Unlock()
	
	err := s.saveAuth()
	if err != nil {
		fmt.Printf("Storage: Error saving auth: %v\n", err)
	} else {
		fmt.Printf("Storage: Auth saved successfully\n")
	}
	
	return err
}

func (s *Storage) ClearAuth() error {
	s.mutex.Lock()
	s.auth = nil
	s.mutex.Unlock()
	
	path := filepath.Join(s.dataDir, "auth.json")
	return os.Remove(path)
}

func (s *Storage) IsAuthenticated() bool {
	auth := s.GetAuth()
	if auth == nil {
		fmt.Printf("Storage: No auth data found\n")
		return false
	}
	
	isValid := auth.ExpiresAt.After(time.Now())
	fmt.Printf("Storage: Auth check - expires: %v, now: %v, valid: %v\n", auth.ExpiresAt, time.Now(), isValid)
	
	return isValid
}

func (s *Storage) loadAuth() error {
	path := filepath.Join(s.dataDir, "auth.json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}
	
	var auth AuthData
	if err := json.Unmarshal(data, &auth); err != nil {
		return err
	}
	
	s.auth = &auth
	return nil
}

func (s *Storage) saveAuth() error {
	if s.auth == nil {
		fmt.Printf("Storage: No auth data to save\n")
		return nil
	}
	
	path := filepath.Join(s.dataDir, "auth.json")
	fmt.Printf("Storage: Saving auth to file: %s\n", path)
	
	data, err := json.MarshalIndent(s.auth, "", "  ")
	if err != nil {
		fmt.Printf("Storage: Error marshaling auth data: %v\n", err)
		return err
	}
	
	fmt.Printf("Storage: Auth JSON data: %s\n", string(data))
	
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Printf("Storage: Error writing auth file: %v\n", err)
	} else {
		fmt.Printf("Storage: Auth file written successfully\n")
		
		// Verify the file was written
		if info, err := os.Stat(path); err == nil {
			fmt.Printf("Storage: Auth file size: %d bytes\n", info.Size())
		}
	}
	
	return err
}

// Drop progress tracking
func (s *Storage) SaveDropProgress(dropID string, minutes int) error {
	path := filepath.Join(s.dataDir, "drop_progress.json")
	
	var progress map[string]int
	
	// Load existing progress
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &progress)
	}
	
	if progress == nil {
		progress = make(map[string]int)
	}
	
	progress[dropID] = minutes
	
	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) GetDropProgress(dropID string) int {
	path := filepath.Join(s.dataDir, "drop_progress.json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	
	var progress map[string]int
	if err := json.Unmarshal(data, &progress); err != nil {
		return 0
	}
	
	if minutes, exists := progress[dropID]; exists {
		return minutes
	}
	
	return 0
}

// Statistics tracking
type Stats struct {
	TotalWatchTime     time.Duration `json:"totalWatchTime"`
	DropsEarned        int           `json:"dropsEarned"`
	StreamersWatched   []string      `json:"streamersWatched"`
	GamesWatched       []string      `json:"gamesWatched"`
	SessionStartTime   time.Time     `json:"sessionStartTime"`
	LastUpdateTime     time.Time     `json:"lastUpdateTime"`
}

func (s *Storage) GetStats() (*Stats, error) {
	path := filepath.Join(s.dataDir, "stats.json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Stats{
				StreamersWatched: make([]string, 0),
				GamesWatched:     make([]string, 0),
				SessionStartTime: time.Now(),
				LastUpdateTime:   time.Now(),
			}, nil
		}
		return nil, err
	}
	
	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	
	return &stats, nil
}

func (s *Storage) SaveStats(stats *Stats) error {
	stats.LastUpdateTime = time.Now()
	
	path := filepath.Join(s.dataDir, "stats.json")
	
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// Backup and restore functionality
func (s *Storage) CreateBackup() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("backup_%s.json", timestamp)
	backupPath := filepath.Join(s.dataDir, "backups", backupName)
	
	// Ensure backup directory exists
	os.MkdirAll(filepath.Dir(backupPath), 0755)
	
	backup := map[string]interface{}{
		"settings": s.GetSettings(),
		"games":    s.GetGames(),
		"auth":     s.GetAuth(),
		"timestamp": time.Now(),
	}
	
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return "", err
	}
	
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", err
	}
	
	return backupName, nil
}