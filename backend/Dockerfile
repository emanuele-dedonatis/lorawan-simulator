# Build stage
FROM golang:alpine AS builder

# Allow Go to download the required toolchain version
ENV GOTOOLCHAIN=auto

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o lorawan-simulator ./cmd/lorawan-simulator

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/lorawan-simulator .

# Expose API port
EXPOSE 2208

# Run the application
CMD ["./lorawan-simulator"]
