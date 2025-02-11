1. 11/02/2025 - reuse the same connection - failed: connection pool it is
2. 11/02/2025 - implement connection pool that handles states and client gracefully
3. collect the information about in which repo all the packages belong and coalesce
4. wait for coredns to start before upgrading
5. mandatory triple buffering
6. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
7. Ping and Pong control frames
8. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
9. fallback transports: comet long polling
10. exponential back off grpc client to kapetanios