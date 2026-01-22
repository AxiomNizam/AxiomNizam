# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

ENV GOFLAGS=-mod=mod
ENV GOPROXY=https://proxy.golang.org,direct

COPY go.mod ./
RUN go mod download || (sleep 5 && go mod download)

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o axiomnizam .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/axiomnizam .
COPY --from=builder /app/.env* ./

EXPOSE 8000

CMD ["./axiomnizam"]
