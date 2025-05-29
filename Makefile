.PHONY: build run test clean docker-build docker-run help

# Application name
APP_NAME=orangefeed

# Build the application
build:
	@echo "🔨 Building OrangeFeed..."
	go build -o bin/$(APP_NAME) cmd/orangefeed/main.go
	@echo "✅ Build complete: bin/$(APP_NAME)"

# Run the application
run: build
	@echo "🚀 Starting OrangeFeed..."
	./bin/$(APP_NAME)

# Run the test application
test:
	@echo "🧪 Running test application..."
	go run test_real_ai.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Run with Docker
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME) .

docker-run: docker-build
	@echo "🐳 Running with Docker..."
	docker-compose up

# Development setup
setup:
	@echo "⚙️ Setting up development environment..."
	cp config.env.example .env
	@echo "📝 Please edit .env file with your credentials"
	@echo "✅ Setup complete"

# Check code quality
lint:
	@echo "🔍 Running linter..."
	go vet ./...
	go fmt ./...

# Show help
help:
	@echo "🎯 OrangeFeed - Truth Social Market Intelligence Bot"
	@echo ""
	@echo "Available commands:"
	@echo "  build        Build the application"
	@echo "  run          Build and run the application"
	@echo "  test         Run the test application"
	@echo "  clean        Clean build artifacts"
	@echo "  deps         Install dependencies"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run   Run with Docker Compose"
	@echo "  setup        Setup development environment"
	@echo "  lint         Run code quality checks"
	@echo "  help         Show this help message"
	@echo ""
	@echo "📚 Documentation: README.md" 