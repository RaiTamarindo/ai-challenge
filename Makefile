# Feature Voting Platform - Infrastructure Management

# Default environment variables
export POSTGRES_HOSTNAME ?= localhost
export POSTGRES_PORT ?= 5432
export POSTGRES_ADMIN_USERNAME ?= postgres
export POSTGRES_ADMIN_PASSWORD ?= postgres_admin_pass
export POSTGRES_STANDARD_USERNAME ?= voting_app
export POSTGRES_STANDARD_PASSWORD ?= voting_app_pass
export POSTGRES_DB ?= feature_voting_platform

.PHONY: help infra infra-up infra-down infra-logs infra-clean migrate-up migrate-down migrate-status migration db-setup api api-build api-down api-logs up up-build down rebuild user

help: ## Show this help message
	@echo "Feature Voting Platform - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

infra: infra-up db-setup migrate-up ## Start complete infrastructure (database + migrations)

infra-up: ## Start database infrastructure
	@echo "Starting database infrastructure..."
	@echo "Environment variables:"
	@echo "  POSTGRES_HOSTNAME=$(POSTGRES_HOSTNAME)"
	@echo "  POSTGRES_PORT=$(POSTGRES_PORT)"
	@echo "  POSTGRES_DB=$(POSTGRES_DB)"
	@echo "  POSTGRES_ADMIN_USERNAME=$(POSTGRES_ADMIN_USERNAME)"
	@echo ""
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@timeout=60; \
	while ! docker-compose exec -T postgres pg_isready -U $(POSTGRES_ADMIN_USERNAME) -d $(POSTGRES_DB) >/dev/null 2>&1; do \
		if [ $$timeout -le 0 ]; then \
			echo "Database failed to start within 60 seconds"; \
			exit 1; \
		fi; \
		echo "Waiting for database... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done
	@echo "Database is ready!"

infra-down: ## Stop infrastructure
	@echo "Stopping infrastructure..."
	docker-compose down

infra-logs: ## Show infrastructure logs
	docker-compose logs -f postgres

infra-clean: ## Clean up infrastructure (removes volumes)
	@echo "WARNING: This will delete all database data!"
	@read -p "Are you sure? (y/N) " confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]
	docker-compose down -v
	docker volume rm feature-voting-platform_postgres_data 2>/dev/null || true

db-setup: ## Create database users and permissions
	@echo "Setting up database users and permissions..."
	@chmod +x migrations/setup_db.sh
	@docker-compose exec -T postgres /bin/sh /docker-entrypoint-initdb.d/setup_db.sh

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	docker-compose --profile migration run --rm sql-migrate -direction=up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	docker-compose --profile migration run --rm sql-migrate -direction=down

migrate-status: ## Check migration status
	docker-compose --profile migration run --rm sql-migrate -direction=status

migration: ## Create new migration file (usage: make migration name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name using: make migration name=your_migration_name"; \
		exit 1; \
	fi
	@timestamp=$$(date "+%Y%m%d%H%M%S"); \
	filename="migrations/$${timestamp}_$(name).sql"; \
	echo "Creating migration file: $$filename"; \
	echo "-- +migrate Up" > $$filename; \
	echo "" >> $$filename; \
	echo "" >> $$filename; \
	echo "-- +migrate Down" >> $$filename; \
	echo "Migration file created successfully: $$filename"

# Database connection helper
db-connect: ## Connect to database via psql
	docker-compose exec postgres psql -U $(POSTGRES_ADMIN_USERNAME) -d $(POSTGRES_DB)

db-connect-app: ## Connect to database as application user
	docker-compose exec postgres psql -U $(POSTGRES_STANDARD_USERNAME) -d $(POSTGRES_DB)

# Development helpers
dev-reset: infra-down infra migrate-up ## Reset development environment

rebuild: ## Force rebuild all images and restart services
	@echo "Stopping all services..."
	docker-compose down
	@echo "Removing existing images..."
	docker-compose build --no-cache
	@echo "Starting services with fresh images..."
	$(MAKE) up

# API management
api: ## Start API service
	docker-compose build --no-cache api
	docker-compose up -d api

api-down: ## Stop API service
	docker-compose stop api

api-logs: ## Show API logs
	docker-compose logs -f api

# Development workflow
up: infra api ## Start complete development environment (database + API)

down: ## Stop complete development environment
	docker-compose down

# CLI commands
user: ## Create a new user (usage: make user name=username email=user@email.com password=password)
	@if [ -z "$(name)" ] || [ -z "$(email)" ] || [ -z "$(password)" ]; then \
		echo "Error: All parameters are required."; \
		echo "Usage: make user name=<username> email=<email> password=<password>"; \
		echo "Example: make user name=Rai email=raitamarindo@gmail.com password=12345"; \
		exit 1; \
	fi
	@echo "Creating user: $(name) <$(email)>"
	@docker-compose --profile cli run --rm cli -command=create-user -name="$(name)" -email="$(email)" -password="$(password)"

# Show current environment
env: ## Show current environment variables
	@echo "Current environment variables:"
	@echo "  POSTGRES_HOSTNAME=$(POSTGRES_HOSTNAME)"
	@echo "  POSTGRES_PORT=$(POSTGRES_PORT)"
	@echo "  POSTGRES_ADMIN_USERNAME=$(POSTGRES_ADMIN_USERNAME)"
	@echo "  POSTGRES_ADMIN_PASSWORD=$(POSTGRES_ADMIN_PASSWORD)"
	@echo "  POSTGRES_STANDARD_USERNAME=$(POSTGRES_STANDARD_USERNAME)"
	@echo "  POSTGRES_STANDARD_PASSWORD=$(POSTGRES_STANDARD_PASSWORD)"
	@echo "  POSTGRES_DB=$(POSTGRES_DB)"