FROM quay.io/klovercloud/golang:1.20 as builder

WORKDIR /app

COPY . ./
RUN go mod download
RUN go build -C ./cmd/cert-renewal -o /app/server

#FROM ubuntu:latest
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
#COPY /app/cmd/cert-renewal/main /app/server
CMD ["/app/server"]