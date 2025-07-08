package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"twitchdropsfarmer/internal/config"
	"twitchdropsfarmer/internal/drops"
	"twitchdropsfarmer/internal/storage"
	"twitchdropsfarmer/internal/twitch"
	"twitchdropsfarmer/internal/web"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logging
	logrus.SetLevel(logrus.InfoLevel) // Set to info to reduce verbose output
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	logrus.Info("Twitch Drops Farmer - Using Android app authentication (like TDM)")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Twitch client
	twitchClient := twitch.NewClient(cfg.TwitchClientID, cfg.TwitchClientSecret)

	// Initialize drop miner
	miner := drops.NewMiner(twitchClient, db)

	// Set miner configuration from loaded config
	minerConfig := &drops.MinerConfig{
		CheckInterval:     time.Duration(cfg.CheckInterval) * time.Second,
		WatchInterval:     20 * time.Second, // Like TDM - every ~20 seconds
		SwitchThreshold:   time.Duration(cfg.SwitchThreshold) * time.Minute,
		MinimumPoints:     cfg.MinimumPoints,
		MaximumStreams:    cfg.MaximumStreams,
		PriorityGames:     cfg.PriorityGames,
		ExcludeGames:      cfg.ExcludeGames,
		WatchUnlisted:     cfg.WatchUnlisted,
		ClaimDrops:        cfg.ClaimDrops,
		WebhookURL:        cfg.WebhookURL,
	}
	miner.SetConfig(minerConfig)

	// Initialize web server
	webServer := web.NewServer(cfg, twitchClient, miner, db)

	// Start web server
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: webServer.Router(),
	}

	// Start server in goroutine
	go func() {
		logrus.Infof("Starting server on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start drop miner
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := miner.Start(ctx); err != nil {
			logrus.Errorf("Drop miner error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Cancel miner context
	cancel()

	// Shutdown server gracefully
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}