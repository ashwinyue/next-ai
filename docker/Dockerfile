# Build stage
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Install dependencies
RUN apt-get update && \
    apt-get install -y git build-essential ca-certificates

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o next-ai ./cmd/next-ai

# Final stage
FROM debian:12.12-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates curl tzdata && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -s /bin/bash appuser && \
    mkdir -p /data/files && \
    chown -R appuser:appuser /app /data/files

# Copy binary from builder
COPY --from=builder /app/next-ai .
COPY --from=builder /app/configs ./configs

# Expose port
EXPOSE 8080

# Switch to non-root user
USER appuser

CMD ["./next-ai"]
