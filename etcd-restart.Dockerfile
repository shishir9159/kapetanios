FROM debian:bookworm-slim
RUN apt update -y
# redirect errors
CMD ["/bin/sh ","-c", "chroot /host /usr/bin/systemctl restart etcd"]