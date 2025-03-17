export class k8sVersion {
    distroPrettyName: string
    targetK8sVersion: string
}

// TODO: camel case
export class SelectedConfig {
    drainNodes:         boolean;
    promptOnError:      boolean;
    certificateRenewal: boolean;
    k8sVersion:         k8sVersion[];
}

export interface nodeStatus {
    NodesUpgraded: [string];
    NodesTobeUpgraded: [string];
    NodesFailedToBeUpgraded: [string];
}

export interface progressDetails {
    StepCount: number;
    Log: string;
    Error: string;
}

export interface progressionState {
    selectedConfig: SelectedConfig;
    nodeStatus: nodeStatus;
    progressDetails: [progressDetails];
}

interface k8sPackageInfo {
    distro:            string
    distroVersion:     string
    distroPrettyName:  string
    repoName:          string
    repoBaseUrl:       string
    availableVersions: [string]
}

interface upgradeConfigSelection{
    K8sPackageInfo: [k8sPackageInfo]
}

export interface UpgradeProgression {
    stateName: string;
    state: progressionState | upgradeConfigSelection;
}
