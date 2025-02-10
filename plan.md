1. 11/02/2025 - reuse the same connection
2. collect the information about in which repo all the packages belong and coalesce
3. wait for coredns to start before upgrading
4. mandatory triple buffering
5. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
6. Ping and Pong control frames
7. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
8. fallback transports: comet long polling
9. exponential back off grpc client to kapetanios