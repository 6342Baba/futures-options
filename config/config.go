package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BinanceAPIKey          string
	BinanceSecretKey       string
	BinanceTestnet         bool
	BinanceFuturesTestnetURL string
	BinanceOptionsTestnetURL string
	MongoDBURI             string
	MongoDBDatabase         string
	Port                   string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		BinanceAPIKey:          getEnv("BINANCE_API_KEY", ""),
		BinanceSecretKey:       getEnv("BINANCE_SECRET_KEY", ""),
		BinanceTestnet:         getEnv("BINANCE_TESTNET", "true") == "true",
		BinanceFuturesTestnetURL: getEnv("BINANCE_FUTURES_TESTNET_URL", "https://testnet.binancefuture.com"),
		BinanceOptionsTestnetURL: getEnv("BINANCE_OPTIONS_TESTNET_URL", "https://testnet.binanceops.com"),
		MongoDBURI:             getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase:         getEnv("MONGODB_DATABASE", "futures_options_db"),
		Port:                   getEnv("PORT", "9090"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

