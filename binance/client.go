package binance

import (
	"context"
	"fmt"
	"time"

	"futures-options/config"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

type Client struct {
	FuturesClient *futures.Client
	OptionsClient *binance.Client
	Config        *config.Config
}

func NewClient(cfg *config.Config) *Client {
	client := &Client{
		Config: cfg,
	}

	// Initialize Futures Client (Testnet)
	if cfg.BinanceTestnet {
		client.FuturesClient = futures.NewClient("", "")
		client.FuturesClient.BaseURL = cfg.BinanceFuturesTestnetURL
	} else {
		client.FuturesClient = futures.NewClient(cfg.BinanceAPIKey, cfg.BinanceSecretKey)
	}

	// Note: Binance Options API might need different initialization
	// For now, using standard client for options
	if cfg.BinanceTestnet {
		client.OptionsClient = binance.NewClient("", "")
		// Options testnet URL might be different
	} else {
		client.OptionsClient = binance.NewClient(cfg.BinanceAPIKey, cfg.BinanceSecretKey)
	}

	return client
}

// SetAPIKeys sets the API keys for authenticated requests
func (c *Client) SetAPIKeys(apiKey, secretKey string) {
	c.FuturesClient = futures.NewClient(apiKey, secretKey)
	if c.Config.BinanceTestnet {
		c.FuturesClient.BaseURL = c.Config.BinanceFuturesTestnetURL
	}
}

// CreateFuturesOrder creates a futures order on Binance
func (c *Client) CreateFuturesOrder(ctx context.Context, symbol string, side futures.SideType, orderType futures.OrderType, quantity, price float64, leverage int) (*futures.CreateOrderResponse, error) {
	// Set leverage first
	if leverage > 1 {
		_, err := c.FuturesClient.NewChangeLeverageService().
			Symbol(symbol).
			Leverage(leverage).
			Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to set leverage: %w", err)
		}
	}

	// Create order
	orderService := c.FuturesClient.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(orderType).
		Quantity(fmt.Sprintf("%.8f", quantity))

	if orderType == futures.OrderTypeLimit {
		orderService = orderService.Price(fmt.Sprintf("%.8f", price)).TimeInForce(futures.TimeInForceTypeGTC)
	}

	order, err := orderService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create futures order: %w", err)
	}

	return order, nil
}

// GetFuturesAccount gets futures account information
func (c *Client) GetFuturesAccount(ctx context.Context) (*futures.Account, error) {
	account, err := c.FuturesClient.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures account: %w", err)
	}
	return account, nil
}

// GetFuturesPositions gets current futures positions
func (c *Client) GetFuturesPositions(ctx context.Context) ([]*futures.PositionRisk, error) {
	positions, err := c.FuturesClient.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures positions: %w", err)
	}
	return positions, nil
}

// CloseFuturesPosition closes a futures position
func (c *Client) CloseFuturesPosition(ctx context.Context, symbol string, side futures.SideType, quantity float64) (*futures.CreateOrderResponse, error) {
	// Close position by placing opposite order
	oppositeSide := futures.SideTypeBuy
	if side == futures.SideTypeBuy {
		oppositeSide = futures.SideTypeSell
	}

	order, err := c.FuturesClient.NewCreateOrderService().
		Symbol(symbol).
		Side(oppositeSide).
		Type(futures.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		ReduceOnly(true).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close futures position: %w", err)
	}

	return order, nil
}

// Note: Binance Options API implementation
// Options trading might require different API endpoints
// This is a placeholder structure - you may need to implement
// options-specific API calls based on Binance Options documentation

// CreateOptionsOrder creates an options order (placeholder - needs actual Options API)
func (c *Client) CreateOptionsOrder(ctx context.Context, symbol string, side string, orderType string, quantity, price, strikePrice float64, expiryDate time.Time, optionType string) (interface{}, error) {
	// This is a placeholder - Binance Options API might have different structure
	// You'll need to implement this based on actual Binance Options API documentation
	return nil, fmt.Errorf("options trading not yet fully implemented - please check Binance Options API documentation")
}

// GetOptionsPositions gets current options positions (placeholder)
func (c *Client) GetOptionsPositions(ctx context.Context) (interface{}, error) {
	// Placeholder for options positions
	return nil, fmt.Errorf("options positions not yet fully implemented")
}

