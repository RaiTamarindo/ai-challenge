# Multi-stage Dockerfile for Feature Voting Platform

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go modules files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ ./

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cli ./cmd/cli

# SQL Migrate stage (for running migrations)
FROM alpine:3.18 AS sql-migrate

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates postgresql-client

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

WORKDIR /app

# Copy the migration binary from builder
COPY --from=builder /app/migrate /app/migrate

# Change ownership
RUN chown -R app:app /app

USER app

# Default command
ENTRYPOINT ["/app/migrate"]

# API stage (for backend API)
FROM alpine:3.18 AS api

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

WORKDIR /app

# Copy API binary from builder
COPY --from=builder /app/api /app/api

# Change ownership
RUN chown -R app:app /app

USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the API
CMD ["/app/api"]

# CLI stage (for running CLI commands)
FROM alpine:3.18 AS cli

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

WORKDIR /app

# Copy CLI binary from builder
COPY --from=builder /app/cli /app/cli

# Change ownership
RUN chown -R app:app /app

USER app

# Run the CLI
ENTRYPOINT ["/app/cli"]