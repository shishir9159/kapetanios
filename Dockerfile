FROM golang:1.23 AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -C ./cmd/kapetanios -o main

FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/cmd/kapetanios/main /app/server
CMD ["/app/server"]