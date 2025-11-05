package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"futures-options/binance"
	"futures-options/config"
	"futures-options/database"
	_ "futures-options/docs" // Swagger docs (blank import to ensure docs package is linked)
	"futures-options/handlers"
	"futures-options/services"
)

// @title           Binance Futures & Options Trading API
// @version         1.0
// @description     A REST API for trading Binance Futures and Options using testnet/demo accounts
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9090
// @BasePath  /

// @schemes http https

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if cfg.BinanceAPIKey == "" || cfg.BinanceSecretKey == "" {
		log.Println("Warning: Binance API keys not set. Some features may not work.")
		log.Println("Please set BINANCE_API_KEY and BINANCE_SECRET_KEY in your .env file")
	}

	// Connect to MongoDB
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	// Create indexes
	if err := database.CreateIndexes(); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize Binance client
	binanceClient := binance.NewClient(cfg)
	
	// Try to load API keys from environment first
	if cfg.BinanceAPIKey != "" && cfg.BinanceSecretKey != "" {
		binanceClient.SetAPIKeys(cfg.BinanceAPIKey, cfg.BinanceSecretKey)
		log.Println("Using API keys from environment variables")
	} else {
		// Try to load from database
		tradingService := services.NewTradingService(binanceClient)
		credentials, err := tradingService.GetActiveAPICredentials(context.Background())
		if err == nil {
			binanceClient.SetAPIKeys(credentials.APIKey, credentials.SecretKey)
			log.Println("Using API keys from database")
		} else {
			log.Println("No API keys found. Please add API keys via POST /api/credentials")
		}
	}

	// Initialize services
	var tradingService *services.TradingService
	if cfg.BinanceAPIKey == "" || cfg.BinanceSecretKey == "" {
		// Reuse the service created above for loading credentials
		tradingService = services.NewTradingService(binanceClient)
	} else {
		tradingService = services.NewTradingService(binanceClient)
	}

	// Initialize handlers
	h := handlers.NewHandlers(tradingService)

	// Setup routes
	router := handlers.SetupRoutes(h)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("Testnet mode: %v", cfg.BinanceTestnet)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

