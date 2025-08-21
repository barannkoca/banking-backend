# Banking Backend Makefile

.PHONY: help run build clean test db-setup db-drop db-reset

# Default goal
.DEFAULT_GOAL := help

## Show this help message
help:
	@echo "🏦 Banking Backend Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m%-15s\033[0m %s\n", "Target", "Description"} /^[a-zA-Z_-]+:.*?##/ { printf "\033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

## Run the server in development mode
run:
	@echo "🚀 Starting Banking Backend Server..."
	go run cmd/server/main.go

## Build the application
build:
	@echo "🔨 Building Banking Backend..."
	go build -o bin/banking-backend cmd/server/main.go
	@echo "✅ Build complete: bin/banking-backend"

## Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

## Format code
fmt:
	@echo "📝 Formatting code..."
	go fmt ./...

## Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	go mod tidy

##@ Utilities

## Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	go clean

## Stop all running instances
stop:
	@echo "🛑 Stopping all banking-backend processes..."
	-pkill -f "banking-backend"
	-pkill -f "go run cmd/server/main.go"

##@ Database

## Setup database
db-setup:
	@echo "🗄️ Setting up database..."
	createdb banking_db || true
	@echo "✅ Database created. Tables will be auto-created on first run."

## Drop database
db-drop:
	@echo "🗑️ Dropping database..."
	dropdb banking_db || true
	@echo "✅ Database dropped"

## Reset database (drop and recreate)
db-reset: db-drop db-setup
	@echo "🔄 Database reset complete"

