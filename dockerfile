FROM golang:1.23.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o tg-nmap-bot .

FROM alpine:3.20.2
RUN apk --no-cache add ca-certificates nmap

COPY --from=builder /app/tg-nmap-bot /usr/local/bin/tg-nmap-bot
COPY --from=builder /app/config.yml /etc/tg-nmap-bot/config.yml

RUN chmod +x /usr/local/bin/tg-nmap-bot


CMD ["tg-nmap-bot"]