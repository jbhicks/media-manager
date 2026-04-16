# Media Manager - VPN Required
FROM golang:1.21-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o media-manager-service ./cmd/media-manager-service

FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    curl \
    bash \
    iptables \
    ip6tables

# Install NordVPN
RUN curl -fsSL https://nordvpn.com/download/linux/install.sh | sh

WORKDIR /app

COPY --from=builder /build/media-manager-service .
COPY scripts/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN mkdir -p /data /mnt/media

ENV NORDVPN_TOKEN=""
ENV NORDVPN_SERVER=""
ENV DB_PATH="/data/media.db"
ENV DOWNLOAD_PATH="/mnt/media/Downloads"
ENV LIBRARY_PATH="/mnt/media/Movies"

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
CMD ["./media-manager-service"]
