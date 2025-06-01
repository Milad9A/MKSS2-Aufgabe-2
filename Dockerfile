# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies and verify go.sum
RUN go mod download && go mod verify

# Copy source code
COPY *.go ./

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o robot-api .

# Run stage
FROM alpine:latest

# Add ca certificates for HTTPS and basic utilities
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh appuser

# Set timezone
ENV TZ=UTC

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/robot-api .

# Change ownership to non-root user
RUN chown appuser:appuser /app/robot-api && \
    chmod +x /app/robot-api

# Switch to non-root user
USER appuser

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./robot-api"]
