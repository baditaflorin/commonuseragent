.PHONY: help build test test-coverage test-race clean run run-demo docker-build docker-run lint security-scan fmt vet install-tools

# Variables
APP_NAME=useragent-demo
BINARY_NAME=useragent-demo
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-w -s"
COVERAGE_FILE=coverage.out
DOCKER_IMAGE=commonuseragent:latest

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

help: ## Show this help message
	@echo '$(GREEN)Available targets:$(RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

install-tools: ## Install development tools
	@echo '$(GREEN)Installing development tools...$(RESET)'
	@which golangci-lint > /dev/null || (echo 'Installing golangci-lint...' && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which gosec > /dev/null || (echo 'Installing gosec...' && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@which staticcheck > /dev/null || (echo 'Installing staticcheck...' && go install honnef.co/go/tools/cmd/staticcheck@latest)

build: ## Build the library
	@echo '$(GREEN)Building library...$(RESET)'
	$(GO) build $(GOFLAGS) ./...

build-demo: ## Build the demo application
	@echo '$(GREEN)Building demo application...$(RESET)'
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/demo

test: ## Run tests
	@echo '$(GREEN)Running tests...$(RESET)'
	$(GO) test $(GOFLAGS) -timeout 30s ./...

test-coverage: ## Run tests with coverage
	@echo '$(GREEN)Running tests with coverage...$(RESET)'
	$(GO) test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo '$(GREEN)Coverage report generated: coverage.html$(RESET)'

test-race: ## Run tests with race detector
	@echo '$(GREEN)Running tests with race detector...$(RESET)'
	$(GO) test -v -race -timeout 60s ./...

bench: ## Run benchmarks
	@echo '$(GREEN)Running benchmarks...$(RESET)'
	$(GO) test -bench=. -benchmem ./...

fmt: ## Format Go code
	@echo '$(GREEN)Formatting code...$(RESET)'
	$(GO) fmt ./...

vet: ## Run go vet
	@echo '$(GREEN)Running go vet...$(RESET)'
	$(GO) vet ./...

lint: install-tools ## Run linter
	@echo '$(GREEN)Running linter...$(RESET)'
	golangci-lint run ./...

security-scan: install-tools ## Run security scanner
	@echo '$(GREEN)Running security scanner...$(RESET)'
	gosec -exclude-generated ./...

staticcheck: install-tools ## Run staticcheck
	@echo '$(GREEN)Running staticcheck...$(RESET)'
	staticcheck ./...

check: fmt vet lint staticcheck security-scan ## Run all checks (fmt, vet, lint, staticcheck, security)
	@echo '$(GREEN)All checks passed!$(RESET)'

run-demo: build-demo ## Run the demo application
	@echo '$(GREEN)Starting demo application...$(RESET)'
	./bin/$(BINARY_NAME)

run-demo-dev: ## Run the demo application in development mode
	@echo '$(GREEN)Starting demo application (development mode)...$(RESET)'
	APP_ENV=development LOG_LEVEL=debug $(GO) run ./cmd/demo

clean: ## Clean build artifacts
	@echo '$(GREEN)Cleaning build artifacts...$(RESET)'
	rm -rf bin/
	rm -f $(COVERAGE_FILE) coverage.html
	rm -f *.db
	$(GO) clean -cache -testcache

docker-build: ## Build Docker image
	@echo '$(GREEN)Building Docker image...$(RESET)'
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo '$(GREEN)Running Docker container...$(RESET)'
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE)

docker-compose-up: ## Start services with docker-compose
	@echo '$(GREEN)Starting services with docker-compose...$(RESET)'
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo '$(GREEN)Stopping services with docker-compose...$(RESET)'
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

mod-download: ## Download dependencies
	@echo '$(GREEN)Downloading dependencies...$(RESET)'
	$(GO) mod download

mod-tidy: ## Tidy dependencies
	@echo '$(GREEN)Tidying dependencies...$(RESET)'
	$(GO) mod tidy

mod-verify: ## Verify dependencies
	@echo '$(GREEN)Verifying dependencies...$(RESET)'
	$(GO) mod verify

deps: mod-download mod-tidy mod-verify ## Manage dependencies (download, tidy, verify)

all: clean deps check test build build-demo ## Run all checks, tests, and build everything
	@echo '$(GREEN)All tasks completed successfully!$(RESET)'

ci: deps check test-coverage ## Run CI pipeline
	@echo '$(GREEN)CI pipeline completed!$(RESET)'

.DEFAULT_GOAL := help
