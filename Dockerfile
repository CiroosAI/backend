# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25-alpine AS builder

# Install build tools and security updates
RUN apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

# Create non-root user for building
RUN adduser -D -g '' appuser

WORKDIR /src

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/server ./main.go

# Runtime stage - use distroless for security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/server /app/server

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "-health-check"] || exit 1

# Run as non-root user
USER nonroot:nonroot

# Start the application
ENTRYPOINT ["/app/server"]
