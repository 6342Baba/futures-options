# Binance Futures & Options Trading System

A Go-based trading system for Binance Futures and Options with MongoDB integration, designed for testnet/demo account usage.

## Features

- **Futures Trading**: Create and manage futures orders on Binance testnet
- **Options Trading**: Support for options orders (structure ready for Binance Options API)
- **MongoDB Integration**: Persistent storage of orders and positions
- **REST API**: HTTP endpoints for trading operations
- **Position Sync**: Sync positions from Binance to local database

## Prerequisites

- Go 1.21 or higher
- MongoDB (local or remote instance)
- Binance testnet account and API keys

## Setup

### 1. Get Binance Testnet API Keys

1. Visit [Binance Testnet](https://testnet.binancefuture.com/)
2. Create a testnet account or log in
3. Generate API keys from the API management section

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Copy the example environment file and update with your credentials:

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```env
BINANCE_API_KEY=your_testnet_api_key
BINANCE_SECRET_KEY=your_testnet_secret_key
BINANCE_TESTNET=true
BINANCE_FUTURES_TESTNET_URL=https://testnet.binancefuture.com
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=futures_options_db
PORT=8080
```

### 4. Start MongoDB

Make sure MongoDB is running:

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or using local MongoDB installation
mongod
```

### 5. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080` (or your configured PORT).

## API Endpoints

### Health Check

```bash
GET /health
```

### Futures Orders

**Create Futures Order**
```bash
POST /api/futures/order
Content-Type: application/json

{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "order_type": "MARKET",
  "quantity": 0.001,
  "leverage": 10,
  "position_side": "LONG"
}
```

**Get Futures Orders**
```bash
GET /api/futures/orders?symbol=BTCUSDT
```

### Options Orders

**Create Options Order**
```bash
POST /api/options/order
Content-Type: application/json

{
  "symbol": "BTC-OPTIONS",
  "side": "BUY",
  "order_type": "MARKET",
  "quantity": 1,
  "strike_price": 50000,
  "expiry_date": "2024-12-31T00:00:00Z",
  "option_type": "CALL"
}
```

**Get Options Orders**
```bash
GET /api/options/orders?symbol=BTC-OPTIONS
```

### Positions

**Get Positions**
```bash
GET /api/positions?type=FUTURES
```

**Sync Positions from Binance**
```bash
POST /api/positions/sync
```

## Example Usage

### Create a Futures Market Order

```bash
curl -X POST http://localhost:8080/api/futures/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "order_type": "MARKET",
    "quantity": 0.001,
    "leverage": 10,
    "position_side": "LONG"
  }'
```

### Get All Futures Orders

```bash
curl http://localhost:8080/api/futures/orders
```

### Sync Positions from Binance

```bash
curl -X POST http://localhost:8080/api/positions/sync
```

## Project Structure

```
futures-options/
├── main.go                 # Application entry point
├── config/
│   └── config.go          # Configuration management
├── models/
│   └── models.go          # Data models
├── database/
│   └── mongodb.go         # MongoDB connection and operations
├── binance/
│   └── client.go          # Binance API client
├── services/
│   └── trading_service.go # Business logic
├── handlers/
│   └── handlers.go        # HTTP handlers
├── go.mod
├── go.sum
└── README.md
```

## Important Notes

1. **Testnet Only**: This application is configured for Binance testnet by default. Never use real API keys in testnet mode.

2. **Options Trading**: The Options API implementation is a placeholder. Binance Options API may have different endpoints and requirements. Please refer to the official Binance Options API documentation for complete implementation.

3. **Error Handling**: The application includes basic error handling. In production, you should add more comprehensive error handling and logging.

4. **Security**: Never commit your `.env` file with real API keys. Always use environment variables or secure secret management in production.

## Testing

You can test the API using tools like:
- `curl` (command line)
- Postman
- `httpie`
- Any HTTP client

## Troubleshooting

### MongoDB Connection Issues
- Ensure MongoDB is running: `mongosh` or check with `docker ps`
- Verify the connection string in `.env` matches your MongoDB setup

### Binance API Issues
- Verify your API keys are correct and have the necessary permissions
- Check that you're using testnet API keys on testnet URLs
- Ensure your testnet account has sufficient testnet funds

### Port Already in Use
- Change the `PORT` in `.env` to a different port
- Or stop the process using the port

## License

This project is for educational purposes. Use at your own risk.

