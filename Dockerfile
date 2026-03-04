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

# Install runtime dependencies
# - ca-certificates: for HTTPS connections
# - tzdata: for timezone support
# - curl: for web scraping
# - pcre-tools (pcregrep): for Perl-compatible regex support
RUN apk --no-cache add ca-certificates tzdata curl pcre-tools

# Create non-root user
RUN addgroup -g 1000 botuser && \
    adduser -D -u 1000 -G botuser botuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bot /app/bot

# Create directory for volume
RUN mkdir -p /app/workplace && \
    chown -R botuser:botuser /app

# Switch to non-root user
USER botuser

# Define volume mount point
VOLUME ["/app/workplace"]

# Environment variables (configurable at runtime)
ENV TELEGRAM_BOT_TOKEN="" \
    OPENROUTER_API_KEY="" \
    MODEL_NAME="google/gemini-2.5-flash-lite" \
    SHELL_TIMEOUT="30s" \
    LOG_LEVEL="info" \
    MEMORY_TOKEN_THRESHOLD="50000" \
    TOPIC_SIZE_THRESHOLD="102400" \
    NOTES_CLEANUP_ENABLED="true" \
    NOTES_MAX_AGE_DAYS="30" \
    NOTES_COMPLETED_RETENTION_DAYS="7" \
    NOTES_SCRATCHPAD_MAX_AGE_DAYS="7" \
    DAILY_MAINTENANCE_TIME="0 4 * * *"

# Expose no ports (bot uses polling, not webhooks)

# Set entrypoint to run bot binary
ENTRYPOINT ["/app/bot"]
