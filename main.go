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
	"futures-options/handlers"
	"futures-options/services"
)

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
	if cfg.BinanceAPIKey != "" && cfg.BinanceSecretKey != "" {
		binanceClient.SetAPIKeys(cfg.BinanceAPIKey, cfg.BinanceSecretKey)
	}

	// Initialize services
	tradingService := services.NewTradingService(binanceClient)

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

