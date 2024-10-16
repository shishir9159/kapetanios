FROM ubuntu:latest
WORKDIR /
# redirect errors
CMD ["/bin/bash -c chroot /host systemctl restart etcd"]