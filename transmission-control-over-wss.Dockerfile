# syntax=docker.io/docker/dockerfile-upstream:1.14.0-labs
FROM golang:1.25.1 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY --parents cmd/transmission-control-over-wss config/ internal/ proto/ utils/ ./
RUN go build -C ./cmd/transmission-control-over-wss -o main

FROM ubuntu:latest
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt/,type=cache,sharing=locked \
    set -x && rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install -y \
    ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/cmd/transmission-control-over-wss/main /app/server
CMD ["/app/server"]