FROM golang:1.22.6 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o tg-nmap-bot .

FROM debian:12.7

EXPOSE 12345

RUN apt update && apt install ca-certificates nmap -y

COPY --from=builder /app/tg-nmap-bot /usr/local/bin/tg-nmap-bot
COPY --from=builder /app/config.yml /etc/tg-nmap-bot/config.yml

RUN chmod +x /usr/local/bin/tg-nmap-bot


CMD ["tg-nmap-bot"]
