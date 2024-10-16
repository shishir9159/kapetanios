FROM debian:bookworm-slim
COPY scripts/etcd-restart.sh .
# redirect errors
CMD ["./etcd-restart.sh"]