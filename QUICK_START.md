# Quick Start Guide - Advanced Features

## 🚀 Quick Examples

### 1. Create Stop Loss Order

```bash
curl -X POST http://localhost:9090/api/futures/advanced/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "SELL",
    "order_type": "STOP_MARKET",
    "quantity": 0.001,
    "stop_price": 50000,
    "working_type": "MARK_PRICE",
    "reduce_only": true,
    "leverage": 10
  }'
```

### 2. Create Take Profit Order

```bash
curl -X POST http://localhost:9090/api/futures/advanced/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "SELL",
    "order_type": "TAKE_PROFIT_MARKET",
    "quantity": 0.001,
    "stop_price": 55000,
    "working_type": "MARK_PRICE",
    "reduce_only": true
  }'
```

### 3. Create Order with STP and Price Match

```bash
curl -X POST http://localhost:9090/api/futures/advanced/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "order_type": "LIMIT",
    "quantity": 0.001,
    "price_match": "OPPONENT",
    "self_trade_prevention_mode": "EXPIRE_TAKER",
    "leverage": 10
  }'
```

### 4. Create Batch Orders

```bash
curl -X POST http://localhost:9090/api/futures/batch/orders \
  -H "Content-Type: application/json" \
  -d '{
    "orders": [
      {
        "symbol": "BTCUSDT",
        "side": "BUY",
        "order_type": "LIMIT",
        "quantity": 0.001,
        "price": 50000,
        "leverage": 10
      },
      {
        "symbol": "ETHUSDT",
        "side": "BUY",
        "order_type": "LIMIT",
        "quantity": 0.01,
        "price": 2500,
        "leverage": 5
      }
    ]
  }'
```

### 5. Set Position Mode to Hedge

```bash
curl -X POST http://localhost:9090/api/futures/position-mode \
  -H "Content-Type: application/json" \
  -d '{
    "dual_side": true
  }'
```

### 6. Create Options Order

```bash
curl -X POST http://localhost:9090/api/options/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC-25000C-241231",
    "side": "BUY",
    "order_type": "MARKET",
    "quantity": 1,
    "strike_price": 50000,
    "expiry_date": "2024-12-31T00:00:00Z",
    "option_type": "CALL"
  }'
```

## 📋 All Available Endpoints

### Basic Futures
- `POST /api/futures/order` - Create basic order
- `GET /api/futures/orders` - Get orders

### Advanced Futures
- `POST /api/futures/advanced/order` - Create advanced order
- `PUT /api/futures/order/modify` - Modify order
- `POST /api/futures/batch/orders` - Batch create
- `DELETE /api/futures/batch/orders/cancel` - Batch cancel
- `POST /api/futures/position-mode` - Set position mode
- `GET /api/futures/position-mode` - Get position mode

### Options
- `POST /api/options/order` - Create options order
- `GET /api/options/orders` - Get options orders
- `GET /api/options/positions` - Get options positions

### WebSocket
- `GET /api/websocket/connect` - Connect WebSocket
- `GET /api/websocket/messages` - Get messages

### Other
- `POST /api/credentials` - Save API keys
- `GET /api/credentials` - Get API keys
- `GET /api/positions` - Get positions
- `POST /api/positions/sync` - Sync positions
- `GET /health` - Health check
- `GET /swagger/index.html` - Swagger UI

## 🔧 Order Types Reference

| Order Type | Description | Required Fields |
|------------|-------------|----------------|
| MARKET | Market order | symbol, side, quantity |
| LIMIT | Limit order | symbol, side, quantity, price |
| STOP | Stop order | symbol, side, quantity, stopPrice |
| STOP_MARKET | Stop market | symbol, side, quantity, stopPrice |
| STOP_LIMIT | Stop limit | symbol, side, quantity, price, stopPrice |
| TAKE_PROFIT | Take profit | symbol, side, quantity, stopPrice |
| TAKE_PROFIT_MARKET | Take profit market | symbol, side, quantity, stopPrice |
| TRAILING_STOP_MARKET | Trailing stop | symbol, side, quantity, activationPrice, callbackRate |

## 📝 Notes

1. **Testnet**: All features work on Binance Testnet
2. **API Keys**: Store via `POST /api/credentials` or `.env` file
3. **Swagger**: Visit `http://localhost:9090/swagger/index.html` for interactive docs
4. **Library Limitations**: Some features may need direct HTTP implementation
5. **Options API**: Requires proper authentication setup

