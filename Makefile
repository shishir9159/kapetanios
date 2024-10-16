##### Variables ######
COLOR := "\e[1;36m%s\e[0m\n"

CGO_ENABLED ?= 0
DOCKER ?= docker buildx
NATIVE_ARCH := amd64
GOMOD := '/dev/null'
GOOS := linux
GOTELEMETRY='local'
GOVERSION := 'go1.23.1'

##### Scripts ######
protogen:
	@printf $(COLOR) "Generating gRPC code"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/cert-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rollback.proto
go-protogen:
	@printf $(COLOR) "Generating gRPC code after installing protoc with go install"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/cert-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rollback.proto;
clean:
	@printf $(COLOR) "Removing built binaries with the build directory..."
	rm -rf ./build

##### Docker #####
.PHONY: kapetanios
docker-build-&-push: kapetanios cert-renewal etcd-restart rollback

kapetanios:
	@printf $(COLOR) "Building docker image for kapetanios and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/kapetanios:latest
	docker push quay.io/klovercloud/kapetanios:latest

.PHONY: cert-renewal
cert-renewal:
	@printf $(COLOR) "Building docker image for cert-renewal minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/certs-renewal:latest -f certs-renewal.Dockerfile
 	docker push quay.io/klovercloud/certs-renewal:latest

.PHONY: etcd-restart
etcd-restart:
	@printf $(COLOR) "Building docker image for etcd-restart minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/etcd-restart:latest -f etcd-restart.Dockerfile
 	docker push quay.io/klovercloud/etcd-restart:latest

.PHONY: rollback
rollback:
	@printf $(COLOR) "Building docker image for etcd-restart minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/rollback:latest -f rollback.Dockerfile
	docker push quay.io/klovercloud/rollback:latest

#### Deploy ####
deploy:
	@printf $(COLOR) "Deploying kapetanios in Kubernetes..."
	kubectl create -f manifests/.