Calculate the Next Version
Download k8s components binaries for installation
Start to pull images on the necessary nodes
Synchronize kubernetes binaries
Upgrade cluster on master nodes
Get kubernetes cluster status
upgrade the worker nodes


1. kubectl drain node <node-to-drain>
2. apt-cache madison kubeadm | awk '{ print $3 }'
3. apt-mark unhold kubeadm && apt-get update && apt-get install -y kubeadm='1.26.5-1.1' && apt-mark hold kubeadm
4. kubeadm upgrade plan
5. kubeadm upgrade apply v1.26.5 --certificate-renewal=false
6. for other master nodes: kubeadm upgrade node v1.26.5 --certificate-renewal=false
7. kubectl drain <node-to-drain> --ignore-daemonsets
8. apt-mark unhold kubelet kubectl && \
   apt-get update && apt-get install -y kubelet='1.26.5-1.1' kubectl='1.26.5-1.1' && \
   apt-mark hold kubelet kubectl

steps are same for the worker nodes

