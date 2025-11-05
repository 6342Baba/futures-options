.PHONY: dev build run install clean swagger

GOPATH := $(shell go env GOPATH)
AIR := $(GOPATH)/bin/air

# Install dependencies for development
install:
	go mod download
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Development mode with auto-reload
dev:
	@if [ ! -f "$(AIR)" ]; then \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	$(AIR)

# Build the application
build:
	go build -o bin/futures-options .

# Run the application
run:
	go run main.go

# Generate Swagger documentation
swagger:
	@if ! command -v swag > /dev/null; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g main.go -o ./docs

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf docs/

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

