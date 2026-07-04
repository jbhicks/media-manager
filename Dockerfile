FROM golang:1.24-bookworm AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y git build-essential && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o media-manager-service ./cmd/media-manager-service

FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    iproute2 \
    bash \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/media-manager-service .

RUN mkdir -p /data /mnt/media

ENV DB_PATH="/data/media.db"
ENV DOWNLOAD_PATH="/mnt/media/Downloads"
ENV LIBRARY_PATH="/mnt/media/Movies"
ENV PORT="8080"

EXPOSE 8080

CMD ["./media-manager-service"]
