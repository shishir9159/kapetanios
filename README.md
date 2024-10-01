# kapetanios

Tested against: ubuntu 22.04 LTS

build requirements:
golang v1.23.1

libprotoc 27.3

plugin install:
```Bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Protoc Code Generation
```Bash
export PATH="$PATH:$(go env GOPATH)/bin"
protoc --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative proto/cert-renewal.proto
```

## Docker Build
```Bash
docker build . -t quay.io/klovercloud/kapetanios:latest
docker push quay.io/klovercloud/kapetanios:latest
docker build -t quay.io/klovercloud/certs-renewal:latest -f certs-renewal.Dockerfile .
docker push quay.io/klovercloud/certs-renewal:latest
```

## Deploy in K8s
```Bash
kubectl create -f manifests/.
```