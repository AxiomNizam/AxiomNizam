# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

ENV GOFLAGS=-mod=mod

COPY go.mod ./
RUN go mod tidy
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o hello .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/hello .

CMD ["./hello"]
