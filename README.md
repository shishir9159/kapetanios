# kapetanios

## Docker Build
```Bash
docker build . -t quay.io/klovercloud/kapetanios:latest
docker push quay.io/klovercloud/kapetanios:latest
docker build ./internal/certs/Dockerfile -t quay.io/klovercloud/certs-renewal:latest
docker push quay.io/klovercloud/certs-renewal:latest
```