# Build stage
FROM golang:1.20-alpine AS build

WORKDIR /app

# Install git and dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod ./
# Try to download dependencies
RUN go mod download

# Copy source
COPY . .

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o rocklist cmd/main.go

# Production stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/rocklist /app/rocklist

# Run as non-root user
USER 65532:65532

# Define volume for rockbox data
VOLUME /rockbox

# Set entrypoint
ENTRYPOINT ["/app/rocklist"] 