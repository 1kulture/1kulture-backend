.PHONY: help start dev run build clean test lint fmt vet swagger deps tidy setup

# Variables
APP_NAME := 1kulture-api
BINARY_NAME := 1kulture-api
GO := go
MAIN_PATH := ./cmd/api
BUILD_DIR := ./bin
SWAGGER_DIR := ./docs
ENV_FILE := .env
ENV_EXAMPLE := .env.example

# Colors
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(GREEN)  1Kulture API - Commands$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@echo "$(GREEN)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'
	@echo "$(BLUE)========================================$(NC)"

setup: ## Setup project (create .env, install deps)
	@echo "$(GREEN)Setting up project...$(NC)"
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(YELLOW)Creating .env from .env.example...$(NC)"; \
		cp $(ENV_EXAMPLE) $(ENV_FILE); \
		echo "$(GREEN).env file created. Please edit it with your values.$(NC)"; \
	else \
		echo "$(YELLOW).env already exists.$(NC)"; \
	fi
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GO) mod download
	@echo "$(GREEN)Setup complete!$(NC)"

start: ## Start the application
	@echo "$(GREEN)Starting $(APP_NAME)...$(NC)"
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(RED)Error: .env not found! Run 'make setup' first.$(NC)"; \
		exit 1; \
	fi
	@$(GO) run $(MAIN_PATH)/main.go

dev: start ## Alias for start
run: start ## Alias for start

build: ## Build the application
	@echo "$(GREEN)Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)/main.go
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GO) test ./... -v

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Skipping.$(NC)"; \
	fi

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GO) fmt ./...

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GO) vet ./...

swagger: ## Generate Swagger docs
	@echo "$(GREEN)Generating Swagger docs...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g $(MAIN_PATH)/main.go -o $(SWAGGER_DIR); \
		echo "$(GREEN)Swagger docs generated$(NC)"; \
	else \
		echo "$(RED)swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
	fi

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GO) mod download

tidy: ## Tidy dependencies
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	@$(GO) mod tidy