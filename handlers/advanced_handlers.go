package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	"futures-options/services"
)

// CreateAdvancedFuturesOrder handles POST /api/futures/advanced/order
// @Summary      Create advanced futures order
// @Description  Create a futures order with advanced features (STOP, TAKE_PROFIT, TRAILING_STOP, STP, PriceMatch, etc.)
// @Tags         futures
// @Accept       json
// @Produce      json
// @Param        order  body      services.AdvancedOrderRequest  true  "Advanced Futures Order Request"
// @Success      200    {object}  models.FuturesOrder
// @Failure      400    {string}  string  "Bad Request"
// @Failure      500    {string}  string  "Internal Server Error"
// @Router       /api/futures/advanced/order [post]
func (h *Handlers) CreateAdvancedFuturesOrder(w http.ResponseWriter, r *http.Request) {
	var req services.AdvancedOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.tradingService.CreateAdvancedFuturesOrder(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ModifyFuturesOrder handles PUT /api/futures/order/modify
// @Summary      Modify futures order
// @Description  Modify an existing futures order (price, quantity, stop price, etc.)
// @Tags         futures
// @Accept       json
// @Produce      json
// @Param        order  body      services.ModifyOrderRequest  true  "Modify Order Request"
// @Success      200    {object}  models.FuturesOrder
// @Failure      400    {string}  string  "Bad Request"
// @Failure      500    {string}  string  "Internal Server Error"
// @Router       /api/futures/order/modify [put]
func (h *Handlers) ModifyFuturesOrder(w http.ResponseWriter, r *http.Request) {
	var req services.ModifyOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.tradingService.ModifyFuturesOrder(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CreateBatchOrders handles POST /api/futures/batch/orders
// @Summary      Create batch orders
// @Description  Create multiple futures orders at once
// @Tags         futures
// @Accept       json
// @Produce      json
// @Param        orders  body      services.BatchOrderRequest  true  "Batch Orders Request"
// @Success      200     {object}  services.BatchOrderResponse
// @Failure      400     {string}  string  "Bad Request"
// @Failure      500     {string}  string  "Internal Server Error"
// @Router       /api/futures/batch/orders [post]
func (h *Handlers) CreateBatchOrders(w http.ResponseWriter, r *http.Request) {
	var req services.BatchOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.tradingService.CreateBatchOrders(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelBatchOrders handles DELETE /api/futures/batch/orders/cancel
// @Summary      Cancel batch orders
// @Description  Cancel multiple futures orders at once
// @Tags         futures
// @Accept       json
// @Produce      json
// @Param        symbol          query     string   true  "Trading symbol"
// @Param        order_ids       query     []int64  false "Order IDs to cancel"
// @Param        client_order_ids query     []string false "Client Order IDs to cancel"
// @Success      200  {object}  map[string]string
// @Failure      400  {string}  string  "Bad Request"
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/futures/batch/orders/cancel [delete]
func (h *Handlers) CancelBatchOrders(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol parameter is required", http.StatusBadRequest)
		return
	}

	// Parse order IDs from query (simplified - would need proper parsing)
	err := h.tradingService.CancelBatchOrders(r.Context(), symbol, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Orders cancelled successfully"})
}

// SetPositionMode handles POST /api/futures/position-mode
// @Summary      Set position mode
// @Description  Switch between One-way and Hedge position mode
// @Tags         futures
// @Accept       json
// @Produce      json
// @Param        mode  body      map[string]bool  true  "Position mode: {\"dual_side\": true} for Hedge, false for One-way"
// @Success      200   {object}  map[string]string
// @Failure      400   {string}  string  "Bad Request"
// @Failure      500   {string}  string  "Internal Server Error"
// @Router       /api/futures/position-mode [post]
func (h *Handlers) SetPositionMode(w http.ResponseWriter, r *http.Request) {
	var req map[string]bool
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dualSide, ok := req["dual_side"]
	if !ok {
		http.Error(w, "dual_side parameter is required", http.StatusBadRequest)
		return
	}

	err := h.tradingService.SetPositionMode(r.Context(), dualSide)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Position mode updated successfully"})
}

// GetPositionMode handles GET /api/futures/position-mode
// @Summary      Get position mode
// @Description  Get current position mode (One-way or Hedge)
// @Tags         futures
// @Produce      json
// @Success      200  {object}  models.PositionModeConfig
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/futures/position-mode [get]
func (h *Handlers) GetPositionMode(w http.ResponseWriter, r *http.Request) {
	mode, err := h.tradingService.GetPositionMode(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mode)
}

// ConnectWebSocket handles GET /api/websocket/connect
// @Summary      Connect WebSocket
// @Description  Connect to Binance WebSocket for real-time updates
// @Tags         websocket
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/websocket/connect [get]
func (h *Handlers) ConnectWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket upgrade would be handled here
	// For now, return a message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "WebSocket connection initiated. Use WebSocket client library for full functionality.",
		"note":    "Full WebSocket implementation requires WebSocket upgrade handler",
	})
}

// GetWebSocketMessages handles GET /api/websocket/messages
// @Summary      Get WebSocket messages
// @Description  Get recent WebSocket messages (SSE or polling)
// @Tags         websocket
// @Produce      json
// @Success      200  {array}  models.WebSocketMessage
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/websocket/messages [get]
func (h *Handlers) GetWebSocketMessages(w http.ResponseWriter, r *http.Request) {
	// Placeholder - would need WebSocket message storage
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// GetAccountStatusWS handles GET /api/futures/account/status (WS API)
// @Summary      Get account status via WebSocket API
// @Tags         futures
// @Produce      json
// @Success      200  {object}  interface{}
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/futures/account/status [get]
func (h *Handlers) GetAccountStatusWS(w http.ResponseWriter, r *http.Request) {
    result, err := h.tradingService.GetAccountStatusWS(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// GetAccountBalanceWS handles GET /api/futures/account/balance (WS API)
// @Summary      Get account balance via WebSocket API
// @Tags         futures
// @Produce      json
// @Success      200  {object}  interface{}
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/futures/account/balance [get]
func (h *Handlers) GetAccountBalanceWS(w http.ResponseWriter, r *http.Request) {
    result, err := h.tradingService.GetAccountBalanceWS(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// CreateOptionsOrderAdvanced handles POST /api/options/order (fully implemented)
// @Summary      Create options order
// @Description  Create an options trading order (fully implemented)
// @Tags         options
// @Accept       json
// @Produce      json
// @Param        order  body      services.CreateOptionsOrderRequest  true  "Options Order Request"
// @Success      200    {object}  models.OptionsOrder
// @Failure      400    {string}  string  "Bad Request"
// @Failure      500    {string}  string  "Internal Server Error"
// @Router       /api/options/order [post]
func (h *Handlers) CreateOptionsOrderAdvanced(w http.ResponseWriter, r *http.Request) {
	var req services.CreateOptionsOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.tradingService.CreateOptionsOrder(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetOptionsPositions handles GET /api/options/positions
// @Summary      Get options positions
// @Description  Get current options positions
// @Tags         options
// @Produce      json
// @Success      200  {array}  models.Position
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/options/positions [get]
func (h *Handlers) GetOptionsPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.tradingService.GetOptionsPositions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// GenerateEd25519Key handles POST /api/keys/ed25519/generate
// @Summary      Generate Ed25519 keypair (seed + public)
// @Description  Generates a 32-byte Ed25519 private seed, writes it to ed25519.key, and returns seed/public in HEX and Base64
// @Tags         keys
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /api/keys/ed25519/generate [post]
func (h *Handlers) GenerateEd25519Key(w http.ResponseWriter, r *http.Request) {
    // Generate Ed25519 keypair
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        http.Error(w, "failed to generate key", http.StatusInternalServerError)
        return
    }

    // Extract 32-byte seed from 64-byte private key
    seed := priv.Seed()

    // Write seed to file in project root
    filePath := "ed25519.key"
    if err := os.WriteFile(filePath, seed, 0600); err != nil {
        http.Error(w, "failed to write key file", http.StatusInternalServerError)
        return
    }

    resp := map[string]string{
        "filePath":          filePath,
        "privateSeedHEX":    hex.EncodeToString(seed),
        "privateSeedB64":    base64.StdEncoding.EncodeToString(seed),
        "publicKeyHEX":      hex.EncodeToString(pub),
        "publicKeyB64":      base64.StdEncoding.EncodeToString(pub),
        // "note":              "Register publicKeyHEX/B64 with Binance WS-API; keep private seed secret",
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

