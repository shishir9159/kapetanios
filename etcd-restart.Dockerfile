FROM debian:bookworm-slim
COPY scripts/etcd-restart.sh .
RUN chmod +x etcd-restart.sh
# TODO:
#  redirect errors and run the command from CMD
#  instead of relying on the scripts
CMD ["./etcd-restart.sh"]