#syntax=docker/dockerfile:1.17.4.0-labs
FROM golang:1.25.1 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -C ./cmd/kapetanios -o main

# TODO: distroless image
FROM debian:bookworm-slim
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt/,type=cache,sharing=locked \
    set -x && rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install -y \
    ca-certificates curl && \
    rm -rf /var/lib/apt/lists/* \

COPY --from=builder /app/cmd/kapetanios/main /app/server
CMD ["/app/server"]