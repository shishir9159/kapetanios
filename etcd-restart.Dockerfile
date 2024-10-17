FROM ubuntu:latest
# TODO:
#  redirect errors and run the command from CMD
#  instead of relying on the scripts

CMD ["/usr/bin/bash", "-c", "chroot /host /usr/bin/systemctl restart etcd"]