package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"futures-options/services"

	"github.com/gorilla/mux"
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
func (h *Handlers) SyncPositions(w http.ResponseWriter, r *http.Request) {
	err := h.tradingService.SyncPositionsFromBinance(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Positions synced successfully"})
}

// HealthCheck handles GET /health
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}

func SetupRoutes(h *Handlers) *mux.Router {
	router := mux.NewRouter()

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
	options.HandleFunc("/order", h.CreateOptionsOrder).Methods("POST")
	options.HandleFunc("/orders", h.GetOptionsOrders).Methods("GET")

	// Positions routes
	api.HandleFunc("/positions", h.GetPositions).Methods("GET")
	api.HandleFunc("/positions/sync", h.SyncPositions).Methods("POST")

	return router
}

