# =============================================================================
# Webhook.inst.lk - Makefile
# =============================================================================

.PHONY: help build dev up down logs migrate test clean

# Default target
help: ## Show this help
	@echo "Webhook.inst.lk Development Commands"
	@echo "======================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Development
# =============================================================================

dev: ## Start development environment
	docker compose up -d postgres redis
	@echo "Waiting for services..."
	@sleep 3
	@echo "Services ready. Start backend and frontend manually."

build: ## Build all services
	docker compose build

up: ## Start all services
	docker compose up -d

down: ## Stop all services
	docker compose down

restart: ## Restart all services
	docker compose restart

logs: ## View logs for all services
	docker compose logs -f

logs-api: ## View API logs
	docker compose logs -f api

logs-frontend: ## View frontend logs
	docker compose logs -f frontend

# =============================================================================
# Database
# =============================================================================

migrate: ## Run database migrations
	docker compose exec postgres psql -U $${POSTGRES_USER:-webhook} -d $${POSTGRES_DB:-webhook} -f /docker-entrypoint-initdb.d/001_initial.up.sql

migrate-down: ## Rollback migrations
	docker compose exec postgres psql -U $${POSTGRES_USER:-webhook} -d $${POSTGRES_DB:-webhook} -f /docker-entrypoint-initdb.d/001_initial.down.sql

db-shell: ## Open PostgreSQL shell
	docker compose exec postgres psql -U $${POSTGRES_USER:-webhook} -d $${POSTGRES_DB:-webhook}

redis-shell: ## Open Redis CLI
	docker compose exec redis redis-cli -a $${REDIS_PASSWORD}

# =============================================================================
# Backend
# =============================================================================

backend-build: ## Build backend binaries
	cd backend && CGO_ENABLED=0 go build -o bin/api ./cmd/api
	cd backend && CGO_ENABLED=0 go build -o bin/dns ./cmd/dns
	cd backend && CGO_ENABLED=0 go build -o bin/smtp ./cmd/smtp
	cd backend && CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

backend-run: ## Run API server locally
	cd backend && go run ./cmd/api

backend-test: ## Run backend tests
	cd backend && go test ./... -v -cover

backend-lint: ## Lint backend code
	cd backend && golangci-lint run

# =============================================================================
# Frontend
# =============================================================================

frontend-install: ## Install frontend dependencies
	cd frontend && npm install

frontend-dev: ## Start frontend dev server
	cd frontend && npm run dev

frontend-build: ## Build frontend
	cd frontend && npm run build

frontend-lint: ## Lint frontend code
	cd frontend && npm run lint

frontend-typecheck: ## TypeScript type check
	cd frontend && npm run type-check

# =============================================================================
# Production
# =============================================================================

prod-build: ## Build for production
	docker compose -f docker-compose.yml build

prod-up: ## Start production
	docker compose -f docker-compose.yml up -d

prod-deploy: prod-build prod-up ## Build and deploy
	@echo "Deployment complete!"

# =============================================================================
# Utilities
# =============================================================================

clean: ## Clean build artifacts and volumes
	docker compose down -v
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules

status: ## Show service status
	docker compose ps

health: ## Check health endpoints
	@echo "API Health:"
	@curl -s http://localhost:2200/health | python3 -m json.tool 2>/dev/null || echo "API not running"
	@echo "\nAPI Ready:"
	@curl -s http://localhost:2200/ready | python3 -m json.tool 2>/dev/null || echo "API not ready"

generate-env: ## Generate .env from .env.example
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Generated .env file. Please update secrets!"; \
	else \
		echo ".env already exists"; \
	fi
