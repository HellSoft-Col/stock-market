# Stock Market Trading Platform - Makefile

# Variables
BINARY_NAME=stock-market-server
AUTOMATED_CLIENT=automated-client
DOCKER_IMAGE=stock-market
DOCKER_TAG=latest
GO_FILES=$(shell find . -name '*.go' -type f)

# Build information
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_COMMIT_SHORT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%d %H:%M:%S UTC")

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: help
help: ## Show this help message
	@echo '$(COLOR_BOLD)Stock Market Trading Platform - Available Commands:$(COLOR_RESET)'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ''

##@ Building

.PHONY: build
build: ## Build the server binary
	@echo "$(COLOR_GREEN)Building server...$(COLOR_RESET)"
	@mkdir -p bin
	@go build -ldflags="-s -w -X 'main.Version=$(GIT_COMMIT_SHORT)' -X 'main.BuildDate=$(BUILD_DATE)'" -o bin/$(BINARY_NAME) ./cmd/server
	@echo "$(COLOR_GREEN)✓ Server built: bin/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-all
build-all: ## Build all binaries (server + clients)
	@echo "$(COLOR_GREEN)Building all binaries...$(COLOR_RESET)"
	@mkdir -p bin
	@go build -ldflags="-s -w -X 'main.Version=$(GIT_COMMIT_SHORT)' -X 'main.BuildDate=$(BUILD_DATE)'" -o bin/$(BINARY_NAME) ./cmd/server
	@go build -ldflags="-s -w" -o bin/$(AUTOMATED_CLIENT) ./cmd/automated-client
	@go build -ldflags="-s -w" -o bin/trading-cli ./cmd/trading-cli
	@echo "$(COLOR_GREEN)✓ All binaries built in bin/$(COLOR_RESET)"

.PHONY: build-linux
build-linux: ## Build Linux binary for deployment
	@echo "$(COLOR_GREEN)Building Linux binary...$(COLOR_RESET)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'main.Version=$(GIT_COMMIT_SHORT)' -X 'main.BuildDate=$(BUILD_DATE)'" -o $(BINARY_NAME)-linux ./cmd/server
	@echo "$(COLOR_GREEN)✓ Linux binary built: $(BINARY_NAME)-linux$(COLOR_RESET)"

##@ Development

.PHONY: run
run: ## Run the server locally
	@echo "$(COLOR_BLUE)Starting server...$(COLOR_RESET)"
	@go run ./cmd/server

.PHONY: dev
dev: fmt lint test build ## Run full development cycle (format, lint, test, build)
	@echo "$(COLOR_GREEN)✓ Development cycle complete$(COLOR_RESET)"

##@ Testing

.PHONY: test
test: ## Run all tests
	@echo "$(COLOR_YELLOW)Running tests...$(COLOR_RESET)"
	@go test -v ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_YELLOW)Running tests with coverage...$(COLOR_RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(COLOR_YELLOW)Running tests with race detector...$(COLOR_RESET)"
	@go test -race -v ./...

##@ Code Quality

.PHONY: fmt
fmt: ## Format code with gofumpt
	@echo "$(COLOR_YELLOW)Formatting code...$(COLOR_RESET)"
	@gofumpt -l -w .
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: lint
lint: ## Run linter (golangci-lint)
	@echo "$(COLOR_YELLOW)Running linter...$(COLOR_RESET)"
	@golangci-lint run ./...
	@echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_YELLOW)Running go vet...$(COLOR_RESET)"
	@go vet ./...
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

.PHONY: check
check: fmt lint vet test ## Run all code quality checks
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_YELLOW)Downloading dependencies...$(COLOR_RESET)"
	@go mod download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(COLOR_YELLOW)Updating dependencies...$(COLOR_RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

.PHONY: deps-tidy
deps-tidy: ## Tidy dependencies
	@echo "$(COLOR_YELLOW)Tidying dependencies...$(COLOR_RESET)"
	@go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies tidied$(COLOR_RESET)"

##@ Docker

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	@docker build --build-arg GIT_COMMIT=$(GIT_COMMIT) --build-arg BUILD_DATE="$(BUILD_DATE)" -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(COLOR_GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	@docker run -p 8080:8080 -p 8081:8081 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "$(COLOR_BLUE)Pushing Docker image...$(COLOR_RESET)"
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

##@ Database

.PHONY: seed
seed: ## Seed database with initial data
	@echo "$(COLOR_YELLOW)Seeding database...$(COLOR_RESET)"
	@go run ./cmd/seed
	@echo "$(COLOR_GREEN)✓ Database seeded$(COLOR_RESET)"

.PHONY: seed-teams
seed-teams: ## Seed database with team data
	@echo "$(COLOR_YELLOW)Seeding teams...$(COLOR_RESET)"
	@go run ./cmd/seed-teams
	@echo "$(COLOR_GREEN)✓ Teams seeded$(COLOR_RESET)"

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-linux $(AUTOMATED_CLIENT)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

.PHONY: clean-all
clean-all: clean ## Clean all generated files (including Docker)
	@echo "$(COLOR_YELLOW)Cleaning Docker images...$(COLOR_RESET)"
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ All cleaned$(COLOR_RESET)"

##@ Information

.PHONY: version
version: ## Show version information
	@echo "$(COLOR_BOLD)Version Information:$(COLOR_RESET)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Short Hash: $(GIT_COMMIT_SHORT)"
	@echo "  Build Date: $(BUILD_DATE)"

.PHONY: info
info: version ## Show project information
	@echo ""
	@echo "$(COLOR_BOLD)Project Information:$(COLOR_RESET)"
	@echo "  Binary:     $(BINARY_NAME)"
	@echo "  Go Version: $$(go version)"
	@echo "  Go Files:   $$(find . -name '*.go' -type f | wc -l | tr -d ' ')"

##@ Production

.PHONY: prod-test
prod-test: ## Run production tests
	@echo "$(COLOR_YELLOW)Running production tests...$(COLOR_RESET)"
	@./test-production.sh

.PHONY: deploy-check
deploy-check: lint test build-linux ## Pre-deployment checks
	@echo "$(COLOR_GREEN)✓ Deployment checks passed$(COLOR_RESET)"

# Default target
.DEFAULT_GOAL := help
