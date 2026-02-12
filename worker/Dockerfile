# Production Dockerfile for Payment Watchdog Worker
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker \
    -ldflags="-w -s -X main.version=${VERSION:-dev} -X main.buildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/worker .

# Create non-root user
USER appuser

# Worker doesn't expose ports, it's a background service
# Health check would need to be implemented differently (e.g., via metrics endpoint)

CMD ["./worker"]
