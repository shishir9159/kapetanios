1. enable true retry from the client side
2. test -> handle error -> show error -> control plane components restarting -> refresh
3. investigate the error: writev: broken pipe
4. a standalone binary to run a specific command and show output
5. error + log wrapping around panicable(pinnacle) errors
6. should I make progressionState part of the upgradeState?
7. only one instance of minion pods in a node
8. do the sequential update with three nodes as blast radius
9. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
10. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
11. no retry logic from frontend for websocket connections
12. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
13. collect the information about in which repo all the packages belong and coalesce
14. wait for coredns to start before upgrading
15. mandatory triple buffering
16. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
17. Ping and Pong control frames
18. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
19. fallback transports: comet long polling
20. exponential back off grpc client to kapetanios