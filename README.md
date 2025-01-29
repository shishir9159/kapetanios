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
3. External Etcd Nodes: kubectl label node <node-name> assigned-node-role-etcd.kubernetes.io=etcd
4. the kubeadm config file location should be in the location /etc/kubernetes/kubeadm/kubeadm-config.yaml

## Protoc Code Generation
```Bash
make protgen
```
docker build . -t quay.io/klovercloud/kapetanios:latest; docker push quay.io/klovercloud/kapetanios:latest; docker build . -t quay.io/klovercloud/certs-expiration:latest -f certs-expiration.Dockerfile; docker push quay.io/klovercloud/certs-expiration:latest; docker build . -t quay.io/klovercloud/certs-renewal:latest -f certs-renewal.Dockerfile; docker push quay.io/klovercloud/certs-renewal:latest; docker build . -t quay.io/klovercloud/etcd-restart:latest -f etcd-restart.Dockerfile; docker push quay.io/klovercloud/etcd-restart:latest; docker build . -t quay.io/klovercloud/minor-upgrade:latest -f minor-upgrade.Dockerfile; docker push quay.io/klovercloud/minor-upgrade:latest; docker build . -t quay.io/klovercloud/rollback:latest -f rollback.Dockerfile; docker push quay.io/klovercloud/rollback:latest;
## Docker Build
```Bash
docker build . -t quay.io/klovercloud/kapetanios:latest; docker push quay.io/klovercloud/kapetanios:latest; docker build . -t quay.io/klovercloud/certs-expiration:latest -f certs-expiration.Dockerfile; docker push quay.io/klovercloud/certs-expiration:latest; docker build . -t quay.io/klovercloud/certs-renewal:latest -f certs-renewal.Dockerfile; docker push quay.io/klovercloud/certs-renewal:latest; docker build . -t quay.io/klovercloud/etcd-migration:latest -f etcd-migration.Dockerfile; docker push quay.io/klovercloud/etcd-migration:latest; docker build . -t quay.io/klovercloud/etcd-restart:latest -f etcd-restart.Dockerfile; docker push quay.io/klovercloud/etcd-restart:latest; docker build . -t quay.io/klovercloud/minor-upgrade:latest -f minor-upgrade.Dockerfile; docker push quay.io/klovercloud/minor-upgrade:latest; docker build . -t quay.io/klovercloud/rollback:latest -f rollback.Dockerfile; docker push quay.io/klovercloud/rollback:latest;
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