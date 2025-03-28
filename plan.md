1. move the workflow inside pkg to share code
2. test -> handle error -> show error -> control plane components restarting -> refresh
3. rename rollback to certs-rollback
4. investigate the error: writev: broken pipe
5. a standalone binary to run a specific command and show output
6. error + log wrapping around panicable(pinnacle) errors
7. should I make progressionState part of the upgradeState?
8. only one instance of minion pods in a node
9. do the sequential update with three nodes as blast radius
10. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
11. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
12. no retry logic from frontend for websocket connections
13. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
14. collect the information about in which repo all the packages belong and coalesce
15. wait for coredns to start before upgrading
16. mandatory triple buffering
17. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
18. Ping and Pong control frames
19. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
20. fallback transports: comet long polling
21. exponential back off grpc client to kapetanios