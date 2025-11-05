package binance

import (
	"context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"futures-options/config"
)

// OptionsClient handles Binance Options API calls
// Note: Binance Options uses different endpoints (/eapi/v1/*)
type OptionsClient struct {
	config     *config.Config
	httpClient *http.Client
    apiKey     string
    secretKey  string
}

// NewOptionsClient creates a new Options client
func NewOptionsClient(cfg *config.Config) *OptionsClient {
	if cfg == nil {
		// Create default config for testnet
		cfg = &config.Config{
			BinanceTestnet: true,
		}
	}
	return &OptionsClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
        apiKey:     cfg.BinanceAPIKey,
        secretKey:  cfg.BinanceSecretKey,
	}
}

func (oc *OptionsClient) signParams(params url.Values) (string, error) {
    if oc.secretKey == "" || oc.apiKey == "" {
        return "", fmt.Errorf("options API keys not configured")
    }
    mac := hmac.New(sha256.New, []byte(oc.secretKey))
    mac.Write([]byte(params.Encode()))
    return hex.EncodeToString(mac.Sum(nil)), nil
}

// CreateOptionsOrder creates an options order
func (oc *OptionsClient) CreateOptionsOrder(ctx context.Context, req *OptionsOrderRequest) (*OptionsOrderResponse, error) {
	baseURL := "https://eapi.binance.com"
	if oc.config.BinanceTestnet {
        return nil, fmt.Errorf("Binance Options testnet is not available. Use mainnet for Options endpoints")
	}

	endpoint := baseURL + "/eapi/v1/order"

	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", req.Side)
	params.Set("type", req.OrderType)
	params.Set("quantity", strconv.FormatFloat(req.Quantity, 'f', -1, 64))

	if req.Price > 0 {
		params.Set("price", strconv.FormatFloat(req.Price, 'f', -1, 64))
	}

	if req.TimeInForce != "" {
		params.Set("timeInForce", req.TimeInForce)
	}

    // Signed parameters
    params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
    sig, err := oc.signParams(params)
	if err != nil {
        return nil, fmt.Errorf("signing failed: %w", err)
	}
    params.Set("signature", sig)

    reqURL := endpoint + "?" + params.Encode()
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to build request: %w", err)
    }
    httpReq.Header.Set("X-MBX-APIKEY", oc.apiKey)
    resp, err := oc.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to create options order: %w", err)
    }
    defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("options order failed with status: %d", resp.StatusCode)
	}

	var result OptionsOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetOptionsPositions gets current options positions
func (oc *OptionsClient) GetOptionsPositions(ctx context.Context) ([]*OptionsPosition, error) {
	baseURL := "https://eapi.binance.com"
	if oc.config.BinanceTestnet {
        return nil, fmt.Errorf("Binance Options testnet is not available. Use mainnet for Options endpoints")
	}

	endpoint := baseURL + "/eapi/v1/account"

    params := url.Values{}
    params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
    sig, err := oc.signParams(params)
    if err != nil {
        return nil, fmt.Errorf("signing failed: %w", err)
    }
    params.Set("signature", sig)

    reqURL := endpoint + "?" + params.Encode()
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to build request: %w", err)
    }
    httpReq.Header.Set("X-MBX-APIKEY", oc.apiKey)
    resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get options positions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get positions with status: %d", resp.StatusCode)
	}

	var account struct {
		Positions []*OptionsPosition `json:"positions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return account.Positions, nil
}

// OptionsOrderRequest represents an options order request
type OptionsOrderRequest struct {
	Symbol      string
	Side        string
	OrderType   string
	Quantity    float64
	Price       float64
	TimeInForce string
}

// OptionsOrderResponse represents an options order response
type OptionsOrderResponse struct {
	OrderID    int64  `json:"orderId"`
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	Side       string `json:"side"`
	Type       string `json:"type"`
	Quantity   string `json:"quantity"`
	Price      string `json:"price"`
	CreateTime int64  `json:"createTime"`
}

// OptionsPosition represents an options position
type OptionsPosition struct {
	Symbol        string  `json:"symbol"`
	Position      float64 `json:"position"`
	EntryPrice    float64 `json:"entryPrice"`
	MarkPrice     float64 `json:"markPrice"`
	UnrealizedPnl float64 `json:"unrealizedPnl"`
}

