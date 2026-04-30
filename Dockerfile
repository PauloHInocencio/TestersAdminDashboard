FROM golang:1.25.9-bookworm AS base

# Stage dedicated to install project dependencies
FROM base AS build-base

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (cached unless go.mod/go.sum changes)
RUN go mod download

# Stage dedicated to hot-reloading application
FROM build-base AS development

# Install the air CLI for auto-reloading
RUN go install github.com/air-verse/air@v1.61.7

# Create directory for SSL certificates
RUN mkdir -p /app/certs

# Start air for live reloading
CMD ["air", "-c", ".air.toml"]

# Stage dedicated to build the application executable
FROM build-base as build-prod

# Copy all source code
COPY . .

# Build static binary with CGO disabled for portability
# Output to bin/testers-admin-api as per project name
RUN CGO_ENABLED=0 go build -o bin/testers-admin-api main.go


# Stage dedicated to run the application binary
FROM debian:bookworm-slim AS deploy

WORKDIR /app

# Install ca-certificates for HTTPS and SSL verification
# Clean up apt cache to minimize image size
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app

# Create directory for SSL certificates
RUN mkdir -p /app/certs && chown appuser:appuser /app/certs

# Copy binary from builder stage
COPY --from=build-prod /app/bin/testers-admin-api ./testers-admin-api

# Switch to non-root user
USER appuser

# Expose API port
EXPOSE 8080

# Run the application
CMD ["./testers-admin-api"]