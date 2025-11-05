package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderSide represents buy or sell
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// PositionSide represents long or short
type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

// FuturesOrder represents a futures trading order
type FuturesOrder struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol        string             `bson:"symbol" json:"symbol"`
	Side          OrderSide          `bson:"side" json:"side"`
	OrderType     OrderType          `bson:"order_type" json:"order_type"`
	Quantity      float64            `bson:"quantity" json:"quantity"`
	Price         float64            `bson:"price,omitempty" json:"price,omitempty"`
	Leverage      int                `bson:"leverage" json:"leverage"`
	PositionSide  PositionSide       `bson:"position_side" json:"position_side"`
	BinanceOrderID int64             `bson:"binance_order_id,omitempty" json:"binance_order_id,omitempty"`
	Status        string             `bson:"status" json:"status"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// OptionsOrder represents an options trading order
type OptionsOrder struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol        string             `bson:"symbol" json:"symbol"`
	Side          OrderSide          `bson:"side" json:"side"`
	OrderType     OrderType          `bson:"order_type" json:"order_type"`
	Quantity      float64            `bson:"quantity" json:"quantity"`
	Price         float64            `bson:"price,omitempty" json:"price,omitempty"`
	StrikePrice   float64            `bson:"strike_price" json:"strike_price"`
	ExpiryDate    time.Time          `bson:"expiry_date" json:"expiry_date"`
	OptionType    string             `bson:"option_type" json:"option_type"` // CALL or PUT
	BinanceOrderID int64             `bson:"binance_order_id,omitempty" json:"binance_order_id,omitempty"`
	Status        string             `bson:"status" json:"status"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// Position represents an open position
type Position struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol        string             `bson:"symbol" json:"symbol"`
	Type          string             `bson:"type" json:"type"` // FUTURES or OPTIONS
	Side          PositionSide       `bson:"side" json:"side"`
	Quantity      float64            `bson:"quantity" json:"quantity"`
	EntryPrice    float64            `bson:"entry_price" json:"entry_price"`
	CurrentPrice  float64            `bson:"current_price,omitempty" json:"current_price,omitempty"`
	UnrealizedPnl float64            `bson:"unrealized_pnl,omitempty" json:"unrealized_pnl,omitempty"`
	Leverage      int                `bson:"leverage,omitempty" json:"leverage,omitempty"`
	StrikePrice   float64            `bson:"strike_price,omitempty" json:"strike_price,omitempty"`
	ExpiryDate    time.Time          `bson:"expiry_date,omitempty" json:"expiry_date,omitempty"`
	OptionType    string             `bson:"option_type,omitempty" json:"option_type,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// APICredentials represents Binance API credentials stored in database
type APICredentials struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	APIKey        string             `bson:"api_key" json:"api_key"`
	SecretKey     string             `bson:"secret_key" json:"secret_key"`
	IsActive      bool               `bson:"is_active" json:"is_active"`
	IsTestnet     bool               `bson:"is_testnet" json:"is_testnet"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

