FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . . 

RUN go build -o url-shortener ./cmd/url-shortener/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-shortener .
COPY ./internal/config ./config
COPY ./tests ./tests

EXPOSE 8082

RUN apk add libc6-compat

CMD ["./url-shortener"]