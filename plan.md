1. a standalone binary to run a specific command and show output
2. error + log wrapping around panicable(pinnacle) errors
3. should I make progressionState part of the upgradeState?
4. only one instance of minion pods in a node
5. do the sequential update with three nodes as blast radius
6. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
7. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
8. no retry logic from frontend for websocket connections
9. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
10. collect the information about in which repo all the packages belong and coalesce
11. wait for coredns to start before upgrading
12. mandatory triple buffering
13. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
14. Ping and Pong control frames
15. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
16. fallback transports: comet long polling
17. exponential back off grpc client to kapetanios