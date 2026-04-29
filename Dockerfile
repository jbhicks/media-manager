# Media Manager - VPN Required
FROM golang:1.24-bookworm AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y git build-essential && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o media-manager-service ./cmd/media-manager-service

FROM ubuntu:24.04

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install base packages
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    wget \
    bash \
    iptables \
    iproute2 \
    net-tools \
    && rm -rf /var/lib/apt/lists/*

# Configure apt to never prompt
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/90forceyes && \
    echo 'APT::Get::allow-downgrades "true";' >> /etc/apt/apt.conf.d/90forceyes && \
    echo 'DPkg::Options {"--force-confdef";"--force-confold";}' >> /etc/apt/apt.conf.d/90forceyes && \
    echo 'Acquire::Retries "3";' >> /etc/apt/apt.conf.d/90forceyes

# Install NordVPN using their official install script
RUN curl -fsSL https://downloads.nordcdn.com/apps/linux/install.sh | bash || \
    (wget -qO- https://downloads.nordcdn.com/apps/linux/install.sh | bash) || \
    echo "NordVPN install failed, continuing without VPN"

# Install NordVPN package if repo was added
RUN apt-get update && \
    (apt-get install -y nordvpn || echo "NordVPN package install failed")

WORKDIR /app

COPY --from=builder /build/media-manager-service .
COPY web/dist /home/josh/media-manager/web/dist
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
