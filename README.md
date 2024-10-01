# kapetanios

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