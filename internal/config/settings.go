package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// GameConfig represents a game with name, slug, and ID
type GameConfig struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	ID   string `json:"id"`
}

type Config struct {
	// Server configuration
	ServerAddress string `json:"server_address"`
	DatabasePath  string `json:"database_path"`

	// Twitch API configuration
	TwitchClientID     string `json:"twitch_client_id"`
	TwitchClientSecret string `json:"twitch_client_secret"`

	// Drop mining configuration
	PriorityGames   []GameConfig `json:"priority_games"`
	ExcludeGames    []GameConfig `json:"exclude_games"`
	WatchUnlisted   bool         `json:"watch_unlisted"`
	ClaimDrops      bool         `json:"claim_drops"`
	WebhookURL      string       `json:"webhook_url"`
	CheckInterval   int          `json:"check_interval"`   // seconds
	SwitchThreshold int          `json:"switch_threshold"` // minutes
	MinimumPoints   int          `json:"minimum_points"`
	MaximumStreams  int          `json:"maximum_streams"`

	// UI configuration
	Theme          string `json:"theme"` // "light" or "dark"
	Language       string `json:"language"`
	ShowTray       bool   `json:"show_tray"`
	StartMinimized bool   `json:"start_minimized"`
}

func Load() (*Config, error) {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		ServerAddress:      getEnv("SERVER_ADDRESS", ":8080"),
		DatabasePath:       getEnv("DATABASE_PATH", filepath.Join(".", "config", "drops.db")),
		TwitchClientID:     getEnv("TWITCH_CLIENT_ID", "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"), // Twitch Android App ID (like TDM)
		TwitchClientSecret: getEnv("TWITCH_CLIENT_SECRET", ""),                            // Not needed for device flow
		PriorityGames:      []GameConfig{},
		ExcludeGames:       []GameConfig{},
		WatchUnlisted:      true,
		ClaimDrops:         true,
		WebhookURL:         getEnv("WEBHOOK_URL", ""),
		CheckInterval:      60,
		SwitchThreshold:    5,
		MinimumPoints:      50,
		MaximumStreams:     3,
		Theme:              "dark",
		Language:           "en",
		ShowTray:           true,
		StartMinimized:     false,
	}

	// Load configuration from file if it exists
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func getConfigPath() string {
	// Store config in ./config directory for portability
	return filepath.Join(".", "config", "config.json")
}

func getTokenPath() string {
	// Store auth tokens in ./config directory
	return filepath.Join(".", "config", "token.json")
}

func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Token storage functions
type StoredToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Expiry      time.Time `json:"expiry"`
}

func SaveToken(token *oauth2.Token) error {
	tokenPath := getTokenPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0755); err != nil {
		return err
	}

	storedToken := StoredToken{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Expiry:      token.Expiry,
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600) // 0600 for security
}

func LoadToken() (*oauth2.Token, error) {
	tokenPath := getTokenPath()

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var storedToken StoredToken
	if err := json.Unmarshal(data, &storedToken); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken: storedToken.AccessToken,
		TokenType:   storedToken.TokenType,
		Expiry:      storedToken.Expiry,
	}

	return token, nil
}

func DeleteToken() error {
	tokenPath := getTokenPath()
	return os.Remove(tokenPath)
}

// AddGameToConfig adds a game to the configuration with slug and ID resolution
func (c *Config) AddGameToConfig(gameName string, gameSlug string, gameID string, toPriority bool) error {
	gameConfig := GameConfig{
		Name: gameName,
		Slug: gameSlug,
		ID:   gameID,
	}

	if toPriority {
		// Check if game already exists in priority games
		for i, existing := range c.PriorityGames {
			if existing.Name == gameName {
				// Update existing entry with slug and ID
				c.PriorityGames[i].Slug = gameSlug
				c.PriorityGames[i].ID = gameID
				logrus.Infof("Updated existing priority game '%s' with slug '%s' and ID '%s'", gameName, gameSlug, gameID)
				return c.Save()
			}
		}
		c.PriorityGames = append(c.PriorityGames, gameConfig)
		logrus.Infof("Added new priority game '%s' with slug '%s' and ID '%s'", gameName, gameSlug, gameID)
	} else {
		// Check if game already exists in exclude games
		for i, existing := range c.ExcludeGames {
			if existing.Name == gameName {
				// Update existing entry with slug and ID
				c.ExcludeGames[i].Slug = gameSlug
				c.ExcludeGames[i].ID = gameID
				logrus.Infof("Updated existing exclude game '%s' with slug '%s' and ID '%s'", gameName, gameSlug, gameID)
				return c.Save()
			}
		}
		c.ExcludeGames = append(c.ExcludeGames, gameConfig)
		logrus.Infof("Added new exclude game '%s' with slug '%s' and ID '%s'", gameName, gameSlug, gameID)
	}

	logrus.Infof("About to save config with %d priority games and %d exclude games", 
		len(c.PriorityGames), len(c.ExcludeGames))
	return c.Save()
}

// GetGameSlugOrEmpty returns the slug for a game name, or empty string if not found
func (c *Config) GetGameSlugOrEmpty(gameName string) string {
	// Check priority games first
	for _, game := range c.PriorityGames {
		if game.Name == gameName {
			return game.Slug
		}
	}

	// Check exclude games
	for _, game := range c.ExcludeGames {
		if game.Name == gameName {
			return game.Slug
		}
	}

	return ""
}
