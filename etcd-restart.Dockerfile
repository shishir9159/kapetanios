FROM ubuntu:latest
# redirect errors
CMD ["/bin/bash -c chroot /host systemctl restart etcd"]