1. packagize or internalize the expiration check functionality: holding on till the merge
2. progression report under the node names
3. pod eviction monitoring
4. repository pattern to switch between different persistent storage
5. refactor UI code
6. move the workflow inside pkg to share code
7. test -> handle error -> show error -> control plane components restarting -> refresh
8. investigate the error: writev: broken pipe
9. a standalone binary to run a specific command and show output
10. error + log wrapping around panicable(pinnacle) errors
11. should I make progressionState part of the upgradeState?
12. only one instance of minion pods in a node
13. when is the config file is necessary when it is already available in configmap that is right only from the perspective of the first node `--config /etc/kubernetes/kubeadm/kubeadm-config.yaml`.
14. do the sequential update with three nodes as blast radius
15. if some granular action is taking place/processing, it can not be stopped. As it has already run on the VM and the additional information scrapping is on the process. maybe next update command can be canceled??? cmd.Process.Kill() or exec.CommandContext(ctx, "sleep", "5").Run()?
16. always reading input or bust of inputs? ping pong implementation? context defer logic to remove the spawned containers?
17. no retry logic from frontend for websocket connections
18. need to monitor successful kubelet restarts https://pkg.go.dev/github.com/chimeracoder/journalctl-go
19. collect the information about in which repo all the packages belong and coalesce
20. wait for coredns to start before upgrading
21. mandatory triple buffering
22. guaranteed ordering and exactly once delivery(data integrity) must be maintained when connection is re-established
23. Ping and Pong control frames
24. monitor the buffers building up on the sockets used to stream data to clients and ensure a buffer never grows beyond what the downstream connection can sustain by handling back pressure
25. fallback transports: comet-long polling
26. exponential back off grpc client to kapetanios
27. check what happens when certificate authority is missing