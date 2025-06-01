# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .


# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o uala-tweets

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/uala-tweets .

# Create and copy migrations
RUN mkdir -p /app/db/migrations
COPY --from=builder /app/db/migrations/ /app/db/migrations/

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./uala-tweets"]
