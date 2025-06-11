package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"twitchdropsfarmer/internal/api"
	"twitchdropsfarmer/internal/config"
	"twitchdropsfarmer/internal/miner"
	"twitchdropsfarmer/internal/storage"
	"twitchdropsfarmer/internal/twitch"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	// Load configuration
	var cfg config.Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set default values
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "/data"
	}

	// Initialize storage
	store := storage.New(cfg.DataDir)

	// Initialize Twitch client
	twitchClient := twitch.New(cfg.TwitchClientID, cfg.TwitchClientSecret)

	// Initialize miner
	minerInstance := miner.New(store, twitchClient)

	// Initialize API
	apiHandler := api.New(store, twitchClient, minerInstance)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Request logging middleware
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// Serve static files
	router.Use(static.Serve("/", static.LocalFile("./web", true)))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	apiV1 := router.Group("/api/v1")
	{
		// Authentication
		apiV1.GET("/auth/url", apiHandler.GetAuthURL)
		apiV1.POST("/auth/callback", apiHandler.HandleAuthCallback)
		apiV1.GET("/auth/status", apiHandler.GetAuthStatus)
		apiV1.DELETE("/auth/logout", apiHandler.Logout)

		// Games and watching
		apiV1.GET("/games", apiHandler.GetGames)
		apiV1.POST("/games", apiHandler.AddGame)
		apiV1.DELETE("/games/:id", apiHandler.RemoveGame)
		apiV1.PUT("/games/reorder", apiHandler.ReorderGames)

		// Drops
		apiV1.GET("/drops", apiHandler.GetDrops)
		apiV1.GET("/drops/current", apiHandler.GetCurrentDrop)
		apiV1.POST("/drops/:id/claim", apiHandler.ClaimDrop)
		apiV1.POST("/points/claim", apiHandler.ClaimCommunityPoints)

		// Miner control
		apiV1.POST("/miner/start", apiHandler.StartMiner)
		apiV1.POST("/miner/stop", apiHandler.StopMiner)
		apiV1.GET("/miner/status", apiHandler.GetMinerStatus)
		apiV1.GET("/miner/logs", apiHandler.GetMinerLogs)

		// Settings
		apiV1.GET("/settings", apiHandler.GetSettings)
		apiV1.PUT("/settings", apiHandler.UpdateSettings)

		// WebSocket for real-time updates
		apiV1.GET("/ws", apiHandler.HandleWebSocket)
	}

	// Fallback for SPA routing
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start miner in background
	go minerInstance.Start()

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("TwitchDropsFarmer started on port %s", cfg.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop miner
	minerInstance.Stop()

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}