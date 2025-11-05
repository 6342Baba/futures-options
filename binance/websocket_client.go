package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"futures-options/config"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/gorilla/websocket"
)

// WebSocketClient handles WebSocket connections for real-time updates
type WebSocketClient struct {
	conn        *websocket.Conn
	client      *futures.Client
	config      *config.Config
	listenKey   string
	stopChan    chan struct{}
	messageChan chan *futures.WsUserDataEvent
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(client *futures.Client, cfg *config.Config) (*WebSocketClient, error) {
	ws := &WebSocketClient{
		client:      client,
		config:      cfg,
		stopChan:    make(chan struct{}),
		messageChan: make(chan *futures.WsUserDataEvent, 100),
	}

	// Get listen key
	service := client.NewStartUserStreamService()
	listenKey, err := service.Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get listen key: %w", err)
	}
	ws.listenKey = listenKey

	return ws, nil
}

// Connect connects to WebSocket and starts listening
func (ws *WebSocketClient) Connect(ctx context.Context) error {
	url := "wss://fstream.binance.com/ws/"
	if ws.config.BinanceTestnet {
		url = "wss://fstream.binancefuture.com/ws/"
	}
	url += ws.listenKey

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	ws.conn = conn

	// Start ping/pong
	go ws.keepAlive(ctx)

	// Start reading messages
	go ws.readMessages()

	return nil
}

// keepAlive sends ping to keep connection alive
func (ws *WebSocketClient) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.stopChan:
			return
		case <-ticker.C:
			// Ping listen key
			err := ws.client.NewKeepaliveUserStreamService().
				ListenKey(ws.listenKey).
				Do(ctx)
			if err != nil {
				log.Printf("Failed to keep alive: %v", err)
			}
		}
	}
}

// readMessages reads messages from WebSocket
func (ws *WebSocketClient) readMessages() {
	defer ws.conn.Close()

	for {
		select {
		case <-ws.stopChan:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			var event futures.WsUserDataEvent
			if err := json.Unmarshal(message, &event); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			select {
			case ws.messageChan <- &event:
			default:
				log.Println("Message channel full, dropping message")
			}
		}
	}
}

// GetMessageChannel returns the message channel
func (ws *WebSocketClient) GetMessageChannel() <-chan *futures.WsUserDataEvent {
	return ws.messageChan
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() error {
	close(ws.stopChan)
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

