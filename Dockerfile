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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o robot-api .

# Run stage
FROM alpine:latest

# Add ca certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/robot-api .

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./robot-api"]
