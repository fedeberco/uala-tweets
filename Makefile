# ------ GO
build:
	go build -o bin/uala-tweets

run:
	go mod tidy
	go run main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

deps:
	go mod download
# ------ GO

# ------ DATABASE
# Database URL for migrations
DB_URL=postgres://uala:ualapass@localhost:5432/uala_tweets?sslmode=disable

# Migration commands using Docker
migrate-up:
	docker run --rm -v $(shell pwd)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" -verbose up

migrate-down:
	docker run --rm -v $(shell pwd)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" -verbose down

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
start: up run

# Stop everything
stop: down

.PHONY: build run test clean deps migrate-up migrate-down up down logs ps start stop
