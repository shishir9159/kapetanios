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
	state: SelectedConfig | upgradeConfigSelection;
}

export class K8sVersion {
	distroPrettyName: string;
	targetK8sVersion: string;
}

export class SelectedConfig {
	drainNodes: boolean;
	promptOnError: boolean;
	certificateRenewal: boolean;
	k8sVersion: [k8sVersion];
}

export interface NodesStatus {
	nodesUpgraded: [string];
	nodesTobeUpgraded: [string];
	nodesFailedToBeUpgraded: [string];
}

export interface ProgressDetails {
	stepCount: number;
	log: string;
	error: string;
}

export interface ProgressionState {
	selectedConfig: SelectedConfig;
	nodeStatus: NodesStatus;
	progressDetails: [ProgressDetails];
}
