interface K8sPackageInfo {
	distro: string;
	distroVersion: string;
	distroPrettyName: string;
	repoName: string;
	repoBaseUrl: string;
	availableVersions: [string];
}

interface upgradeConfigSelection {
	k8sPackagesInfo: [K8sPackageInfo];
}

export interface UpgradeProgression {
	stateName: string;
	state: progressionState | upgradeConfigSelection;
}

export class K8sVersion {
	distroPrettyName: string;
	targetK8sVersion: string;
}

export interface SelectedConfig {
	drainNodes: boolean;
	promptOnError: boolean;
	certificateRenewal: boolean;
	k8sVersion: [K8sVersion];
}

export interface NodesStatus {
	nodesUpgraded: [string];
	nodesToBeUpgraded: [string];
	nodesFailedToBeUpgraded: [string] | null;
}

export interface ProgressDetail {
	stepCount: number;
	log: string;
	error: string;
}

// {"stateName":"upgrade progression","state":
// 	{"selectedConfig":{"drainNodes":false,"promptOnLock":false,"certificateRenewal":false,
// 		"k8sVersion":[{"distroPrettyName":"Red Hat Enterprise Linux 8.10 (Ootpa)","targetK8sVersion":"1.29.0-150500.1.1"}]},
// 		"nodesStatus":{"nodesUpgraded":[],"nodesToBeUpgraded":["robi-infra-poc-2","robi-infra-poc-3","robi-infra-poc-1"],"nodesFailedToBeUpgraded":null},
// 		"progressDetails":[{"stepCount":1,"log":"","err":""},{"stepCount":1,"log":"","err":""},{"stepCount":1,"log":"","err":""}]}}

interface progressionState {
	selectedConfig: SelectedConfig;
	nodesStatus: NodesStatus;
	progressDetails: [ProgressDetail];
}
