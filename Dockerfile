# Gowright Testing Framework Dockerfile

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    chromium \
    chromium-chromedriver \
    sqlite \
    postgresql-client \
    mysql-client \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 gowright && \
    adduser -D -s /bin/sh -u 1001 -G gowright gowright

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/gowright /usr/local/bin/gowright

# Copy configuration and examples
COPY --from=builder /app/examples ./examples
COPY --from=builder /app/gowright-config.json ./gowright-config.json

# Create directories for reports and logs
RUN mkdir -p /app/reports /app/logs && \
    chown -R gowright:gowright /app

# Switch to non-root user
USER gowright

# Set environment variables
ENV GOWRIGHT_CONFIG=/app/gowright-config.json
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser

# Expose port for documentation server
EXPOSE 6060

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD gowright version || exit 1

# Default command
CMD ["gowright", "--help"]

# Labels
LABEL maintainer="Gowright Team <team@gowright.dev>"
LABEL version="1.0.0"
LABEL description="Gowright Testing Framework - Comprehensive testing for Go applications"
LABEL org.opencontainers.image.source="https://github.com/your-org/gowright"
LABEL org.opencontainers.image.documentation="https://github.com/your-org/gowright/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"