# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    postgresql-client \
    libpq-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set Go environment variables
ENV GOFLAGS=-mod=mod
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=off
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# Copy go.mod first for better layer caching
COPY go.mod go.sum* ./

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Run go mod tidy to ensure consistency
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod tidy

# Build the application - no verbose to reduce output
RUN --mount=type=cache,target=/go/pkg/mod \
    go build -o axiomnizam . 2>&1 || (echo "Build failed with exit code $?" && exit 1)

# Build CLI tool
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o axiomnizamctl ./cmd/axiomnizamctl 2>&1 || (echo "CLI build failed with exit code $?" && exit 1)

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /root/

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

# Copy binaries from builder
COPY --from=builder /app/axiomnizam .
COPY --from=builder /app/axiomnizamctl /usr/local/bin/axiomnizamctl

# Copy environment file if exists
COPY --from=builder /app/.env* ./

# Make CLI available globally
RUN chmod +x /usr/local/bin/axiomnizamctl && ln -sf /usr/local/bin/axiomnizamctl /usr/local/bin/axctl

EXPOSE 8000

CMD ["./axiomnizam"]
