# Production Dockerfile - Multi-stage build
FROM golang:1.24.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary with static linking
RUN CGO_ENABLED=0 go build -a \
    -ldflags '-extldflags "-static"' \
    -o ./build/app ./cmd/...

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/build/app .

# Change ownership to non-root user
RUN chown appuser:appuser /app/app

# Switch to non-root user
USER appuser

# Set entrypoint to the app binary
ENTRYPOINT ["./app"]

# Default command if no args provided
CMD ["--help"]