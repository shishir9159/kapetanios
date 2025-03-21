export const progressInfo: { stepCount: number; title: string; log: string; error: string }[] =
	$state([
		{
			stepCount: 1,
			title: 'upgrade prerequisites check',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 2,
			title: 'checking if the package version exists',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 3,
			title: 'kubernetes version upgrade diff',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 4,
			title: 'Kubeadm update',
			log: 'still processing',
			error: ''
		},

		// TODO:
		//  optional step
		{
			stepCount: 5,
			title: 'Cluster Upgrade Plan',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 6,
			title: '',
			log: 'Kubernetes Cluster Upgrade',
			error: ''
		},

		{
			stepCount: 7,
			title: 'Kubelet update',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 8,
			title: 'Restarting Kubelet',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 9,
			title: 'Kubectl update',
			log: 'still processing',
			error: ''
		}
	]);
