1. 11/02/2025 - reuse the same connection - failed: connection pool it is
2. 11/02/2025 - implement connection pool that handles states and client gracefully
3. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned contaienrs?
4. no retry logic from frontend for websocket connections
5. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
6. collect the information about in which repo all the packages belong and coalesce
7. wait for coredns to start before upgrading
8. mandatory triple buffering
9. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
10. Ping and Pong control frames
11. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
12. fallback transports: comet long polling
13. exponential back off grpc client to kapetanios