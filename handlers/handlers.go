package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"futures-options/services"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handlers struct {
	tradingService *services.TradingService
}

func NewHandlers(tradingService *services.TradingService) *Handlers {
	return &Handlers{
		tradingService: tradingService,
	}
}

// CreateFuturesOrder handles POST /api/futures/order
// @Summary      Create a futures order
// @Description  Create a new futures trading order on Binance
// @Tags         futures
// @Accept       json
// @Produce      json
// @Param        order  body      services.CreateFuturesOrderRequest  true  "Futures Order Request"
// @Success      200    {object}  models.FuturesOrder
// @Failure      400    {string}  string  "Bad Request"
// @Failure      500    {string}  string  "Internal Server Error"
// @Router       /api/futures/order [post]
func (h *Handlers) CreateFuturesOrder(w http.ResponseWriter, r *http.Request) {
	var req services.CreateFuturesOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.tradingService.CreateFuturesOrder(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CreateOptionsOrder handles POST /api/options/order
// @Summary      Create an options order
// @Description  Create a new options trading order
// @Tags         options
// @Accept       json
// @Produce      json
// @Param        order  body      services.CreateOptionsOrderRequest  true  "Options Order Request"
// @Success      200    {object}  models.OptionsOrder
// @Failure      400    {string}  string  "Bad Request"
// @Failure      500    {string}  string  "Internal Server Error"
// @Router       /api/options/order [post]
func (h *Handlers) CreateOptionsOrder(w http.ResponseWriter, r *http.Request) {
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

// GetFuturesOrders handles GET /api/futures/orders
// @Summary      Get futures orders
// @Description  Retrieve all futures orders, optionally filtered by symbol
// @Tags         futures
// @Produce      json
// @Param        symbol  query     string  false  "Filter by symbol (e.g., BTCUSDT)"
// @Success      200     {array}   models.FuturesOrder
// @Failure      500     {string}  string  "Internal Server Error"
// @Router       /api/futures/orders [get]
func (h *Handlers) GetFuturesOrders(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")

	orders, err := h.tradingService.GetFuturesOrders(r.Context(), symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetOptionsOrders handles GET /api/options/orders
// @Summary      Get options orders
// @Description  Retrieve all options orders, optionally filtered by symbol
// @Tags         options
// @Produce      json
// @Param        symbol  query     string  false  "Filter by symbol"
// @Success      200     {array}   models.OptionsOrder
// @Failure      500     {string}  string  "Internal Server Error"
// @Router       /api/options/orders [get]
func (h *Handlers) GetOptionsOrders(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")

	orders, err := h.tradingService.GetOptionsOrders(r.Context(), symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetPositions handles GET /api/positions
// @Summary      Get positions
// @Description  Retrieve all positions, optionally filtered by type (FUTURES or OPTIONS)
// @Tags         positions
// @Produce      json
// @Param        type  query     string  false  "Filter by position type (FUTURES or OPTIONS)"
// @Success      200   {array}   models.Position
// @Failure      500   {string}  string  "Internal Server Error"
// @Router       /api/positions [get]
func (h *Handlers) GetPositions(w http.ResponseWriter, r *http.Request) {
	positionType := r.URL.Query().Get("type")

	positions, err := h.tradingService.GetPositions(r.Context(), positionType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// SyncPositions handles POST /api/positions/sync
// @Summary      Sync positions from Binance
// @Description  Sync current positions from Binance to local database
// @Tags         positions
// @Produce      json
// @Success      200   {object}  map[string]string
// @Failure      500   {string}  string  "Internal Server Error"
// @Router       /api/positions/sync [post]
func (h *Handlers) SyncPositions(w http.ResponseWriter, r *http.Request) {
	err := h.tradingService.SyncPositionsFromBinance(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Positions synced successfully"})
}

// SaveAPICredentials handles POST /api/credentials
// @Summary      Save API credentials
// @Description  Save Binance API credentials to the database
// @Tags         credentials
// @Accept       json
// @Produce      json
// @Param        credentials  body      services.SaveAPICredentialsRequest  true  "API Credentials"
// @Success      200          {object}  models.APICredentials
// @Failure      400          {string}  string  "Bad Request"
// @Failure      500          {string}  string  "Internal Server Error"
// @Router       /api/credentials [post]
func (h *Handlers) SaveAPICredentials(w http.ResponseWriter, r *http.Request) {
	var req services.SaveAPICredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	credentials, err := h.tradingService.SaveAPICredentials(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credentials)
}

// GetAPICredentials handles GET /api/credentials
// @Summary      Get API credentials
// @Description  Retrieve stored API credentials, optionally filtered to active only
// @Tags         credentials
// @Produce      json
// @Param        active_only  query     bool    false  "Filter to active credentials only"
// @Success      200          {array}   models.APICredentials
// @Failure      500          {string}  string  "Internal Server Error"
// @Router       /api/credentials [get]
func (h *Handlers) GetAPICredentials(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	credentials, err := h.tradingService.GetAPICredentials(r.Context(), activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credentials)
}

// HealthCheck handles GET /health
// @Summary      Health check
// @Description  Check if the API server is running
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}

func SetupRoutes(h *Handlers) *mux.Router {
	router := mux.NewRouter()

	// Request logging middleware
	router.Use(loggingMiddleware)

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Futures routes
	futures := api.PathPrefix("/futures").Subrouter()
	futures.HandleFunc("/order", h.CreateFuturesOrder).Methods("POST")
	futures.HandleFunc("/orders", h.GetFuturesOrders).Methods("GET")

	// Options routes
	options := api.PathPrefix("/options").Subrouter()
	options.HandleFunc("/orders", h.GetOptionsOrders).Methods("GET")

	// Positions routes
	api.HandleFunc("/positions", h.GetPositions).Methods("GET")
	api.HandleFunc("/positions/sync", h.SyncPositions).Methods("POST")

	// API Credentials routes
	api.HandleFunc("/credentials", h.SaveAPICredentials).Methods("POST")
	api.HandleFunc("/credentials", h.GetAPICredentials).Methods("GET")

	// Advanced Futures routes
	api.HandleFunc("/futures/advanced/order", h.CreateAdvancedFuturesOrder).Methods("POST")
	api.HandleFunc("/futures/order/modify", h.ModifyFuturesOrder).Methods("PUT")
	api.HandleFunc("/futures/batch/orders", h.CreateBatchOrders).Methods("POST")
	api.HandleFunc("/futures/batch/orders/cancel", h.CancelBatchOrders).Methods("DELETE")
	api.HandleFunc("/futures/position-mode", h.SetPositionMode).Methods("POST")
	api.HandleFunc("/futures/position-mode", h.GetPositionMode).Methods("GET")
    api.HandleFunc("/futures/account/status", h.GetAccountStatusWS).Methods("GET")
    api.HandleFunc("/futures/account/balance", h.GetAccountBalanceWS).Methods("GET")

    // Key utilities
    api.HandleFunc("/keys/ed25519/generate", h.GenerateEd25519Key).Methods("POST")

	// WebSocket routes
	api.HandleFunc("/websocket/connect", h.ConnectWebSocket).Methods("GET")
	api.HandleFunc("/websocket/messages", h.GetWebSocketMessages).Methods("GET")

	// Options routes (fully implemented)
	options.HandleFunc("/order", h.CreateOptionsOrderAdvanced).Methods("POST")
	options.HandleFunc("/positions", h.GetOptionsPositions).Methods("GET")

	return router
}

// statusRecorder wraps http.ResponseWriter to capture status code and size
type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		// Status not set yet; default to 200 OK
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// loggingMiddleware logs each HTTP request with method, path, status and duration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		dur := time.Since(start)
		log.Printf("%s %s %d %dB %s", r.Method, r.URL.Path, rec.status, rec.size, dur)
	})
}

