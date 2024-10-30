Calculate the Next Version
Download k8s components binaries for installation
Start to pull images on the necessary nodes
Synchronize kubernetes binaries
Upgrade cluster on master nodes
Get kubernetes cluster status
upgrade the worker nodes

- [ ] kubectl drain node <node-to-drain>
- [ ] apt-cache madison kubeadm | awk '{ print $3 }'
- [ ] apt-mark unhold kubeadm && apt-get update && apt-get install -y kubeadm='1.26.5-1.1' && apt-mark hold kubeadm
- [ ] kubeadm upgrade plan -- not for the worker
- [ ] kubeadm upgrade apply v1.26.5 --certificate-renewal=false
- [ ] for other master nodes: kubeadm upgrade node --certificate-renewal=false
- [ ] kubectl drain <node-to-drain> --ignore-daemonsets -- recommended draining point
- [ ] apt-mark unhold kubelet kubectl && \
   apt-get update && apt-get install -y kubelet='1.26.5-1.1' kubectl='1.26.5-1.1' && \
   apt-mark hold kubelet kubectl
- [ ] systemctl daemon-reload
- [ ] systemctl restart kubelet
- [ ] kubectl uncordon <node-to-drain>