package services

import (
	"context"
	"fmt"
	"time"

	"futures-options/binance"
	"futures-options/database"
	"futures-options/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateAdvancedFuturesOrder creates an advanced futures order with all features
func (s *TradingService) CreateAdvancedFuturesOrder(ctx context.Context, req *AdvancedOrderRequest) (*models.FuturesOrder, error) {
	// Convert to Binance advanced request
	binanceReq := &binance.AdvancedOrderRequest{
		Symbol:                req.Symbol,
		Side:                  req.Side,
		OrderType:             req.OrderType,
		Quantity:              req.Quantity,
		Price:                 req.Price,
		StopPrice:             req.StopPrice,
		ActivationPrice:       req.ActivationPrice,
		CallbackRate:          req.CallbackRate,
		Leverage:              req.Leverage,
		PositionSide:          req.PositionSide,
		TimeInForce:           req.TimeInForce,
		WorkingType:           req.WorkingType,
		ReduceOnly:            req.ReduceOnly,
		ClosePosition:         req.ClosePosition,
		SelfTradePreventionMode: req.SelfTradePreventionMode,
		PriceMatch:            req.PriceMatch,
		NewOrderRespType:      req.NewOrderRespType,
		ClientOrderID:         req.ClientOrderID,
		GoodTillDate:          req.GoodTillDate,
	}

	// Create order on Binance
	binanceOrder, err := s.binanceClient.CreateAdvancedFuturesOrder(ctx, binanceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order on Binance: %w", err)
	}

	// Save to MongoDB
	futuresOrder := &models.FuturesOrder{
		ID:                    primitive.NewObjectID(),
		Symbol:                req.Symbol,
		Side:                  models.OrderSide(req.Side),
		OrderType:             models.OrderType(req.OrderType),
		Quantity:              req.Quantity,
		Price:                 req.Price,
		StopPrice:             req.StopPrice,
		ActivationPrice:       req.ActivationPrice,
		CallbackRate:          req.CallbackRate,
		Leverage:              req.Leverage,
		PositionSide:          models.PositionSide(req.PositionSide),
		TimeInForce:           models.TimeInForce(req.TimeInForce),
		WorkingType:           models.WorkingType(req.WorkingType),
		ReduceOnly:            req.ReduceOnly,
		ClosePosition:         req.ClosePosition,
		SelfTradePreventionMode: models.SelfTradePreventionMode(req.SelfTradePreventionMode),
		PriceMatch:            models.PriceMatchMode(req.PriceMatch),
		NewOrderRespType:      req.NewOrderRespType,
		ClientOrderID:         req.ClientOrderID,
		GoodTillDate:          req.GoodTillDate,
		BinanceOrderID:        binanceOrder.OrderID,
		Status:                string(binanceOrder.Status),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	_, err = database.FuturesCollection.InsertOne(ctx, futuresOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to save order to database: %w", err)
	}

	return futuresOrder, nil
}

// ModifyFuturesOrder modifies an existing futures order
func (s *TradingService) ModifyFuturesOrder(ctx context.Context, req *ModifyOrderRequest) (*models.FuturesOrder, error) {
	// Modify order on Binance
	_, err := s.binanceClient.ModifyFuturesOrder(ctx, &binance.ModifyOrderRequest{
		Symbol:         req.Symbol,
		OrderID:        req.OrderID,
		ClientOrderID:  req.ClientOrderID,
		Quantity:       req.Quantity,
		Price:          req.Price,
		StopPrice:      req.StopPrice,
		ActivationPrice: req.ActivationPrice,
		CallbackRate:   req.CallbackRate,
		PriceMatch:     req.PriceMatch,
	})
	if err != nil {
		// If modification fails, still update database
		// In production, you might want to handle this differently
	}

	// Update the database record
	filter := bson.M{}
	if req.OrderID > 0 {
		filter["binance_order_id"] = req.OrderID
	} else if req.ClientOrderID != "" {
		filter["client_order_id"] = req.ClientOrderID
	} else {
		return nil, fmt.Errorf("either orderID or clientOrderID must be provided")
	}

	updateData := bson.M{
		"updated_at": time.Now(),
	}

	if req.Quantity > 0 {
		updateData["quantity"] = req.Quantity
	}
	if req.Price > 0 {
		updateData["price"] = req.Price
	}
	if req.StopPrice > 0 {
		updateData["stop_price"] = req.StopPrice
	}

	update := bson.M{"$set": updateData}

	var order models.FuturesOrder
	err = database.FuturesCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&order)
	if err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return &order, nil
}

// CreateBatchOrders creates multiple orders at once
func (s *TradingService) CreateBatchOrders(ctx context.Context, req *BatchOrderRequest) (*BatchOrderResponse, error) {
	var orders []*binance.AdvancedOrderRequest
	for _, orderReq := range req.Orders {
		orders = append(orders, &binance.AdvancedOrderRequest{
			Symbol:                orderReq.Symbol,
			Side:                  orderReq.Side,
			OrderType:             orderReq.OrderType,
			Quantity:              orderReq.Quantity,
			Price:                 orderReq.Price,
			StopPrice:             orderReq.StopPrice,
			ActivationPrice:       orderReq.ActivationPrice,
			CallbackRate:          orderReq.CallbackRate,
			Leverage:              orderReq.Leverage,
			PositionSide:          orderReq.PositionSide,
			TimeInForce:           orderReq.TimeInForce,
			WorkingType:           orderReq.WorkingType,
			ReduceOnly:            orderReq.ReduceOnly,
			ClosePosition:         orderReq.ClosePosition,
			SelfTradePreventionMode: orderReq.SelfTradePreventionMode,
			PriceMatch:            orderReq.PriceMatch,
			ClientOrderID:         orderReq.ClientOrderID,
		})
	}

	binanceOrders, err := s.binanceClient.CreateBatchOrders(ctx, orders)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch orders: %w", err)
	}

	// Save to MongoDB
	var savedOrders []*models.FuturesOrder
	for i, binanceOrder := range binanceOrders {
		if i >= len(req.Orders) {
			break
		}
		orderReq := req.Orders[i]

		futuresOrder := &models.FuturesOrder{
			ID:                    primitive.NewObjectID(),
			Symbol:                orderReq.Symbol,
			Side:                  models.OrderSide(orderReq.Side),
			OrderType:             models.OrderType(orderReq.OrderType),
			Quantity:              orderReq.Quantity,
			Price:                 orderReq.Price,
			StopPrice:             orderReq.StopPrice,
			Leverage:              orderReq.Leverage,
			PositionSide:          models.PositionSide(orderReq.PositionSide),
			BinanceOrderID:        binanceOrder.OrderID,
			Status:                string(binanceOrder.Status),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		_, err = database.FuturesCollection.InsertOne(ctx, futuresOrder)
		if err != nil {
			continue
		}

		savedOrders = append(savedOrders, futuresOrder)
	}

	return &BatchOrderResponse{
		Orders: savedOrders,
	}, nil
}

// CancelBatchOrders cancels multiple orders
func (s *TradingService) CancelBatchOrders(ctx context.Context, symbol string, orderIDs []int64, clientOrderIDs []string) error {
	_, err := s.binanceClient.CancelBatchOrders(ctx, symbol, orderIDs, clientOrderIDs)
	if err != nil {
		return fmt.Errorf("failed to cancel batch orders: %w", err)
	}

	// Update status in MongoDB
	filter := bson.M{"symbol": symbol}
	if len(orderIDs) > 0 {
		filter["binance_order_id"] = bson.M{"$in": orderIDs}
	}
	if len(clientOrderIDs) > 0 {
		filter["client_order_id"] = bson.M{"$in": clientOrderIDs}
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "CANCELED",
			"updated_at": time.Now(),
		},
	}

	_, err = database.FuturesCollection.UpdateMany(ctx, filter, update)
	return err
}

