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

	// Note: API keys will be loaded from database first (if saved via POST /api/credentials),
	// then fall back to environment variables if not found in database

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
	
	// Create temporary service to check database for credentials
	tempService := services.NewTradingService(binanceClient)
	
	// Priority: Database first, then environment variables
	var apiKey, secretKey string
	var keySource string
	
	// Try to load from database first (credentials saved via API)
	credentials, err := tempService.GetActiveAPICredentials(context.Background())
	if err == nil && credentials.APIKey != "" && credentials.SecretKey != "" {
		apiKey = credentials.APIKey
		secretKey = credentials.SecretKey
		keySource = "database"
		log.Printf("✓ Using API keys from database (saved via POST /api/credentials)")
		// Show masked API key for security
		keyLen := len(credentials.APIKey)
		prefix := ""
		suffix := ""
		if keyLen > 8 {
			prefix = credentials.APIKey[:8]
		} else {
			prefix = credentials.APIKey
		}
		if keyLen > 4 {
			suffix = credentials.APIKey[keyLen-4:]
		}
		if keyLen > 12 {
			log.Printf("  API Key: %s...%s (testnet: %v)", prefix, suffix, credentials.IsTestnet)
		} else {
			log.Printf("  API Key: [configured] (testnet: %v)", credentials.IsTestnet)
		}
	} else if cfg.BinanceAPIKey != "" && cfg.BinanceSecretKey != "" {
		// Fall back to environment variables
		apiKey = cfg.BinanceAPIKey
		secretKey = cfg.BinanceSecretKey
		keySource = "environment"
		log.Println("✓ Using API keys from environment variables")
	} else {
		log.Println("⚠ Warning: No API keys found in database or environment")
		log.Println("  Please add API keys via: POST /api/credentials")
		log.Println("  Or set BINANCE_API_KEY and BINANCE_SECRET_KEY in .env file")
	}
	
	// Set API keys if we found them
	if apiKey != "" && secretKey != "" {
		binanceClient.SetAPIKeys(apiKey, secretKey)
		log.Printf("✓ Binance client configured with API keys from %s", keySource)
	}

	// Initialize services (reuse the temp service)
	tradingService := tempService

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

