#syntax=docker/dockerfile:1.7-labs
FROM golang:1.23.6 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --parents cmd/transmission-control-over-wss config/ internal/ proto/ utils/ ./
RUN go build -C ./cmd/transmission-control-over-wss -o main

# the cache is mounted only
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt/,type=cache,sharing=locked \
    set -x && rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install -y curl \
    ca-certificates && rm -rf /var/lib/apt/lists/*

COPY ./cmd/transmission-control-over-wss/main ./server
RUN export PATH="$PATH:$(go env GOPATH)/bin"
#CMD ["sh", "-c", "tail -f /dev/null"]
CMD ["dlv", "--listen=:1234", "--headless=true", "--api-version=2", "exec", "./app/server"]