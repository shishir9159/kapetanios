# kapetanios

#### Tested against:
ubuntu 22.04 LTS
ubuntu 24.04 LTS
RHEL 8
RHEL 9

#### build requirements:
golang v1.23.1 \
libprotoc 27.3

#### cluster requirement
It is assumed that the following roles would be assigned to the respective nodes:
1. assumption about package naming format
2. assumption is that all the components(kubeadm, kubelet, kubectl) has the same specific version
3. Master Nodes: kubectl label node <node-name> assigned-node-role-certs.kubernetes.io=certs and "node-role.kubernetes.io/control-plane" annotation
4. assigned-node-role-minor-upgrade.kubernetes.io=minor-upgrade
5. External Etcd Nodes: kubectl label node <node-name> assigned-node-role-etcd.kubernetes.io=etcd
6. the kubeadm config file location should be in the location /etc/kubernetes/kubeadm/kubeadm-config.yaml

## Protoc Code Generation
```Bash
make protgen
```
## Docker Build and Push
```Bash
# there's some bugs in the current format
make docker-build-and-push
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

## Stale Dependency
```textmate
https://github.com/go-yaml/yaml
```