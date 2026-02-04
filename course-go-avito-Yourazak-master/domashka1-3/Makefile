BINARY_NAME=courier-service
GO=go

.PHONY: all build clean test lint run help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 GOOS=linux $(GO) build -ldflags="-w -s" -o $(BINARY_NAME) ./cmd
	@echo "✅ Build completed: $(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@$(GO) clean
	@echo "✅ Clean completed"

test:
	@echo "Running tests..."
	@$(GO) test ./... -v

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

run:
	@echo "Starting service..."
	@$(GO) run ./cmd

format:
	@echo "Formatting code..."
	@gofmt -w .
	@goimports -w .
	@echo "✅ Formatting completed"

deps:
	@echo "Installing dependencies..."
	@$(GO) mod download
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2
	@echo "✅ Dependencies installed"

help:
	@echo "Available commands:"
	@echo "  make build   - Build the binary"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make test    - Run tests"
	@echo "  make lint    - Run linter"
	@echo "  make run     - Run the application"
	@echo "  make format  - Format code"
	@echo "  make deps    - Install dependencies"