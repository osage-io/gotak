# GoTAK Server Makefile

.PHONY: dev dev-up dev-down dev-logs dev-status dev-clean dev-debug build run clean test test-integration test-scripts lint security security-audit security-audit-quick security-audit-headers security-audit-tls security-audit-auth docker precommit-install precommit-run precommit-update help

# Variables
BINARY_NAME=gotak-server
BINARY_PATH=./cmd/gotak-server
BUILD_DIR=./bin
VERSION=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker Hub settings (override with DOCKERHUB_USERNAME=youruser make docker-push)
DOCKERHUB_USERNAME ?= dfedick
SERVER_IMAGE_NAME=$(DOCKERHUB_USERNAME)/gotak-server
WEB_IMAGE_NAME=$(DOCKERHUB_USERNAME)/gotak-web

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_TIME) -X main.commit=$(GIT_COMMIT)"

# Development environment
dev-up: ## Start development environment services
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up -d --wait
	@echo "Development environment started. Services available at:"
	@echo "  PostgreSQL: localhost:5432 (user: gotak, db: gotak_dev)"
	@echo "  PostgreSQL Test: localhost:5433 (user: gotak, db: gotak_test)"
	@echo "  Redis: localhost:6379 (password: dev_redis_pass)"
	@echo "  NATS: localhost:4222 (monitoring: localhost:8222)"
	@echo "  Vault: localhost:8200 (running on host system)"
	@echo "  Jaeger UI: localhost:16686"
	@echo "  Adminer: localhost:8081 (database admin)"

dev-down: ## Stop development environment services
	@echo "Stopping development environment..."
	docker-compose -f docker-compose.dev.yml down

dev-logs: ## Show logs from development services
	docker-compose -f docker-compose.dev.yml logs -f

dev-status: ## Check status of development services
	docker-compose -f docker-compose.dev.yml ps

dev-clean: ## Stop services and clean volumes
	@echo "Cleaning development environment..."
	docker-compose -f docker-compose.dev.yml down -v
	docker system prune -f

dev: dev-up ## Start development server with hot reload
	@echo "Starting development server with hot reload..."
	@echo "Waiting for services to be ready..."
	@sleep 3
	air -c .air.toml

dev-web: ## Start complete development stack including web UI
	@echo "Starting complete development stack with web UI..."
	@./scripts/start-dev-stack.sh

web-build: ## Build web UI Docker image
	@echo "Building web UI Docker image..."
	@docker build -t gotak-web:dev ./web

web-dev: ## Start web UI in development mode
	@echo "Starting web UI in development mode..."
	@cd web && npm run dev

web-install: ## Install web UI dependencies
	@echo "Installing web UI dependencies..."
	@cd web && npm ci

# Default target
help: ## Show this help message
	@echo "GoTAK Server - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## Build the server binary
	@echo "Building GoTAK Server..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_PATH)
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(BINARY_PATH)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_PATH)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_PATH)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_PATH)
	@echo "All binaries built in $(BUILD_DIR)/"

run: build ## Build and run the server
	@echo "Starting GoTAK Server..."
	@$(BUILD_DIR)/$(BINARY_NAME) -config config/server.yaml

dev-debug: ## Run in development mode with debug logging
	@echo "Starting GoTAK Server in development mode..."
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@$(BUILD_DIR)/$(BINARY_NAME) -config config/server.yaml -debug

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: dev-up ## Run integration tests with Docker services
	@echo "Running integration tests..."
	@echo "Waiting for services to be ready..."
	@sleep 10
	@$(GOTEST) -v -tags=integration ./tests/integration/... || (make dev-down; exit 1)
	@make dev-down

test-scripts: ## Run script and configuration tests
	@echo "Running script tests..."
	@./tests/scripts/test_scripts.sh

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@$(GOFMT) ./...
	@goimports -w .

# Security scanning
security: ## Run security scans
	@echo "Running security scans..."
	@gosec ./...
	@trivy fs --security-checks vuln,secret,config .

security-audit: ## Run comprehensive security audit
	@echo "Running comprehensive security audit..."
	@./scripts/security-audit.sh full

security-audit-quick: ## Run quick security audit
	@echo "Running quick security audit..."
	@./scripts/security-audit.sh quick

security-audit-headers: ## Check HTTP security headers
	@echo "Checking HTTP security headers..."
	@./scripts/security-audit.sh headers

