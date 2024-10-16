FROM ubuntu:latest
#WORKDIR /
#RUN apt update -y && apt install -y systemd
# redirect errors
CMD ["/bin/bash -c chroot /host /usr/bin/systemctl restart etcd"]