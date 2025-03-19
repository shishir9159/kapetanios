1. investigate the error: writev: broken pipe
2. a standalone binary to run a specific command and show output
3. error + log wrapping around panicable(pinnacle) errors
4. should I make progressionState part of the upgradeState?
5. only one instance of minion pods in a node
6. do the sequential update with three nodes as blast radius
7. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
8. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
9. no retry logic from frontend for websocket connections
10. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
11. collect the information about in which repo all the packages belong and coalesce
12. wait for coredns to start before upgrading
13. mandatory triple buffering
14. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
15. Ping and Pong control frames
16. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
17. fallback transports: comet long polling
18. exponential back off grpc client to kapetanios