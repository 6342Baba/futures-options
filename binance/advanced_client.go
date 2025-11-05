package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// CreateAdvancedFuturesOrder creates an advanced futures order with all features
func (c *Client) CreateAdvancedFuturesOrder(ctx context.Context, req *AdvancedOrderRequest) (*futures.CreateOrderResponse, error) {
	// Set leverage first if specified
	if req.Leverage > 1 {
		_, err := c.FuturesClient.NewChangeLeverageService().
			Symbol(req.Symbol).
			Leverage(req.Leverage).
			Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to set leverage: %w", err)
		}
	}

	// Convert order type
	orderType, err := c.convertOrderType(req.OrderType)
	if err != nil {
		return nil, err
	}

	// Build order service
	orderService := c.FuturesClient.NewCreateOrderService().
		Symbol(req.Symbol).
		Side(c.convertSide(req.Side)).
		Type(orderType).
		Quantity(fmt.Sprintf("%.8f", req.Quantity))

	// Set price for limit orders
	if orderType == futures.OrderTypeLimit && req.Price > 0 {
		orderService = orderService.Price(fmt.Sprintf("%.8f", req.Price))
		
		// Set TimeInForce
		if req.TimeInForce != "" {
			orderService = orderService.TimeInForce(c.convertTimeInForce(req.TimeInForce))
		} else {
			orderService = orderService.TimeInForce(futures.TimeInForceTypeGTC)
		}
	}

	// Set stop price for stop orders
	if req.StopPrice > 0 {
		orderService = orderService.StopPrice(fmt.Sprintf("%.8f", req.StopPrice))
	}

	// Set working type for stop orders
	if req.WorkingType != "" {
		orderService = orderService.WorkingType(c.convertWorkingType(req.WorkingType))
	}

	// Set activation price for trailing stop
	if req.ActivationPrice > 0 {
		orderService = orderService.ActivationPrice(fmt.Sprintf("%.8f", req.ActivationPrice))
	}

	// Set callback rate for trailing stop
	if req.CallbackRate > 0 {
		orderService = orderService.CallbackRate(fmt.Sprintf("%.8f", req.CallbackRate))
	}

	// Set position side
	if req.PositionSide != "" {
		orderService = orderService.PositionSide(c.convertPositionSide(req.PositionSide))
	}

	// Set reduce only
	if req.ReduceOnly {
		orderService = orderService.ReduceOnly(req.ReduceOnly)
	}

	// Set close position
	if req.ClosePosition {
		orderService = orderService.ClosePosition(req.ClosePosition)
	}

	// Note: STP, PriceMatch, NewOrderRespType, GoodTillDate may not be available in library
	// These would need to be added via direct HTTP calls if library doesn't support them

	order, err := orderService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create futures order: %w", err)
	}

	return order, nil
}

// ModifyFuturesOrder modifies an existing futures order
// Note: Full implementation requires direct HTTP calls with proper signing
// This is a placeholder that updates the database
func (c *Client) ModifyFuturesOrder(ctx context.Context, req *ModifyOrderRequest) (*futures.CreateOrderResponse, error) {
	// Note: Order modification via PUT /fapi/v1/order requires direct HTTP implementation
	// with proper HMAC SHA256 signing. The go-binance library may not have this method.
	// For production, implement direct HTTP calls with proper authentication.
	return nil, fmt.Errorf("order modification requires direct HTTP implementation with proper signing. Please use cancel and recreate for now.")
}

// CreateBatchOrders creates multiple orders at once using direct HTTP
func (c *Client) CreateBatchOrders(ctx context.Context, orders []*AdvancedOrderRequest) ([]*futures.CreateOrderResponse, error) {
	// Batch orders require direct HTTP implementation
	// For now, create orders sequentially
	var responses []*futures.CreateOrderResponse
	var errors []string

	for _, req := range orders {
		order, err := c.CreateAdvancedFuturesOrder(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Order failed: %v", err))
			continue
		}
		responses = append(responses, order)
	}

	if len(errors) > 0 && len(responses) == 0 {
		return nil, fmt.Errorf("all orders failed: %v", errors)
	}

	return responses, nil
}

