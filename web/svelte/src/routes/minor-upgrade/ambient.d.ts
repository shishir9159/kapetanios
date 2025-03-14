interface upgradeConfig {
    drainNodes: boolean;
    certificateRenewal: boolean;
    promptOnError: boolean;
    nodesUpgraded: string;
    nodesToBeUpgraded: string;
    ubuntuK8sVersion: string;
    ubuntuK8sRepoName: string;
    ubuntuK8sRepoBaseURL: string;
    rhel8K8sVersion: string;
    rhel8K8sRepoName: string;
    rhel8K8sRepoBaseURL: string;
    rhel9K8sVersion: string;
    rhel9K8sRepoName: string;
    rhel9K8sRepoBaseURL: string;
    ubuntuTargetK8sVersion: string;
    rhel8TargetK8sVersion: string;
    rhel9TargetK8sVersion: string;
    ubuntuAvailableK8sVersion: [""];
    rhel8AvailableK8sVersion: [""];
    rhel9AvailableK8sVersion: [""];
}

interface UpdateVersionSelection {
    step: 1
    log: string
}

interface MinorUpdate {
    step: 0
    state: upgradeConfig
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
    state: MinorUpdate | upgradeConfigSelection;
}

export interface k8sVersion {
    distroPrettyName: string
    targetK8sVersion: string
}

interface SelectedConfig {
    drainNodes:         boolean;
    promptOnError:      boolean;
    certificateRenewal: boolean;
    k8sVersion:         [k8sVersion];
}