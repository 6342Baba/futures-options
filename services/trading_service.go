package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"futures-options/binance"
	"futures-options/database"
	"futures-options/models"

	"github.com/adshao/go-binance/v2/futures"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TradingService struct {
	binanceClient *binance.Client
}

func NewTradingService(binanceClient *binance.Client) *TradingService {
	return &TradingService{
		binanceClient: binanceClient,
	}
}

// CreateFuturesOrder creates a futures order and saves it to MongoDB
func (s *TradingService) CreateFuturesOrder(ctx context.Context, req *CreateFuturesOrderRequest) (*models.FuturesOrder, error) {
	// Convert to Binance types
	var side futures.SideType
	if req.Side == string(models.OrderSideBuy) {
		side = futures.SideTypeBuy
	} else {
		side = futures.SideTypeSell
	}

	var orderType futures.OrderType
	if req.OrderType == string(models.OrderTypeMarket) {
		orderType = futures.OrderTypeMarket
	} else {
		orderType = futures.OrderTypeLimit
	}

	// Create order on Binance
	binanceOrder, err := s.binanceClient.CreateFuturesOrder(
		ctx,
		req.Symbol,
		side,
		orderType,
		req.Quantity,
		req.Price,
		req.Leverage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order on Binance: %w", err)
	}

	// Save to MongoDB
	futuresOrder := &models.FuturesOrder{
		ID:            primitive.NewObjectID(),
		Symbol:        req.Symbol,
		Side:          models.OrderSide(req.Side),
		OrderType:     models.OrderType(req.OrderType),
		Quantity:      req.Quantity,
		Price:         req.Price,
		Leverage:      req.Leverage,
		PositionSide:  models.PositionSide(req.PositionSide),
		BinanceOrderID: binanceOrder.OrderID,
		Status:        string(binanceOrder.Status),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = database.FuturesCollection.InsertOne(ctx, futuresOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to save order to database: %w", err)
	}

	return futuresOrder, nil
}

// CreateOptionsOrder creates an options order and saves it to MongoDB
func (s *TradingService) CreateOptionsOrder(ctx context.Context, req *CreateOptionsOrderRequest) (*models.OptionsOrder, error) {
	// Note: This is a placeholder - you'll need to implement actual Options API calls
	// For now, we'll just save to database

	optionsOrder := &models.OptionsOrder{
		ID:            primitive.NewObjectID(),
		Symbol:        req.Symbol,
		Side:          models.OrderSide(req.Side),
		OrderType:     models.OrderType(req.OrderType),
		Quantity:      req.Quantity,
		Price:         req.Price,
		StrikePrice:   req.StrikePrice,
		ExpiryDate:    req.ExpiryDate,
		OptionType:    req.OptionType,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := database.OptionsCollection.InsertOne(ctx, optionsOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to save order to database: %w", err)
	}

	return optionsOrder, nil
}

// GetFuturesOrders retrieves futures orders from MongoDB
func (s *TradingService) GetFuturesOrders(ctx context.Context, symbol string) ([]*models.FuturesOrder, error) {
	filter := bson.M{}
	if symbol != "" {
		filter["symbol"] = symbol
	}

	cursor, err := database.FuturesCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query futures orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*models.FuturesOrder
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode futures orders: %w", err)
	}

	return orders, nil
}

// GetOptionsOrders retrieves options orders from MongoDB
func (s *TradingService) GetOptionsOrders(ctx context.Context, symbol string) ([]*models.OptionsOrder, error) {
	filter := bson.M{}
	if symbol != "" {
		filter["symbol"] = symbol
	}

	cursor, err := database.OptionsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query options orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*models.OptionsOrder
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode options orders: %w", err)
	}

	return orders, nil
}

// GetPositions retrieves positions from MongoDB
func (s *TradingService) GetPositions(ctx context.Context, positionType string) ([]*models.Position, error) {
	filter := bson.M{}
	if positionType != "" {
		filter["type"] = positionType
	}

	cursor, err := database.PositionsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer cursor.Close(ctx)

	var positions []*models.Position
	if err = cursor.All(ctx, &positions); err != nil {
		return nil, fmt.Errorf("failed to decode positions: %w", err)
	}

	return positions, nil
}

// SyncPositionsFromBinance syncs positions from Binance to MongoDB
func (s *TradingService) SyncPositionsFromBinance(ctx context.Context) error {
	// Get positions from Binance
	binancePositions, err := s.binanceClient.GetFuturesPositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get positions from Binance: %w", err)
	}

	// Update positions in MongoDB
	for _, bp := range binancePositions {
		positionSize, _ := strconv.ParseFloat(bp.PositionAmt, 64)
		if positionSize == 0 {
			continue // Skip zero positions
		}

		entryPrice, _ := strconv.ParseFloat(bp.EntryPrice, 64)
		unrealizedPnl, _ := strconv.ParseFloat(bp.UnRealizedProfit, 64)
		leverage, _ := strconv.Atoi(bp.Leverage)

		position := &models.Position{
			Symbol:       bp.Symbol,
			Type:         "FUTURES",
			Side:         models.PositionSide(bp.PositionSide),
			Quantity:     positionSize,
			EntryPrice:   entryPrice,
			UnrealizedPnl: unrealizedPnl,
			Leverage:     leverage,
			UpdatedAt:    time.Now(),
		}

		// Check if position exists
		filter := bson.M{"symbol": bp.Symbol, "type": "FUTURES"}
		update := bson.M{"$set": position}

		opts := options.Update().SetUpsert(true)
		_, err = database.PositionsCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("failed to update position: %w", err)
		}
	}

	return nil
}

// Request types
type CreateFuturesOrderRequest struct {
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"` // BUY or SELL
	OrderType    string  `json:"order_type"` // MARKET or LIMIT
	Quantity     float64 `json:"quantity"`
	Price        float64 `json:"price,omitempty"`
	Leverage     int     `json:"leverage"`
	PositionSide string  `json:"position_side"` // LONG or SHORT
}

type CreateOptionsOrderRequest struct {
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"` // BUY or SELL
	OrderType  string    `json:"order_type"` // MARKET or LIMIT
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price,omitempty"`
	StrikePrice float64  `json:"strike_price"`
	ExpiryDate time.Time `json:"expiry_date"`
	OptionType string    `json:"option_type"` // CALL or PUT
}

