FROM ubuntu:latest
COPY scripts/etcd-restart.sh .
RUN chmod +x etcd-restart.sh
# TODO:
#  redirect errors and run the command from CMD
#  instead of relying on the scripts

CMD ["/usr/bin/bash", "-c", "chroot /host /usr/bin/systemctl restart etcd"]