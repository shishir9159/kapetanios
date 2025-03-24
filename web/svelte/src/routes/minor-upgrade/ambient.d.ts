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

interface progressionState {
	selectedConfig: SelectedConfig;
	nodesStatus: NodesStatus;
	progressDetails: [ProgressDetail];
}

interface finalizer {
	nodesStatus: NodesStatus;
	progressDetails: [ProgressDetail];
}

export interface UpgradeProgression {
	stateName: string;
	state: progressionState | upgradeConfigSelection | finalizer;
}
