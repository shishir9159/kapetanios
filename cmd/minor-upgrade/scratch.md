Calculate the Next Version
Download k8s components binaries for installation
Start to pull images on the necessary nodes
Synchronize kubernetes binaries
Upgrade cluster on master nodes
Get kubernetes cluster status
upgrade the worker nodes


1. kubectl drain node
2. apt-cache madison kubeadm | awk '{ print $3 }'
3. apt-mark unhold kubeadm && apt-get update && apt-get install -y kubeadm='1.27.5-1.1' && apt-mark hold kubeadm
4. kubeadm upgrade plan

