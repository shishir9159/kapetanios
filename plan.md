1. pod eviction monitoring
2. certs expiration pod monitoring
3. packagize or internalize the expiration check functionality
4. repository pattern to switch between different persistent storage
5. refactor UI code
6. progression report under the node names
7. move the workflow inside pkg to share code
8. test -> handle error -> show error -> control plane components restarting -> refresh
9. investigate the error: writev: broken pipe
10. a standalone binary to run a specific command and show output
11. error + log wrapping around panicable(pinnacle) errors
12. should I make progressionState part of the upgradeState?
13. only one instance of minion pods in a node
14. when is the config file is necessary when it is already available in configmap that is right only from the perspective of the first node `--config /etc/kubernetes/kubeadm/kubeadm-config.yaml`.
15. do the sequential update with three nodes as blast radius
16. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
17. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
18. no retry logic from frontend for websocket connections
19. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
20. collect the information about in which repo all the packages belong and coalesce
21. wait for coredns to start before upgrading
22. mandatory triple buffering
23. guaranteed ordering and exactly once delivery(data integrity) must be maintained when connection is re-established
24. Ping and Pong control frames
25. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
26. fallback transports: comet-long polling
27. exponential back off grpc client to kapetanios
28. check what happens when certificate authority is missing