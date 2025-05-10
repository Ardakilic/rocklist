# Build stage
FROM golang:1.24.3 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Try to download dependencies
RUN go mod download

# Copy source
COPY . .

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -o rocklist cmd/main.go

# Production stage
FROM gcr.io/distroless/base-debian12:nonroot AS final

LABEL org.opencontainers.image.authors="Arda Kilicdagi <arda@kilicdagi.com>" \
      org.opencontainers.image.url="https://github.com/ardakilic/rocklist" \
      org.opencontainers.image.documentation="https://github.com/ardakilic/rocklist" \
      org.opencontainers.image.source="https://github.com/ardakilic/rocklist" \
      org.opencontainers.image.title="rocklist" \
      org.opencontainers.image.description="RockList - A dynamic playlist generator for Rockbox"

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /app/rocklist /app/rocklist


# Define volume for rockbox data
VOLUME /rockbox

# Set entrypoint
ENTRYPOINT ["/app/rocklist"]
