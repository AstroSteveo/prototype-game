# Development Dockerfile for Prototype Game Backend
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make git

# Set working directory
WORKDIR /app

# Copy Go modules first for better layer caching
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

# Copy source code
COPY backend/ .
COPY Makefile /app/

# Build the services
WORKDIR /app
RUN make build

# Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl

# Create non-root user
RUN addgroup -g 1000 gameuser && \
    adduser -D -s /bin/sh -u 1000 -G gameuser gameuser

# Set working directory
WORKDIR /app

# Copy built binaries
COPY --from=builder /app/backend/bin/ ./bin/

# Create directories for runtime data
RUN mkdir -p logs .pids && \
    chown -R gameuser:gameuser /app

# Switch to non-root user
USER gameuser

# Expose ports
EXPOSE 8080 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

# Default command runs both services
CMD ["sh", "-c", "./bin/sim -port 8081 & ./bin/gateway -port 8080 -sim localhost:8081"]