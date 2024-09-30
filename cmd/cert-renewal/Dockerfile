FROM golang:1.23-bookworm AS builder
WORKDIR /app

COPY go.* ./
RUN rm -rf go.mod
COPY ./cmd/cert-renewal/go.mod .
RUN go mod download
COPY . ./
RUN go build -C ./cmd/cert-renewal -o main

FROM ubuntu:latest
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/cmd/cert-renewal/main /app/server

CMD ["/app/server"]