# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO
RUN apk add --no-cache \
    build-base \
    pkgconf \
    postgresql-dev

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

# Build the application
RUN --mount=type=cache,target=/go/pkg/mod \
    go build -v -o axiomnizam .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache \
    ca-certificates \
    postgresql-libs

# Copy binary from builder
COPY --from=builder /app/axiomnizam .

# Copy environment file if exists
COPY --from=builder /app/.env* ./

EXPOSE 8000

CMD ["./axiomnizam"]
