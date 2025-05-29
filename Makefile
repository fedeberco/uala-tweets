build:
	go build -o bin/uala-tweets

run:
	go mod tidy
	go run main.go

test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

deps:
	go mod download

.PHONY: build run test clean deps
