#syntax=docker/dockerfile:1.7-labs
FROM golang:1.24.2 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY --parents cmd/transmission-control-over-wss config/ internal/ proto/ utils/ ./
RUN go build -C ./cmd/transmission-control-over-wss -o main

FROM debian:bookworm
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt/,type=cache,sharing=locked \
    set -x && rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install -y curl \
    ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/cmd/transmission-control-over-wss/main /app/server
CMD ["/app/server"]