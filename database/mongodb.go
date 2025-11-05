package database

import (
	"context"
	"fmt"
	"time"

	"futures-options/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client     *mongo.Client
	DB         *mongo.Database
	FuturesCollection *mongo.Collection
	OptionsCollection *mongo.Collection
	PositionsCollection *mongo.Collection
)

func Connect(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoDBURI)

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	err = Client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	DB = Client.Database(cfg.MongoDBDatabase)
	FuturesCollection = DB.Collection("futures_orders")
	OptionsCollection = DB.Collection("options_orders")
	PositionsCollection = DB.Collection("positions")

	fmt.Println("Connected to MongoDB successfully!")
	return nil
}

func Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Client.Disconnect(ctx)
}

// CreateIndexes creates indexes for better query performance
func CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Futures orders indexes
	futuresIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"symbol": 1, "created_at": -1}},
		{Keys: map[string]interface{}{"binance_order_id": 1}, Options: options.Index().SetUnique(true)},
	}

	// Options orders indexes
	optionsIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"symbol": 1, "created_at": -1}},
		{Keys: map[string]interface{}{"binance_order_id": 1}, Options: options.Index().SetUnique(true)},
	}

	// Positions indexes
	positionsIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"symbol": 1, "type": 1}},
		{Keys: map[string]interface{}{"created_at": -1}},
	}

	_, err := FuturesCollection.Indexes().CreateMany(ctx, futuresIndexes)
	if err != nil {
		return fmt.Errorf("failed to create futures indexes: %w", err)
	}

	_, err = OptionsCollection.Indexes().CreateMany(ctx, optionsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create options indexes: %w", err)
	}

	_, err = PositionsCollection.Indexes().CreateMany(ctx, positionsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create positions indexes: %w", err)
	}

	fmt.Println("Indexes created successfully!")
	return nil
}

