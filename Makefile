# Rocklist Makefile
# Author: Arda Kılıçdağı <arda@kilicdagi.com>

.PHONY: help setup build build-linux build-windows build-darwin build-all dev test test-coverage test-quick lint lint-go lint-frontend fmt clean clean-all install frontend-install frontend-build shell docker-build docker-rebuild deps update-deps generate version pre-commit-install pre-commit

# Variables
APP_NAME := rocklist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/Ardakilic/rocklist/cmd.Version=$(VERSION) -X github.com/Ardakilic/rocklist/cmd.GitCommit=$(GIT_COMMIT) -X github.com/Ardakilic/rocklist/cmd.BuildDate=$(BUILD_DATE)"

# Docker compose command
DC := docker compose

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "Rocklist - Playlist Generator for Rockbox"
	@echo ""
	@echo "All commands run via Docker - no local Go/Node.js required"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Development
dev: ## Run the app in development mode
	@echo "$(GREEN)Starting development server...$(NC)"
	$(DC) run --rm dev wails dev

install: ## Install dependencies
	@echo "$(GREEN)Installing Go dependencies...$(NC)"
	$(DC) run --rm dev go mod download
	$(DC) run --rm dev go mod tidy
	@echo "$(GREEN)Installing frontend dependencies...$(NC)"
	$(DC) run --rm dev sh -c "cd frontend && npm install"

frontend-install: ## Install frontend dependencies only
	@echo "$(GREEN)Installing frontend dependencies...$(NC)"
	$(DC) run --rm dev sh -c "cd frontend && npm install"

frontend-build: ## Build frontend only
	@echo "$(GREEN)Building frontend...$(NC)"
	$(DC) run --rm dev sh -c "cd frontend && npm run build"

# Building
build: frontend-build ## Build for current platform
	@echo "$(GREEN)Building $(APP_NAME)...$(NC)"
	$(DC) run --rm build wails build $(LDFLAGS)

build-linux: frontend-build ## Build for Linux (amd64)
	@echo "$(GREEN)Building $(APP_NAME) for Linux...$(NC)"
	$(DC) run --rm build wails build $(LDFLAGS) -platform linux/amd64

build-windows: frontend-build ## Build for Windows (amd64)
	@echo "$(GREEN)Building $(APP_NAME) for Windows...$(NC)"
	$(DC) run --rm build wails build $(LDFLAGS) -platform windows/amd64

build-darwin: frontend-build ## Build for macOS (universal)
	@echo "$(GREEN)Building $(APP_NAME) for macOS (universal)...$(NC)"
	$(DC) run --rm build wails build $(LDFLAGS) -platform darwin/universal

build-darwin-amd64: frontend-build ## Build for macOS (amd64)
	@echo "$(GREEN)Building $(APP_NAME) for macOS (amd64)...$(NC)"
	$(DC) run --rm build wails build $(LDFLAGS) -platform darwin/amd64

build-darwin-arm64: frontend-build ## Build for macOS (arm64)
	@echo "$(GREEN)Building $(APP_NAME) for macOS (arm64)...$(NC)"
	$(DC) run --rm build wails build $(LDFLAGS) -platform darwin/arm64

build-all: build-linux build-windows build-darwin ## Build for all platforms

# Testing
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(DC) run --rm test

test-quick: ## Run tests quickly (no race detection, for pre-commit)
	@echo "$(GREEN)Running quick tests...$(NC)"
	$(DC) run --rm dev go test -short ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(DC) run --rm dev go test -v -coverprofile=coverage.out -covermode=atomic ./...
	$(DC) run --rm dev go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-coverage-check: test-coverage ## Check if coverage is above 90%
	@echo "$(GREEN)Checking coverage threshold...$(NC)"
	@$(DC) run --rm dev sh -c 'coverage=$$(go tool cover -func=coverage.out | grep total | awk "{print \$$3}" | sed "s/%//"); \
	if [ $$(echo "$$coverage < 90" | bc) -eq 1 ]; then \
		echo "Coverage is $$coverage%, which is below 90%"; \
		exit 1; \
	else \
		echo "Coverage is $$coverage%"; \
	fi'

# Linting
lint: lint-go lint-frontend ## Run all linters

lint-go: ## Run Go linter only
	@echo "$(GREEN)Running Go linter...$(NC)"
	$(DC) run --rm lint

lint-frontend: ## Run frontend linter only
	@echo "$(GREEN)Running frontend linter...$(NC)"
	$(DC) run --rm dev sh -c "cd frontend && npm run lint"

# Formatting
fmt: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	$(DC) run --rm dev go fmt ./...

# Pre-commit
pre-commit-install: ## Install pre-commit hooks
	@echo "$(GREEN)Installing pre-commit hooks...$(NC)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(YELLOW)Installing pre-commit...$(NC)"; pip install pre-commit; }
	pre-commit install
	@echo "$(GREEN)Pre-commit hooks installed!$(NC)"

pre-commit: ## Run pre-commit on all files
	@echo "$(GREEN)Running pre-commit checks...$(NC)"
	pre-commit run --all-files

# Setup
setup: ## Create required directories for caching
	@echo "$(GREEN)Creating cache directories...$(NC)"
	mkdir -p .cache/go-mod
	mkdir -p frontend/node_modules
	mkdir -p build

# Cleaning
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf build/
	rm -rf frontend/dist/
	rm -f coverage.out coverage.html

clean-all: clean ## Clean everything including caches
	@echo "$(YELLOW)Cleaning all caches...$(NC)"
	rm -rf .cache/
	rm -rf frontend/node_modules/

# Docker
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(APP_NAME):$(VERSION) --target development .

docker-rebuild: ## Force rebuild Docker image (no cache)
	@echo "$(GREEN)Rebuilding Docker image (no cache)...$(NC)"
	docker build --no-cache -t $(APP_NAME):$(VERSION) --target development .

shell: ## Open a shell in Docker container
	@echo "$(GREEN)Opening shell in Docker...$(NC)"
	$(DC) run --rm dev bash

# Utilities
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

deps: ## Show dependencies
	@echo "$(GREEN)Go dependencies:$(NC)"
	$(DC) run --rm dev go list -m all

update-deps: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	$(DC) run --rm dev go get -u ./...
	$(DC) run --rm dev go mod tidy
	@echo "$(GREEN)Updating frontend dependencies...$(NC)"
	$(DC) run --rm dev sh -c "cd frontend && npm update"

generate: ## Generate Wails bindings
	@echo "$(GREEN)Generating Wails bindings...$(NC)"
	$(DC) run --rm dev wails generate module
