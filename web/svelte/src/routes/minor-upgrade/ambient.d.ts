interface upgradeConfig {
    drainNodes: boolean;
    certificateRenewal: boolean;
    promptOnError: boolean;
    nodesUpgraded: string;
    nodesToBeUpgraded: string;
    ubuntuK8sVersion: string;
    ubuntuK8sRepoName: string;
    ubuntuK8sRepoBaseURL: string;
    redhat8K8sVersion: string;
    redhat8K8sRepoName: string;
    redhat8K8sRepoBaseURL: string;
    redhat9K8sVersion: string;
    redhat9K8sRepoName: string;
    redhat9K8sRepoBaseURL: string;
    ubuntuTargetK8sVersion: string;
    rhel8TargetK8sVersion: string;
    rhel9TargetK8sVersion: string;
    ubuntuAvailableK8sVersion: [""];
    rhel8AvailableK8sVersion: [""];
    rhel9AvailableK8sVersion: [""];
}

interface MinorUpdate {
    step: 0
    state: upgradeConfig
}

interface UpdateVersionSelection {
    step: 1
    state: upgradeConfig
}

export interface State {
    id: number
    stateName: string
    state: MinorUpdate | UpdateVersionSelection
}