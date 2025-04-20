##### Variables ######
GOOS := linux
CGO_ENABLED ?= 0
GOTELEMETRY='local'
GOMOD := '/dev/null'
NATIVE_ARCH := amd64
GOVERSION := 'go1.24.2'
DOCKER ?= docker buildx
COLOR := "\e[1;36m%s\e[0m\n"

##### Scripts ######
update: go-sqlc go-protogen

go-update:
	@printf $(COLOR) "Updating go.lock from go.mod and listing all upgradable stale direct dependencies"
	go list -m -u -json all | jq -r 'select(.Indirect == false) | .Path'

go-sqlc:
	@printf $(COLOR) "Generating sql boilerplate code after installing sqlc with go install"
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	sqlc generate -f ./sqlc/db/postgres.yaml

protogen:
	@printf $(COLOR) "Generating gRPC code"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rollback.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-expiration.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-update.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-upgrade.proto

go-protogen:
	@printf $(COLOR) "Generating gRPC code after installing protoc with go install"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rollback.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-expiration.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-update.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-upgrade.proto

clean:
	@printf $(COLOR) "Removing built binaries with the build directory..."
	rm -rf ./build

##### Docker #####
.PHONY: kapetanios
docker-build-and-push: kapetanios etcd-restart certs-expiration certs-renewal minor-update minor-upgrade rollback svelte

kapetanios:
	@printf $(COLOR) "Building docker image for kapetanios and pushing it to the registry..."
	#docker build . -t quay.io/klovercloud/kapetanios:latest
	docker build . -t quay.io/klovercloud/kapetanios:latest -f transmission-control-over-wss.Dockerfile
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

.PHONY: minor-update
minor-update:
	@printf $(COLOR) "Building docker image for minor-update minions and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/minor-update:latest -f minor-update.Dockerfile
	docker push quay.io/klovercloud/minor-update:latest

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

.PHONY: svelte
svelte:
	@printf $(COLOR) "Building docker image for kapetanios frontend and pushing it to the registry..."
	docker build . -t quay.io/klovercloud/svelte:latest -f svelte.Dockerfile
	docker push quay.io/klovercloud/svelte:latest

#### Deploy ####
deploy:
	@printf $(COLOR) "Deploying kapetanios in Kubernetes..."
	kubectl create -f manifests/.