# kapetanios

#### Tested against:
ubuntu 22.04 LTS

#### build requirements:
golang v1.23.1 \
libprotoc 27.3

#### cluster requirement
It is assumed that the following roles would be assigned to the respective nodes:
1. Master Nodes: kubectl label node <node-name> assigned-node-role-certs.kubernetes.io=certs and "node-role.kubernetes.io/control-plane" annotation
2. assigned-node-role-minor-upgrade.kubernetes.io
2. External Etcd Nodes: kubectl label node <node-name> assigned-node-role-etcd.kubernetes.io=etcd
3. the kubeadm config file location should be in the location /etc/kubernetes/kubeadm/kubeadm-config.yaml

#### plugin install:
```Bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Protoc Code Generation
```Bash
export PATH="$PATH:$(go env GOPATH)/bin"
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/certs-renewal.proto
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

## Testing
### grpc
```Bash
grpcurl -v -plaintext kapetanios.default.svc.cluster.local:50051 Renewal/StatusUpdate{
  "backupSuccess" : true,
  "renewalSuccess" : true,
  "restartSuccess" : true
}
```

### Possible errors:

```
domain cluster.local
search cluster.local
nameserver 10.96.0.10 
```