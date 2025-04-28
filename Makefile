##### Variables ######
COLOR := "\e[1;36m%s\e[0m\n"
CGO_ENABLED ?= 0
DOCKER ?= docker buildx
GOMOD := '/dev/null'
GOOS := linux
GOTELEMETRY='local'
GOVERSION := 'go1.24.2'
MAKEFLAGS += -j$(shell grep -c 'processor' /proc/cpuinfo)
NATIVE_ARCH := amd64

##### Scripts ######
update: go-sqlc go-protogen

go-update:
	@printf $(COLOR) "Updating go.lock from go.mod and listing all upgradable stale direct dependencies"
	@printf $(COLOR) "direct dependencies"
	go list -u -m -json all | jq -r 'select(.Update) | "\(.Path) (\(.Version) → \(.Update.Version))"'
	@printf $(COLOR) "indirect dependencies"
	go list -u -m -json all | jq -r 'select(.Update and (.Indirect != true)) | "\(.Path) (\(.Version) → \(.Update.Version))"'

go-sqlc:
	@printf $(COLOR) "Generating sql boilerplate code after installing sqlc with go install"
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	sqlc generate -f ./sqlc/db/postgres.yaml

protogen:
	@printf $(COLOR) "Generating gRPC code"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-rollback.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-expiration.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-update.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-upgrade.proto

go-protogen:
	@printf $(COLOR) "Generating gRPC code after installing protoc with go install"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-rollback.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-expiration.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-update.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/minor-upgrade.proto

clean:
	@printf $(COLOR) "Removing cached docker images and removing the kubernetes resources..."
#   remove the node labelings
	docker system prune -ya --volumes
	kubectl create -f manifests/.
	#rm -rf ./build

##### Docker #####
.PHONY: kapetanios
docker-build-and-push: kapetanios etcd-restart certs-expiration certs-renewal minor-update minor-upgrade certs-rollback svelte

kapetanios:
	@printf $(COLOR) "Building docker image for kapetanios and pushing it to the registry..."
	#docker build . -t laplaciandemon/kapetanios:latest
	docker build . -t laplaciandemon/kapetanios:latest -f transmission-control-over-wss.Dockerfile
	docker push laplaciandemon/kapetanios:latest

.PHONY: certs-expiration
certs-expiration:
	@printf $(COLOR) "Building docker image for certs-expiration minions and pushing it to the registry..."
	docker build . -t laplaciandemon/certs-expiration:latest -f certs-expiration.Dockerfile
	docker push laplaciandemon/certs-expiration:latest

.PHONY: certs-renewal
certs-renewal:
	@printf $(COLOR) "Building docker image for certs-renewal minions and pushing it to the registry..."
	docker build . -t laplaciandemon/certs-renewal:latest -f certs-renewal.Dockerfile
	docker push laplaciandemon/certs-renewal:latest

.PHONY: etcd-migration
etcd-migration:
	@printf $(COLOR) "Building docker image for etcd-migration minions and pushing it to the registry..."
	docker build . -t laplaciandemon/etcd-migration:latest -f etcd-migration.Dockerfile
	docker push laplaciandemon/etcd-migration:latest

.PHONY: etcd-restart
etcd-restart:
	@printf $(COLOR) "Building docker image for etcd-restart minions and pushing it to the registry..."
	docker build . -t laplaciandemon/etcd-restart:latest -f etcd-restart.Dockerfile
	docker push laplaciandemon/etcd-restart:latest

.PHONY: minor-update
minor-update:
	@printf $(COLOR) "Building docker image for minor-update minions and pushing it to the registry..."
	docker build . -t laplaciandemon/minor-update:latest -f minor-update.Dockerfile
	docker push laplaciandemon/minor-update:latest

.PHONY: minor-upgrade
minor-upgrade:
	@printf $(COLOR) "Building docker image for minor-upgrade minions and pushing it to the registry..."
	docker build . -t laplaciandemon/minor-upgrade:latest -f minor-upgrade.Dockerfile
	docker push laplaciandemon/minor-upgrade:latest

.PHONY: certs-rollback
certs-rollback:
	@printf $(COLOR) "Building docker image for etcd-restart minions and pushing it to the registry..."
	docker build . -t laplaciandemon/certs-rollback:latest -f certs-rollback.Dockerfile
	docker push laplaciandemon/certs-rollback:latest

.PHONY: svelte
svelte:
	@printf $(COLOR) "Building docker image for kapetanios frontend and pushing it to the registry..."
	docker build . -t laplaciandemon/kapetanios-svelte:latest -f svelte.Dockerfile
	docker push laplaciandemon/kapetanios-svelte:latest

#### Deploy ####
.PHONY: deploy
deploy:
	@printf $(COLOR) "Deploying kapetanios in Kubernetes..."
	kubectl create -f manifests/.