security-audit-tls: ## Check TLS/SSL configuration
	@echo "Checking TLS/SSL configuration..."
	@./scripts/security-audit.sh tls

security-audit-auth: ## Check authentication security
	@echo "Checking authentication security..."
	@./scripts/security-audit.sh auth

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) tidy

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t gotak-server:$(VERSION) .
	@docker tag gotak-server:$(VERSION) gotak-server:latest

docker-build-nomad: ## Build Docker images for Nomad deployment
	@echo "Building GoTAK images for Nomad..."
	@docker build --build-arg VERSION=$(VERSION) --build-arg BUILD_TIME=$(BUILD_TIME) --build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t gotak/server:$(VERSION) -t gotak/server:latest .
	@docker build -t gotak/web:$(VERSION) -t gotak/web:latest ./web
	@echo "Nomad images built:"
	@echo "  gotak/server:$(VERSION) (includes both server and web UI)"
	@echo "  gotak/web:$(VERSION) (web UI only)"

docker-push-nomad: docker-build-nomad ## Push Docker images for Nomad
	@echo "Pushing GoTAK images for Nomad..."
	@docker push gotak/server:$(VERSION)
	@docker push gotak/server:latest
	@docker push gotak/web:$(VERSION)
	@docker push gotak/web:latest

docker-build-hub: ## Build Docker images for DockerHub
	@echo "Building GoTAK images for DockerHub..."
	@docker build --build-arg VERSION=$(VERSION) --build-arg BUILD_TIME=$(BUILD_TIME) --build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(SERVER_IMAGE_NAME):$(VERSION) -t $(SERVER_IMAGE_NAME):latest .
	@docker build -t $(WEB_IMAGE_NAME):$(VERSION) -t $(WEB_IMAGE_NAME):latest ./web
	@echo "DockerHub images built:"
	@echo "  $(SERVER_IMAGE_NAME):$(VERSION) and $(SERVER_IMAGE_NAME):latest"
	@echo "  $(WEB_IMAGE_NAME):$(VERSION) and $(WEB_IMAGE_NAME):latest"

docker-push-hub: docker-build-hub ## Build and push Docker images to DockerHub
	@echo "Pushing GoTAK images to DockerHub as $(DOCKERHUB_USERNAME)..."
	@docker push $(SERVER_IMAGE_NAME):$(VERSION)
	@docker push $(SERVER_IMAGE_NAME):latest
	@docker push $(WEB_IMAGE_NAME):$(VERSION)
	@docker push $(WEB_IMAGE_NAME):latest
	@echo "Images pushed successfully to DockerHub!"

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8087:8087 -p 8089:8089 -p 8080:8080 --name gotak-server gotak-server:latest

docker-stop: ## Stop Docker container
	@docker stop gotak-server
	@docker rm gotak-server

# Development helpers
serve: ## Start development server with hot reload
	@echo "Starting development server with hot reload..."
	@air -c .air.toml

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@pip3 install pre-commit || echo "Pre-commit installation skipped (pip3 not available)"

precommit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@pre-commit install
	@pre-commit install --hook-type commit-msg

precommit-run: ## Run pre-commit hooks on all files
	@echo "Running pre-commit hooks..."
	@pre-commit run --all-files

precommit-update: ## Update pre-commit hooks
	@echo "Updating pre-commit hooks..."
	@pre-commit autoupdate

# Certificate generation for TLS
certs: ## Generate self-signed certificates for TLS
	@echo "Generating self-signed certificates..."
	@mkdir -p certs
	@openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
		-subj "/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=localhost" \
		-keyout certs/server.key -out certs/server.crt
	@echo "Certificates generated in certs/ directory"

# Database helpers (if using PostgreSQL)
db-create: ## Create database (requires PostgreSQL)
	@echo "Creating database..."
	@createdb gotak || echo "Database might already exist"

db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	# TODO: Implement database migrations

# Installation
install: build ## Install the binary to system path
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(BINARY_NAME) installed to /usr/local/bin/"

uninstall: ## Uninstall the binary from system path
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled"

# Release
release: clean build-all test ## Build release artifacts
	@echo "Creating release artifacts..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@tar -czf release/gotak-server-$(VERSION).tar.gz -C release/ .
	@echo "Release artifacts created in release/ directory"

version: ## Show version information
	@echo "GoTAK Server"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
