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

# Docker Compose commands
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

# Start everything
start: up run

# Stop everything
stop: down

.PHONY: build run test clean deps up down logs ps start stop
