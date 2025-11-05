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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TradingService struct {
	binanceClient *binance.Client
	wsClient      *binance.WebSocketClient
}

func NewTradingService(binanceClient *binance.Client) *TradingService {
	return &TradingService{
		binanceClient: binanceClient,
	}
}

// GetAccountStatusWS retrieves account.status via WebSocket API
func (s *TradingService) GetAccountStatusWS(ctx context.Context) (interface{}, error) {
    ws, err := binance.NewWSAPIClient(s.binanceClient.Config)
    if err != nil { return nil, fmt.Errorf("failed to connect WS API: %w", err) }
    defer ws.Close()

    var result interface{}
    if err := ws.SendSignedRequest(ctx, fmt.Sprintf("status-%d", time.Now().UnixMilli()), "account.status", nil, &result); err != nil {
        return nil, err
    }
    return result, nil
}

// GetAccountBalanceWS retrieves account.balance via WebSocket API
func (s *TradingService) GetAccountBalanceWS(ctx context.Context) (interface{}, error) {
    ws, err := binance.NewWSAPIClient(s.binanceClient.Config)
    if err != nil { return nil, fmt.Errorf("failed to connect WS API: %w", err) }
    defer ws.Close()

    var result interface{}
    if err := ws.SendSignedRequest(ctx, fmt.Sprintf("bal-%d", time.Now().UnixMilli()), "account.balance", nil, &result); err != nil {
        return nil, err
    }
    return result, nil
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
	// Use Options client - create a config from binance client
	// For now, create a basic config (this would ideally come from binance.Client)
	// Note: We'll need to pass config through or store it in Client
	// Temporary workaround: create options client directly
	optionsClient := binance.NewOptionsClient(nil) // Will need proper config
	
	binanceReq := &binance.OptionsOrderRequest{
		Symbol:      req.Symbol,
		Side:        req.Side,
		OrderType:   req.OrderType,
		Quantity:    req.Quantity,
		Price:       req.Price,
		TimeInForce: "GTC",
	}

	binanceOrder, err := optionsClient.CreateOptionsOrder(ctx, binanceReq)
	if err != nil {
		// If API call fails, save as pending
		binanceOrder = nil
	}

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

	if binanceOrder != nil {
		optionsOrder.BinanceOrderID = binanceOrder.OrderID
		optionsOrder.Status = binanceOrder.Status
	}

	_, err = database.OptionsCollection.InsertOne(ctx, optionsOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to save order to database: %w", err)
	}

	return optionsOrder, nil
}

// GetOptionsPositions gets options positions
func (s *TradingService) GetOptionsPositions(ctx context.Context) ([]*models.Position, error) {
	optionsClient := binance.NewOptionsClient(nil) // Will need proper config
	binancePositions, err := optionsClient.GetOptionsPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get options positions: %w", err)
	}

	var positions []*models.Position
	for _, bp := range binancePositions {
		position := &models.Position{
			Symbol:       bp.Symbol,
			Type:         "OPTIONS",
			Quantity:     bp.Position,
			EntryPrice:   bp.EntryPrice,
			CurrentPrice: bp.MarkPrice,
			UnrealizedPnl: bp.UnrealizedPnl,
			UpdatedAt:    time.Now(),
		}
		positions = append(positions, position)
	}

	return positions, nil
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

// SaveAPICredentials saves API credentials to MongoDB
func (s *TradingService) SaveAPICredentials(ctx context.Context, req *SaveAPICredentialsRequest) (*models.APICredentials, error) {
	// Check if API key already exists
	filter := bson.M{"api_key": req.APIKey}
	existing := &models.APICredentials{}
	err := database.APICredentialsCollection.FindOne(ctx, filter).Decode(existing)
	
	if err == nil || err == mongo.ErrNoDocuments {
		if err == mongo.ErrNoDocuments {
			// Create new credentials
			credentials := &models.APICredentials{
				ID:        primitive.NewObjectID(),
				APIKey:    req.APIKey,
				SecretKey: req.SecretKey,
				IsActive:  req.IsActive,
				IsTestnet: req.IsTestnet,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err = database.APICredentialsCollection.InsertOne(ctx, credentials)
			if err != nil {
				return nil, fmt.Errorf("failed to save API credentials: %w", err)
			}

			return credentials, nil
		}
		// Update existing credentials
		existing.SecretKey = req.SecretKey
		existing.IsActive = req.IsActive
		existing.IsTestnet = req.IsTestnet
		existing.UpdatedAt = time.Now()

		update := bson.M{"$set": existing}
		_, err = database.APICredentialsCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update API credentials: %w", err)
		}
		return existing, nil
	}
	
	// If we got here, there was an unexpected error
	return nil, fmt.Errorf("unexpected error checking for existing credentials: %w", err)
}

// GetAPICredentials retrieves API credentials from MongoDB
func (s *TradingService) GetAPICredentials(ctx context.Context, activeOnly bool) ([]*models.APICredentials, error) {
	filter := bson.M{}
	if activeOnly {
		filter["is_active"] = true
	}

	cursor, err := database.APICredentialsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query API credentials: %w", err)
	}
	defer cursor.Close(ctx)

	var credentials []*models.APICredentials
	if err = cursor.All(ctx, &credentials); err != nil {
		return nil, fmt.Errorf("failed to decode API credentials: %w", err)
	}

	return credentials, nil
}

// GetActiveAPICredentials gets the first active API credentials
func (s *TradingService) GetActiveAPICredentials(ctx context.Context) (*models.APICredentials, error) {
	filter := bson.M{"is_active": true}
	credentials := &models.APICredentials{}
	err := database.APICredentialsCollection.FindOne(ctx, filter).Decode(credentials)
	if err != nil {
		return nil, fmt.Errorf("no active API credentials found: %w", err)
	}
	return credentials, nil
}

type SaveAPICredentialsRequest struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	IsActive  bool   `json:"is_active"`
	IsTestnet bool   `json:"is_testnet"`
}

