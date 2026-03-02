# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Runtime stage
FROM alpine:latest

# Install runtime dependencies including curl for web scraping
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 botuser && \
    adduser -D -u 1000 -G botuser botuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bot /app/bot

# Create directories for volumes
RUN mkdir -p /app/agent /app/logs && \
    chown -R botuser:botuser /app

# Switch to non-root user
USER botuser

# Define volume mount points
VOLUME ["/app/agent", "/app/logs"]

# Environment variables (configurable at runtime)
ENV TELEGRAM_BOT_TOKEN="" \
    OPENROUTER_API_KEY="" \
    MODEL_NAME="google/gemini-2.5-flash-lite" \
    SHELL_TIMEOUT="30s" \
    LOG_LEVEL="info"

# Expose no ports (bot uses polling, not webhooks)

# Set entrypoint to run bot binary
ENTRYPOINT ["/app/bot"]
