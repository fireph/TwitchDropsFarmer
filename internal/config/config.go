package config

// Config holds the application configuration
type Config struct {
	Port        string `env:"PORT"`
	Environment string `env:"ENVIRONMENT,default=development"`
	DataDir     string `env:"DATA_DIR"`
	
	// Twitch API credentials - using the same client ID as TwitchDropsMiner
	TwitchClientID     string `env:"TWITCH_CLIENT_ID,default=kimne78kx3ncx6brgo4mv6wki5h1ko"`
	TwitchClientSecret string `env:"TWITCH_CLIENT_SECRET"`
}