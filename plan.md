1. refactor UI code
2. rename rollback to certs-rollback
3. progression report under the node names
4. move the workflow inside pkg to share code
5. stale dependency list: https://github.com/go-yaml/yaml
6. test -> handle error -> show error -> control plane components restarting -> refresh
7. investigate the error: writev: broken pipe
8. a standalone binary to run a specific command and show output
9. error + log wrapping around panicable(pinnacle) errors
10. should I make progressionState part of the upgradeState?
11. only one instance of minion pods in a node
12. do the sequential update with three nodes as blast radius
13. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
14. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
15. no retry logic from frontend for websocket connections
16. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
17. collect the information about in which repo all the packages belong and coalesce
18. wait for coredns to start before upgrading
19. mandatory triple buffering
20. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
21. Ping and Pong control frames
22. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
23. fallback transports: comet long polling
24. exponential back off grpc client to kapetanios