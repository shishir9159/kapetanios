FROM golang:1.23 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download

COPY cmd/minor-upgrade config/ internal/ proto/ ./
RUN go build -C ./cmd/minor-upgrade -o main

FROM ubuntu:latest
RUN set -x && apt-get update && apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/cmd/minor-upgrade/main /app/server
WORKDIR /app
CMD ["/app/server"]