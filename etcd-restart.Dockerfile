# syntax=docker.io/docker/dockerfile-upstream:1.14.0-labs
FROM debian:bookworm-slim
# TODO:
#   redirect errors and run the command from CMD
#   instead of relying on the scripts

CMD ["/usr/bin/bash", "-c", "chroot /host /usr/bin/systemctl restart etcd"]