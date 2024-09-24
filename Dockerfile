FROM golang:1.23-bookworm AS builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY . ./
RUN go build ./cmd/main.go

FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main /app/server
CMD ["/app/server"]