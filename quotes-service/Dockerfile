# Build stage
FROM golang:1.22.1-alpine AS builder

# Install git and ca-certificates (needed for downloading modules)
RUN apk update && apk add --no-cache git ca-certificates curl && update-ca-certificates

# Create appuser for security
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o quotes-service \
    cmd/server/main.go

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy user from builder
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary
COPY --from=builder /build/quotes-service /quotes-service

# Use unprivileged user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/quotes-service", "-health-check"]

# Run binary
ENTRYPOINT ["/quotes-service"]