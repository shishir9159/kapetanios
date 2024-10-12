FROM golang:1.23 AS builder
WORKDIR /app

COPY . ./
RUN go mod download
RUN go build -C ./cmd/cert-renewal -o main

FROM ubuntu:latest
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/cmd/cert-renewal/main /app/server
WORKDIR /app
CMD ["/app/server"]

#FROM tarrunkhosla/grpcio:v1
#WORKDIR /
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
#COPY --from=builder /etc/passwd /etc/passwd
#COPY --from=builder /grpcurl /bin/grpcurl
#USER grpcurl
#
#ENTRYPOINT ["/bin/grpcurl"]