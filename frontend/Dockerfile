FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go mod tidy
RUN go mod download

RUN go build -o frontend .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/frontend .
COPY --from=builder /app/templates ./templates

EXPOSE 7000

ENV BACKEND_URL=http://axiomnizam:8000
ENV FRONTEND_PORT=7000

CMD ["./frontend"]
