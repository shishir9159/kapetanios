FROM golang:1.23 AS builder
WORKDIR /app

COPY . ./
RUN go mod download
RUN go build -C ./cmd/rollback -o main

FROM ubuntu:latest
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/cmd/rollback/main /app/server
WORKDIR /app
CMD ["/app/server"]