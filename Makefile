# ------ GO
build:
	go build -o bin/uala-tweets

run:
	go mod tidy
	DB_URL=$(DB_URL) go run main.go

test: test-db-up
	@echo "Running tests..."
	go test -v ./...

cover: test-db-up
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ coverage.out coverage.html

deps:
	go mod download

# ------ /GO


# ------ DATABASE
# Database URL for migrations
DB_URL=postgres://uala:ualapass@localhost:5432/uala_tweets?sslmode=disable

# Test database management
test-db-up:
	@echo "Starting test database..."
	@docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 2

# Migration commands using Docker
migrate-up: up
	@echo "Running migrations..."
	@docker run --rm -v $(shell pwd)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" -verbose up

migrate-down: up
	@echo "Rolling back last migration..."
	@docker run --rm -v $(shell pwd)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" -verbose down

migrate-down-all: up
	@echo "Rolling back all migrations..."
	@docker run --rm -v $(shell pwd)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" -verbose down -all

# ------ /DATABASE

# ------ DOCKER
# Docker Compose commands
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps
# ------/DOCKER

# Start everything
start: up migrate-up run

# Stop everything
stop: down

.PHONY: build run test cover test-db-up clean deps migrate-up migrate-down migrate-down-all up down logs ps start stop
