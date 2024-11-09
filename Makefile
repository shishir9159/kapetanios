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
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-expiration.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-upgrade.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rollback.proto
go-protogen:
	@printf $(COLOR) "Generating gRPC code after installing protoc with go install"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-expiration.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-upgrade.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rollback.proto

clean:
	@printf $(COLOR) "Removing built binaries with the build directory..."
	rm -rf ./build

##### Docker #####
.PHONY: kapetanios
docker-build-and-push: kapetanios certs-renewal certs-expiration etcd-restart rollback

kapetanios:
	@printf $(COLOR) "Building docker image for kapetanios and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/kapetanios:latest
	docker push quay.io/klovercloud/kapetanios:latest

.PHONY: certs-expiration
certs-expiration:
	@printf $(COLOR) "Building docker image for certs-expiration minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/certs-expiration:latest -f certs-expiration.Dockerfile
 	docker push quay.io/klovercloud/certs-expiration:latest

.PHONY: certs-renewal
certs-renewal:
	@printf $(COLOR) "Building docker image for certs-renewal minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/certs-renewal:latest -f certs-renewal.Dockerfile
 	docker push quay.io/klovercloud/certs-renewal:latest

.PHONY: etcd-migration
etcd-migration:
	@printf $(COLOR) "Building docker image for etcd-migration minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/etcd-migration:latest -f etcd-migration.Dockerfile
 	docker push quay.io/klovercloud/etcd-migration:latest

.PHONY: etcd-restart
etcd-restart:
	@printf $(COLOR) "Building docker image for etcd-restart minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/etcd-restart:latest -f etcd-restart.Dockerfile
 	docker push quay.io/klovercloud/etcd-restart:latest

.PHONY: minor-upgrade
minor-upgrade:
	@printf $(COLOR) "Building docker image for minor-upgrade minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/minor-upgrade:latest -f minor-upgrade.Dockerfile
	docker push quay.io/klovercloud/minor-upgrade:latest

.PHONY: rollback
rollback:
	@printf $(COLOR) "Building docker image for etcd-restart minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/rollback:latest -f rollback.Dockerfile
	docker push quay.io/klovercloud/rollback:latest

#### Deploy ####
deploy:
	@printf $(COLOR) "Deploying kapetanios in Kubernetes..."
	kubectl create -f manifests/.