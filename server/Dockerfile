# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies (CGO requires gcc and musl-dev for SQLite)
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application (CGO disabled since we use PostgreSQL in Docker)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/migrate cmd/migrate/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/cleanup cmd/cleanup/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies and health check tools
RUN apk --no-cache add ca-certificates tzdata postgresql-client wget curl

# Copy binaries from builder
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/bin/cleanup /app/cleanup

# Copy entrypoint script
COPY scripts/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Expose port
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]

