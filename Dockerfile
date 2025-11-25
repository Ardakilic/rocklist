# Rocklist Dockerfile
# Multi-stage build for development and production

# =============================================================================
# Stage 1: Base image with all build tools
# =============================================================================
FROM golang:1.23-bookworm AS base

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    npm \
    nodejs \
    git \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Install Wails CLI
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Set working directory
WORKDIR /app

# =============================================================================
# Stage 2: Development image
# =============================================================================
FROM base AS development

# Install additional development tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Copy go mod files first for better caching
COPY go.mod ./
COPY go.sum* ./
RUN go mod download || true

# Copy package.json for npm caching
COPY frontend/package*.json ./frontend/
RUN cd frontend && npm install

# Volume for source code
VOLUME /app

# Default command for development
CMD ["wails", "dev"]

# =============================================================================
# Stage 3: Builder image
# =============================================================================
FROM base AS builder

# Copy source code
COPY . .

# Install dependencies
RUN go mod download
RUN cd frontend && npm install

# Build the application
RUN wails build -platform linux/amd64 -tags webkit2_41

# =============================================================================
# Stage 4: Production image (minimal)
# =============================================================================
FROM debian:bookworm-slim AS production

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libgtk-3-0 \
    libwebkit2gtk-4.1-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -s /bin/bash rocklist
USER rocklist
WORKDIR /home/rocklist

# Copy built binary
COPY --from=builder /app/build/bin/Rocklist /usr/local/bin/rocklist

# Set entrypoint
ENTRYPOINT ["rocklist"]
