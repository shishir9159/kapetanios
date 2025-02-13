1. 11/02/2025 - reuse the same connection - failed: connection pool it is
2. 11/02/2025 - implement connection pool that handles states and client gracefully
3. no retry logic from frontend for websocket connections
4. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
5. collect the information about in which repo all the packages belong and coalesce
6. wait for coredns to start before upgrading
7. mandatory triple buffering
8. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
9. Ping and Pong control frames
10. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
11. fallback transports: comet long polling
12. exponential back off grpc client to kapetanios