# ==============================================================================
# Makefile for uala-tweets
# 
# Available commands:
#   Local Development:
#     make start-local      # Start all services in Docker and run app locally
#     make run-local        # Run the application locally (assumes services are running)
#     make test-local       # Run tests locally
#     make install-deps     # Install Go dependencies
#     make migrate-up-local # Run database migrations
# 
#   Docker Development:
#     make start-docker     # Start all services in Docker
#     make stop-docker      # Stop all Docker services
#     make docker-logs      # View Docker logs
#
#   Database:
#     make migrate-up       # Run migrations (Docker)
#     make migrate-down     # Rollback last migration (Docker)
#     make migrate-up-local # Run migrations (local)
#     make migrate-down-local # Rollback last migration (local)
# ==============================================================================

# ==============================================================================
# Configuration
# ==============================================================================

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= uala_tweets
DB_USER ?= uala
DB_PASS ?= ualapass
DB_SSLMODE ?= disable
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Application configuration
APP_NAME := uala-tweets
APP_PORT ?= 8080

# ==============================================================================
# Local Development
# ==============================================================================

## Start all services in Docker and run the application locally
start-local: docker-up migrate-up
	@echo "Starting all services in Docker..."
	@echo "\nStarting $(APP_NAME) locally on port 8000..."
	@echo "\nTo stop services, run: make stop-docker\n"
	@echo "Application will be available at: http://localhost:8000\n"
	@DB_URL="$(DB_URL)" \
	 REDIS_ADDR="localhost:6379" \
	 KAFKA_BROKER="localhost:29092" \
	 PORT=8000 \
	 go run main.go

## Run the application locally (assumes services are already running)
run-local: install-deps
	@echo "Starting $(APP_NAME) locally..."
	@DB_URL="$(DB_URL)" \
	 REDIS_ADDR="localhost:6379" \
	 KAFKA_BROKER="localhost:29092" \
	 go run main.go

## Install Go dependencies
install-deps:
	@echo "Installing dependencies..."
	@go mod download

## Start test environment
test-env-up:
	@echo "Starting test environment..."
	@docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for test database to be ready..."
	@until docker compose -f docker-compose.test.yml exec -T testdb pg_isready -U postgres; do \
		echo "Waiting for PostgreSQL..."; \
		sleep 2; \
	done

## Stop test environment
test-env-down:
	@echo "Stopping test environment..."
	@docker compose -f docker-compose.test.yml down

## Run tests locally
test-local: test-env-up
	@echo "Running tests..."
	@TEST_DB_HOST=localhost \
	 TEST_DB_PORT=5433 \
	 TEST_DB_USER=postgres \
	 TEST_DB_PASSWORD=postgres \
	 TEST_DB_NAME=testdb \
	 go test -v ./...
	@make test-env-down

## Build the application
build:
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME)

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

# ==============================================================================
# Database Migrations (Local)
# ==============================================================================

## Run database migrations locally
migrate-up-local:
	@echo "Running migrations..."
	@migrate -path ./db/migrations -database "$(DB_URL)" up

## Rollback the last migration locally
migrate-down-local:
	@echo "Rolling back last migration..."
	@migrate -path ./db/migrations -database "$(DB_URL)" down

## Rollback all migrations locally
migrate-down-all-local:
	@echo "Rolling back all migrations..."
	@migrate -path ./db/migrations -database "$(DB_URL)" down -all

# ==============================================================================
# Docker Development
# ==============================================================================

## Start all services in Docker
start-docker: docker-up migrate-up
	@echo "\n$(APP_NAME) is now running at http://localhost:8081"
	@echo "\nTo stop the services, run: make stop-docker\n"

## Stop all Docker services
stop-docker: docker-down

## View Docker logs
docker-logs:
	@docker compose logs -f

# ==============================================================================
# Docker Compose Commands
# ==============================================================================

## Build Docker images
docker-build:
	@docker compose build

## Start services in detached mode
docker-up:
	@docker compose up -d

## Stop and remove containers
docker-down:
	@docker compose down

## Restart all services
docker-restart: docker-down docker-up

# ==============================================================================
# Database Migrations (Docker)
# ==============================================================================

## Run database migrations (Docker)
migrate-up:
	@echo "Running migrations in Docker..."
	@docker run --rm -v $(shell pwd)/db/migrations:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "$(DB_URL)" \
		-verbose up

## Rollback last migration (Docker)
migrate-down:
	@docker run --rm -v $(shell pwd)/db/migrations:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "$(DB_URL)" \
		-verbose down

## Rollback all migrations (Docker)
migrate-down-all:
	@docker run --rm -v $(shell pwd)/db/migrations:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "$(DB_URL)" \
		-verbose down -all

# ==============================================================================
# Help
# ==============================================================================

## Show this help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: help run-local install-deps test-local build clean \
        migrate-up-local migrate-down-local migrate-down-all-local \
        start-docker stop-docker docker-logs \
        docker-build docker-up docker-down docker-restart \
        migrate-up migrate-down migrate-down-all
