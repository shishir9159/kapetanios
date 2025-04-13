1. employ the following in deployment and use it instead of querying
       env:
         - name: CURRENT_NODE_NAME
           valueFrom:
             fieldRef:
               fieldPath: spec.nodeName
2. refactor UI code
3. rename rollback to certs-rollback
4. progression report under the node names
5. move the workflow inside pkg to share code
6. stale dependency list: https://github.com/go-yaml/yaml
7. test -> handle error -> show error -> control plane components restarting -> refresh
8. investigate the error: writev: broken pipe
9. a standalone binary to run a specific command and show output
10. error + log wrapping around panicable(pinnacle) errors
11. should I make progressionState part of the upgradeState?
12. only one instance of minion pods in a node
13. do the sequential update with three nodes as blast radius
14. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
15. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
16. no retry logic from frontend for websocket connections
17. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
18. collect the information about in which repo all the packages belong and coalesce
19. wait for coredns to start before upgrading
20. mandatory triple buffering
21. guaranteed ordering and exactly-once delivery(data integrity) must be maintained when connection is re-established
22. Ping and Pong control frames
23. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
24. fallback transports: comet long polling
25. exponential back off grpc client to kapetanios