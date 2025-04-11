1. rename rollback to certs-rollback
2. progression report under the node names
3. move the workflow inside pkg to share code
4. test -> handle error -> show error -> control plane components restarting -> refresh
5. investigate the error: writev: broken pipe
6. a standalone binary to run a specific command and show output
7. error + log wrapping around panicable(pinnacle) errors
8. should I make progressionState part of the upgradeState?
9. only one instance of minion pods in a node
10. do the sequential update with three nodes as blast radius
11. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
12. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
13. no retry logic from frontend for websocket connections
14. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
15. collect the information about in which repo all the packages belong and coalesce
16. wait for coredns to start before upgrading
17. mandatory triple buffering
18. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
19. Ping and Pong control frames
20. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
21. fallback transports: comet long polling
22. exponential back off grpc client to kapetanios