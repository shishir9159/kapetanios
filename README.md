# kapetanios

## Docker Build
```Bash
docker build . -t quay.io/klovercloud/kapetanios:latest
docker push quay.io/klovercloud/kapetanios:latest
docker build -t quay.io/klovercloud/certs-renewal:latest -f ./internal/certs/Dockerfile .
docker push quay.io/klovercloud/certs-renewal:latest
```