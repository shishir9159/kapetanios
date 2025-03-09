#syntax=docker/dockerfile:1.7-labs
FROM golang:1.23.6 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY --parents cmd/minor-update config/ internal/ proto/ utils/ ./

RUN go build -C ./cmd/minor-update -o main

FROM debian:bookworm-slim
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt/,type=cache,sharing=locked \
    set -x && apt-get update && apt-get install -y \
    ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/cmd/minor-update/main /app/server
WORKDIR /app
CMD ["/app/server"]