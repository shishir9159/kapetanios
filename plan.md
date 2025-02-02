1. collect the information about in which repo all the packages belong and coalesce
2. wait for coredns to start before upgrading
3. mandatory triple buffering
4. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
5. Ping and Pong control frames
6. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
7. fallback transports: comet long polling
8. 