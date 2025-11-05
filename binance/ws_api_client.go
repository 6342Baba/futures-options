package binance

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"futures-options/config" // <-- change to your actual module path

	"github.com/gorilla/websocket"
)

// WSAPIClient is a minimal client for Binance Futures WebSocket API
type WSAPIClient struct {
    conn *websocket.Conn
    cfg  *config.Config
}

// NewWSAPIClient connects to the appropriate ws-fapi endpoint
func NewWSAPIClient(cfg *config.Config) (*WSAPIClient, error) {
    url := cfg.BinanceFuturesWSAPIURL
    if cfg.BinanceTestnet {
        url = cfg.BinanceFuturesWSAPIURLTest
    }

    c, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to WebSocket API: %w", err)
    }

    return &WSAPIClient{conn: c, cfg: cfg}, nil
}

// getServerTimeMs fetches Binance serverTime in ms to avoid client clock skew.
func getServerTimeMs(cfg *config.Config) int64 {
    base := "https://fapi.binance.com"
    if cfg.BinanceTestnet {
        // cfg.BinanceFuturesTestnetURL e.g. https://demo-fapi.binance.com
        base = cfg.BinanceFuturesTestnetURL
    }
    url := strings.TrimRight(base, "/") + "/fapi/v1/time"
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return time.Now().UnixMilli()
    }
    client := &http.Client{Timeout: 2 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return time.Now().UnixMilli()
    }
    defer resp.Body.Close()
    var body struct{ ServerTime int64 `json:"serverTime"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.ServerTime == 0 {
        return time.Now().UnixMilli()
    }
    return body.ServerTime
}

// Close closes the WebSocket connection
func (w *WSAPIClient) Close() error {
    if w.conn != nil {
        return w.conn.Close()
    }
    return nil
}

// WSRequest represents a generic WS API request
type WSRequest struct {
    ID     interface{}            `json:"id"`
    Method string                 `json:"method"`
    Params map[string]interface{} `json:"params,omitempty"`
}

// WSResponse represents a generic WS API response envelope
type WSResponse struct {
    ID     interface{} `json:"id"`
    Status int         `json:"status"`
    Result interface{} `json:"result,omitempty"`
    Error  *struct {
        Code int    `json:"code"`
        Msg  string `json:"msg"`
    } `json:"error,omitempty"`
}

//
// ---------- KEY RESOLUTION ----------
//

// resolvePrivateKey reads an Ed25519 private key from file (PEM or raw seed/key).
// If no path is provided, defaults to ./ed25519.key. Returns error if not found/invalid.
func resolvePrivateKey(cfg *config.Config) (ed25519.PrivateKey, error) {
    path := cfg.Ed25519PrivateKeyPath
    if strings.TrimSpace(path) == "" {
        path = "./ed25519.key"
    }
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("no Ed25519 key found at %s", path)
    }
    data = []byte(strings.TrimSpace(string(data)))

    if blk, _ := pem.Decode(data); blk != nil {
        keyAny, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
        if err == nil {
            if pk, ok := keyAny.(ed25519.PrivateKey); ok {
                return pk, nil
            }
        }
    }
    switch len(data) {
    case ed25519.SeedSize:
        return ed25519.NewKeyFromSeed(data), nil
    case ed25519.PrivateKeySize:
        return ed25519.PrivateKey(data), nil
    }
    return nil, errors.New("invalid Ed25519 key content (expect raw 32-byte seed, 64-byte key, or PKCS#8 PEM)")
}

//
// ---------- CORE SEND / READ ----------
//

// SendRequest sends an arbitrary WS API request and decodes the response into out (if non-nil)
func (w *WSAPIClient) SendRequest(ctx context.Context, id interface{}, method string, params map[string]interface{}, out interface{}) error {
    req := WSRequest{ID: id, Method: method, Params: params}

    if deadline, ok := ctx.Deadline(); ok {
        _ = w.conn.SetWriteDeadline(deadline)
    }
    if err := w.conn.WriteJSON(req); err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }

    if deadline, ok := ctx.Deadline(); ok {
        _ = w.conn.SetReadDeadline(deadline)
    }
    var resp WSResponse
    if err := w.conn.ReadJSON(&resp); err != nil {
        return fmt.Errorf("failed to read response: %w", err)
    }
    if resp.Status != 200 {
        b, _ := json.Marshal(resp)
        return fmt.Errorf("request failed: %s", string(b))
    }
    if out != nil && resp.Result != nil {
        b, _ := json.Marshal(resp.Result)
        if err := json.Unmarshal(b, out); err != nil {
            return fmt.Errorf("failed to decode result: %w", err)
        }
    }
    return nil
}


//
// ---------- SIGNING HELPERS ----------
//

// buildSignaturePayload builds a sorted key=value&... payload from params (skipping "signature")
func buildSignaturePayload(params map[string]interface{}) (string, error) {
    keys := make([]string, 0, len(params))
    for k := range params {
        if k == "signature" {
            continue
        }
        keys = append(keys, k)
    }
    sort.Strings(keys)

    var b strings.Builder
    for i, k := range keys {
        if i > 0 {
            b.WriteByte('&')
        }
        b.WriteString(k)
        b.WriteByte('=')

        v := params[k]
        switch vv := v.(type) {
        case string:
            b.WriteString(vv)
        case int:
            b.WriteString(strconv.FormatInt(int64(vv), 10))
        case int64:
            b.WriteString(strconv.FormatInt(vv, 10))
        case float64:
            // better to avoid floats; if present, stringify
            b.WriteString(strconv.FormatFloat(vv, 'f', -1, 64))
        case bool:
            if vv {
                b.WriteString("true")
            } else {
                b.WriteString("false")
            }
        default:
            b.WriteString(fmt.Sprintf("%v", vv))
        }
    }
    return b.String(), nil
}

// SendSignedRequest signs params with Ed25519 (base64) and sends the request.
// It injects apiKey and timestamp if not provided.
func (w *WSAPIClient) SendSignedRequest(ctx context.Context, id interface{}, method string, params map[string]interface{}, out interface{}) error {
    priv, err := resolvePrivateKey(w.cfg)
    if err != nil {
        return err
    }

    if params == nil {
        params = map[string]interface{}{}
    }
    // inject apiKey + timestamp
    if _, ok := params["apiKey"]; !ok {
        params["apiKey"] = w.cfg.BinanceAPIKey
    }
    if _, ok := params["timestamp"]; !ok {
        ts := getServerTimeMs(w.cfg)
        ts = (ts / 1000) * 1000
        params["timestamp"] = ts
    }
    // (optional but good) add recvWindow
    // if _, ok := params["recvWindow"]; !ok {
    //     params["recvWindow"] = 5000
    // }

    payload, err := buildSignaturePayload(params)
    log.Printf("Payload: %s", payload)
    if err != nil {
        return err
    }

    sig := ed25519.Sign(priv, []byte(payload))
    params["signature"] = base64.StdEncoding.EncodeToString(sig)

    return w.SendRequest(ctx, id, method, params, out)
}


