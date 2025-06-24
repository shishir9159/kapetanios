# kapetanios

#### Tested against:
ubuntu 22.04 LTS
ubuntu 24.04 LTS
RHEL 8.0-8.3
RHEL 9.0-9.4

#### build requirements:
golang v1.23.2-v1.24.4
libprotoc 27.3-30.2

[//]: # (make)

#### cluster requirement
It is assumed that the following roles would be assigned to the respective nodes:
1. assumption about package naming format
2. the assumption is that all the components(kubeadm, kubelet, kubectl) have the same specific version in the os
3. remove depenedancy: Master Nodes: kubectl label node <node-name> assigned-node-role-certs.kubernetes.io=certs
4. remove depenedancy: assigned-node-role-minor-upgrade.kubernetes.io=minor-upgrade
5. remove depenedancy: External Etcd Nodes: kubectl label node <node-name> assigned-node-role-etcd.kubernetes.io=etcd
6. the kubeadm config file location should be in the location /etc/kubernetes/kubeadm/kubeadm-config.yaml

## Notes    
As a master node role is deprecated, only control-plane annotation is used: https://kubernetes.io/docs/reference/labels-annotations-taints/#node-role-kubernetes-io-master-taint

## Protoc Code Generation
```Zsh
make protgen
```
## Docker images Build and Push
```Zsh
make docker-build-and-push
```

## Deploy in K8s
```Zsh
kubectl create -f manifests/.
```

## Testing
### grpc
```Zsh
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

## Concerns
updating a node can take a long time, how to determine the application is running the process?

when is the config file is necessary when it is already available in configmap that is right only from the perspective of the first node `--config /etc/kubernetes/kubeadm/kubeadm-config.yaml`.