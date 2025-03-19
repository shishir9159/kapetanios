export const progressInfo: { stepCount: number; title: string; log: string; error: string }[] =
	$state([
		{
			stepCount: 1,
			title: 'update progress',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 2,
			title: 'upgrade configuration selection',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 3,
			title: 'Cluster Health Checking',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 4,
			title: 'package availability check',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 5,
			title: 'Cluster Compatibility',
			log: 'still processing',
			error: ''
		},

		// TODO:
		//  multiple
		{
			stepCount: 6,
			title: '',
			log: 'Cluster Component Upgrade',
			error: ''
		},

		{
			stepCount: 7,
			title: 'Cluster Upgrade Plan',
			log: 'still processing',
			error: ''
		},

		{
			stepCount: 8,
			title: 'Cluster Upgrade',
			log: 'still processing',
			error: ''
		},

		// TODO:
		//  multiple
		{
			stepCount: 9,
			title: 'Cluster Component Restart',
			log: 'still processing',
			error: ''
		}
	]);
