1. package install version
2. repository the package belongs to
3. do the sequential update with three nodes as blast radius
4. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
5. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
6. no retry logic from frontend for websocket connections
7. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
8. collect the information about in which repo all the packages belong and coalesce
9. wait for coredns to start before upgrading
10. mandatory triple buffering
11. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
12. Ping and Pong control frames
13. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
14. fallback transports: comet long polling
15. exponential back off grpc client to kapetanios