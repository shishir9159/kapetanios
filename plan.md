1. test -> handle error -> show error -> control plane components restarting -> refresh
2. investigate the error: writev: broken pipe
3. a standalone binary to run a specific command and show output
4. error + log wrapping around panicable(pinnacle) errors
5. should I make progressionState part of the upgradeState?
6. only one instance of minion pods in a node
7. do the sequential update with three nodes as blast radius
8. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
9. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
10. no retry logic from frontend for websocket connections
11. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
12. collect the information about in which repo all the packages belong and coalesce
13. wait for coredns to start before upgrading
14. mandatory triple buffering
15. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
16. Ping and Pong control frames
17. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
18. fallback transports: comet long polling
19. exponential back off grpc client to kapetanios