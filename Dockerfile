# Multi-stage Dockerfile for Feature Voting Platform

# Build stage
FROM golang:1.21-alpine AS builder

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

# Build the migration binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

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

# API stage (for future backend API - placeholder for now)
FROM alpine:3.18 AS api

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

WORKDIR /app

# Copy API binary (placeholder - will be created in Stage 2)
# COPY --from=builder /app/api /app/api

# Change ownership
RUN chown -R app:app /app

USER app

# Placeholder command
CMD ["echo", "API stage ready for implementation"]