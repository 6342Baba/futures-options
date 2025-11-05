package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket          OrderType = "MARKET"
	OrderTypeLimit           OrderType = "LIMIT"
	OrderTypeStop            OrderType = "STOP"
	OrderTypeStopMarket      OrderType = "STOP_MARKET"
	OrderTypeStopLimit       OrderType = "STOP_LIMIT"
	OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitMarket OrderType = "TAKE_PROFIT_MARKET"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"
)

// TimeInForce represents order time in force
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate Or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill Or Kill
	TimeInForceGTX TimeInForce = "GTX" // Good Till Crossing (Post Only)
	TimeInForceGTD TimeInForce = "GTD" // Good Till Date
)

// SelfTradePreventionMode represents STP mode
type SelfTradePreventionMode string

const (
	STPNone        SelfTradePreventionMode = "NONE"
	STPExpireTaker SelfTradePreventionMode = "EXPIRE_TAKER"
	STPExpireBoth  SelfTradePreventionMode = "EXPIRE_BOTH"
	STPExpireMaker SelfTradePreventionMode = "EXPIRE_MAKER"
)

// PriceMatchMode represents price match mode
type PriceMatchMode string

const (
	PriceMatchNone      PriceMatchMode = "NONE"
	PriceMatchOpponent  PriceMatchMode = "OPPONENT"
	PriceMatchOpponent5 PriceMatchMode = "OPPONENT_5"
	PriceMatchQueue     PriceMatchMode = "QUEUE"
	PriceMatchQueue5    PriceMatchMode = "QUEUE_5"
	PriceMatchQueue10   PriceMatchMode = "QUEUE_10"
	PriceMatchQueue20   PriceMatchMode = "QUEUE_20"
)

// PositionMode represents position mode
type PositionMode string

const (
	PositionModeOneWay PositionMode = "ONEWAY"
	PositionModeHedge  PositionMode = "HEDGE"
)

// WorkingType represents working type for stop orders
type WorkingType string

const (
	WorkingTypeMarkPrice     WorkingType = "MARK_PRICE"
	WorkingTypeContractPrice WorkingType = "CONTRACT_PRICE"
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
	ID                    primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Symbol                string               `bson:"symbol" json:"symbol"`
	Side                  OrderSide            `bson:"side" json:"side"`
	OrderType             OrderType            `bson:"order_type" json:"order_type"`
	Quantity              float64              `bson:"quantity" json:"quantity"`
	Price                 float64              `bson:"price,omitempty" json:"price,omitempty"`
	StopPrice             float64              `bson:"stop_price,omitempty" json:"stop_price,omitempty"`
	ActivationPrice       float64              `bson:"activation_price,omitempty" json:"activation_price,omitempty"` // For TRAILING_STOP_MARKET
	CallbackRate          float64              `bson:"callback_rate,omitempty" json:"callback_rate,omitempty"`         // For TRAILING_STOP_MARKET
	Leverage              int                  `bson:"leverage" json:"leverage"`
	PositionSide          PositionSide          `bson:"position_side" json:"position_side"`
	TimeInForce           TimeInForce          `bson:"time_in_force,omitempty" json:"time_in_force,omitempty"`
	GoodTillDate          *time.Time           `bson:"good_till_date,omitempty" json:"good_till_date,omitempty"`
	WorkingType           WorkingType          `bson:"working_type,omitempty" json:"working_type,omitempty"`
	ReduceOnly            bool                 `bson:"reduce_only,omitempty" json:"reduce_only,omitempty"`
	ClosePosition         bool                 `bson:"close_position,omitempty" json:"close_position,omitempty"`
	SelfTradePreventionMode SelfTradePreventionMode `bson:"stp_mode,omitempty" json:"stp_mode,omitempty"`
	PriceMatch            PriceMatchMode       `bson:"price_match,omitempty" json:"price_match,omitempty"`
	NewOrderRespType      string               `bson:"new_order_resp_type,omitempty" json:"new_order_resp_type,omitempty"` // ACK, RESULT
	BinanceOrderID        int64                `bson:"binance_order_id,omitempty" json:"binance_order_id,omitempty"`
	ClientOrderID         string                `bson:"client_order_id,omitempty" json:"client_order_id,omitempty"`
	Status                string                `bson:"status" json:"status"`
	CreatedAt             time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time             `bson:"updated_at" json:"updated_at"`
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

// PositionModeConfig represents position mode configuration
type PositionModeConfig struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Mode          PositionMode       `bson:"mode" json:"mode"` // ONEWAY or HEDGE
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	EventType string      `json:"e"`
	EventTime int64       `json:"E"`
	Data      interface{} `json:"data"`
}

