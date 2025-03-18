interface K8sPackageInfo {
	Distro: string;
	DistroVersion: string;
	DistroPrettyName: string;
	RepoName: string;
	RepoBaseUrl: string;
	AvailableVersions: [string];
}

interface upgradeConfigSelection {
	K8sPackagesInfo: [k8sPackageInfo];
}

export interface UpgradeProgression {
	StateName: string;
	State: SelectedConfig | upgradeConfigSelection;
}

export class K8sVersion {
	DistroPrettyName: string;
	TargetK8sVersion: string;
}

// TODO: camel case
export class SelectedConfig {
	DrainNodes: boolean;
	PromptOnError: boolean;
	CertificateRenewal: boolean;
	K8sVersion: k8sVersion[];
}

export interface NodesStatus {
	NodesUpgraded: [string];
	NodesTobeUpgraded: [string];
	NodesFailedToBeUpgraded: [string];
}

export interface ProgressDetails {
	StepCount: number;
	Log: string;
	Error: string;
}

export interface ProgressionState {
	selectedConfig: SelectedConfig;
	nodeStatus: NodesStatus;
	progressDetails: [ProgressDetails];
}
