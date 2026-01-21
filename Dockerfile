# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o hello .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/hello .

CMD ["./hello"]
