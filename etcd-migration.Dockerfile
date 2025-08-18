#syntax=docker/dockerfile:1.7-labs
FROM golang:1.25.0 AS builder
WORKDIR /app

COPY . ./
RUN go mod download
RUN go build -C ./cmd/etcd-migration -o main

FROM ubuntu:latest
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt/,type=cache,sharing=locked \
    set -x && rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install -y \
    ca-certificates curl && \
    rm -rf /var/lib/apt/lists/* \

COPY --from=builder /app/cmd/etcd-migration/main /app/server
WORKDIR /app
CMD ["/app/server"]