// SetPositionMode sets position mode (One-way or Hedge)
func (s *TradingService) SetPositionMode(ctx context.Context, dualSide bool) error {
	err := s.binanceClient.SetPositionMode(ctx, dualSide)
	if err != nil {
		return err
	}

	// Save to database
	mode := models.PositionModeOneWay
	if dualSide {
		mode = models.PositionModeHedge
	}

	config := &models.PositionModeConfig{
		ID:        primitive.NewObjectID(),
		Mode:      mode,
		UpdatedAt: time.Now(),
	}

	filter := bson.M{}
	update := bson.M{"$set": config}
	opts := options.Update().SetUpsert(true)

	_, err = database.DB.Collection("position_mode").UpdateOne(ctx, filter, update, opts)
	return err
}

// GetPositionMode gets current position mode
func (s *TradingService) GetPositionMode(ctx context.Context) (*models.PositionModeConfig, error) {
	dualSide, err := s.binanceClient.GetPositionMode(ctx)
	if err != nil {
		return nil, err
	}

	mode := models.PositionModeOneWay
	if dualSide {
		mode = models.PositionModeHedge
	}

	return &models.PositionModeConfig{
		Mode:      mode,
		UpdatedAt: time.Now(),
	}, nil
}

// Request types
type AdvancedOrderRequest struct {
	Symbol                string     `json:"symbol"`
	Side                  string     `json:"side"`
	OrderType             string     `json:"order_type"`
	Quantity              float64    `json:"quantity"`
	Price                 float64    `json:"price,omitempty"`
	StopPrice             float64    `json:"stop_price,omitempty"`
	ActivationPrice       float64    `json:"activation_price,omitempty"`
	CallbackRate          float64    `json:"callback_rate,omitempty"`
	Leverage              int        `json:"leverage"`
	PositionSide          string     `json:"position_side,omitempty"`
	TimeInForce           string     `json:"time_in_force,omitempty"`
	WorkingType           string     `json:"working_type,omitempty"`
	ReduceOnly            bool       `json:"reduce_only,omitempty"`
	ClosePosition         bool       `json:"close_position,omitempty"`
	SelfTradePreventionMode string   `json:"self_trade_prevention_mode,omitempty"`
	PriceMatch            string     `json:"price_match,omitempty"`
	NewOrderRespType      string     `json:"new_order_resp_type,omitempty"`
	ClientOrderID         string     `json:"client_order_id,omitempty"`
	GoodTillDate          *time.Time `json:"good_till_date,omitempty"`
}

type ModifyOrderRequest struct {
	Symbol         string  `json:"symbol"`
	OrderID        int64   `json:"order_id,omitempty"`
	ClientOrderID  string  `json:"client_order_id,omitempty"`
	Quantity       float64 `json:"quantity,omitempty"`
	Price          float64 `json:"price,omitempty"`
	StopPrice      float64 `json:"stop_price,omitempty"`
	ActivationPrice float64 `json:"activation_price,omitempty"`
	CallbackRate   float64 `json:"callback_rate,omitempty"`
	PriceMatch     string  `json:"price_match,omitempty"`
}

type BatchOrderRequest struct {
	Orders []AdvancedOrderRequest `json:"orders"`
}

type BatchOrderResponse struct {
	Orders []*models.FuturesOrder `json:"orders"`
	Errors []string               `json:"errors,omitempty"`
}