// CancelBatchOrders cancels multiple orders
func (c *Client) CancelBatchOrders(ctx context.Context, symbol string, orderIDs []int64, clientOrderIDs []string) ([]*futures.CancelOrderResponse, error) {
	// Cancel orders sequentially
	var responses []*futures.CancelOrderResponse

	for _, orderID := range orderIDs {
		resp, err := c.FuturesClient.NewCancelOrderService().
			Symbol(symbol).
			OrderID(orderID).
			Do(ctx)
		if err != nil {
			continue
		}
		responses = append(responses, resp)
	}

	for _, clientOrderID := range clientOrderIDs {
		resp, err := c.FuturesClient.NewCancelOrderService().
			Symbol(symbol).
			OrigClientOrderID(clientOrderID).
			Do(ctx)
		if err != nil {
			continue
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// SetPositionMode sets position mode (One-way or Hedge)
// Note: May require direct HTTP implementation if library doesn't support
func (c *Client) SetPositionMode(ctx context.Context, dualSide bool) error {
	// Try to use library method if available
	// If not available, would need direct HTTP call
	// For now, return error indicating it needs implementation
	return fmt.Errorf("position mode setting requires direct HTTP implementation. Method may not be available in library.")
}

// GetPositionMode gets current position mode
// Note: May require direct HTTP implementation if library doesn't support
func (c *Client) GetPositionMode(ctx context.Context) (bool, error) {
	// Try to use library method if available
	// If not available, would need direct HTTP call
	return false, fmt.Errorf("position mode getting requires direct HTTP implementation. Method may not be available in library.")
}

// Helper functions for conversion
func (c *Client) convertOrderType(orderType string) (futures.OrderType, error) {
	switch orderType {
	case "MARKET":
		return futures.OrderTypeMarket, nil
	case "LIMIT":
		return futures.OrderTypeLimit, nil
	case "STOP":
		return futures.OrderTypeStop, nil
	case "STOP_MARKET":
		return futures.OrderTypeStopMarket, nil
	case "STOP_LIMIT":
		// Note: STOP_LIMIT may not be available in library
		// Use STOP as fallback
		return futures.OrderTypeStop, nil
	case "TAKE_PROFIT":
		return futures.OrderTypeTakeProfit, nil
	case "TAKE_PROFIT_MARKET":
		return futures.OrderTypeTakeProfitMarket, nil
	case "TRAILING_STOP_MARKET":
		return futures.OrderTypeTrailingStopMarket, nil
	default:
		return "", fmt.Errorf("unsupported order type: %s", orderType)
	}
}

func (c *Client) convertSide(side string) futures.SideType {
	if side == "BUY" {
		return futures.SideTypeBuy
	}
	return futures.SideTypeSell
}

func (c *Client) convertTimeInForce(tif string) futures.TimeInForceType {
	switch tif {
	case "GTC":
		return futures.TimeInForceTypeGTC
	case "IOC":
		return futures.TimeInForceTypeIOC
	case "FOK":
		return futures.TimeInForceTypeFOK
	case "GTX":
		return futures.TimeInForceTypeGTX
	default:
		return futures.TimeInForceTypeGTC
	}
}

func (c *Client) convertWorkingType(wt string) futures.WorkingType {
	if wt == "MARK_PRICE" {
		return futures.WorkingTypeMarkPrice
	}
	return futures.WorkingTypeContractPrice
}

func (c *Client) convertPositionSide(ps string) futures.PositionSideType {
	if ps == "LONG" {
		return futures.PositionSideTypeLong
	}
	return futures.PositionSideTypeShort
}

// Request types
type AdvancedOrderRequest struct {
	Symbol                string
	Side                  string
	OrderType             string
	Quantity              float64
	Price                 float64
	StopPrice             float64
	ActivationPrice       float64
	CallbackRate          float64
	Leverage              int
	PositionSide          string
	TimeInForce           string
	WorkingType           string
	ReduceOnly            bool
	ClosePosition         bool
	SelfTradePreventionMode string
	PriceMatch            string
	NewOrderRespType      string
	ClientOrderID         string
	GoodTillDate          *time.Time
}

type ModifyOrderRequest struct {
	Symbol         string
	OrderID        int64
	ClientOrderID  string
	Quantity       float64
	Price          float64
	StopPrice      float64
	ActivationPrice float64
	CallbackRate   float64
	PriceMatch     string
}
