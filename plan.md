1. package install version
2. do the sequential update with three nodes as blast radius
3. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
4. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
5. no retry logic from frontend for websocket connections
6. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
7. collect the information about in which repo all the packages belong and coalesce
8. wait for coredns to start before upgrading
9. mandatory triple buffering
10. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
11. Ping and Pong control frames
12. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
13. fallback transports: comet long polling
14. exponential back off grpc client to kapetanios