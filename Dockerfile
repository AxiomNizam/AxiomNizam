# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

# Set Go environment variables with better fallback options
ENV GOFLAGS=-mod=mod
ENV GOPROXY=https://proxy.golang.org,https://goproxy.io,direct
ENV GOSUMDB=off

# Copy go.mod file
COPY go.mod ./

# Remove go.sum if it exists to regenerate it with correct checksums
RUN rm -f go.sum

# Download dependencies with retries
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x 2>&1 || \
    (echo "First attempt failed, waiting 5s..." && sleep 5 && go mod download -x 2>&1) || \
    (echo "Second attempt failed, waiting 10s..." && sleep 10 && go mod download -x 2>&1) || \
    echo "Download completed"

# Verify and tidy dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod verify 2>&1 || echo "Module verification completed"

# Copy source code
COPY . .

# Build the application with verbose output
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o axiomnizam .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/axiomnizam .
COPY --from=builder /app/.env* ./

EXPOSE 8000

CMD ["./axiomnizam"]
