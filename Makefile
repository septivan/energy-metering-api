.PHONY: help build run test clean docker-build docker-run deploy-fly

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary
	@echo "Building server..."
	go build -o bin/server.exe ./cmd/server

run: ## Run the server locally
	@echo "Running server..."
	go run ./cmd/server

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf tmp/
	go clean

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t energy-metering-api:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env energy-metering-api:latest

deploy-fly: ## Deploy to fly.io
	@echo "Deploying to fly.io..."
	flyctl deploy

fly-logs: ## View fly.io logs
	@echo "Viewing logs..."
	flyctl logs

fly-status: ## Check fly.io app status
	flyctl status

fly-open: ## Open app in browser
	flyctl apps open

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

dev: ## Run in development mode with air (requires air)
	air
