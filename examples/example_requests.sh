#!/bin/bash

# Example API requests for the Futures & Options Trading System
# Make sure the server is running on localhost:8080

BASE_URL="http://localhost:8080"

echo "=== Health Check ==="
curl -X GET "${BASE_URL}/health"
echo -e "\n\n"

echo "=== Create Futures Market Order (BUY) ==="
curl -X POST "${BASE_URL}/api/futures/order" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "order_type": "MARKET",
    "quantity": 0.001,
    "leverage": 10,
    "position_side": "LONG"
  }'
echo -e "\n\n"

echo "=== Create Futures Limit Order (SELL) ==="
curl -X POST "${BASE_URL}/api/futures/order" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "ETHUSDT",
    "side": "SELL",
    "order_type": "LIMIT",
    "quantity": 0.01,
    "price": 2500.0,
    "leverage": 5,
    "position_side": "SHORT"
  }'
echo -e "\n\n"

echo "=== Get All Futures Orders ==="
curl -X GET "${BASE_URL}/api/futures/orders"
echo -e "\n\n"

echo "=== Get Futures Orders for BTCUSDT ==="
curl -X GET "${BASE_URL}/api/futures/orders?symbol=BTCUSDT"
echo -e "\n\n"

echo "=== Create Options Order ==="
curl -X POST "${BASE_URL}/api/options/order" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC-OPTIONS",
    "side": "BUY",
    "order_type": "MARKET",
    "quantity": 1,
    "strike_price": 50000,
    "expiry_date": "2024-12-31T00:00:00Z",
    "option_type": "CALL"
  }'
echo -e "\n\n"

echo "=== Get All Options Orders ==="
curl -X GET "${BASE_URL}/api/options/orders"
echo -e "\n\n"

echo "=== Get All Positions ==="
curl -X GET "${BASE_URL}/api/positions"
echo -e "\n\n"

echo "=== Sync Positions from Binance ==="
curl -X POST "${BASE_URL}/api/positions/sync"
echo -e "\n\n"

