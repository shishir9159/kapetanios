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

interface MinorUpdate {
    step: 0
    state: upgradeConfig
}

interface UpdateVersionSelection {
    step: 1
    state: upgradeConfig
}

export interface UpgradeProgression {
    id: number
    stateName: string
    state: MinorUpdate | UpdateVersionSelection
}

type UpgradeProgression struct {
    StateName string `json:"stateName"`
    State     State  `json:"state"`
}

type selectedConfig struct {
    DrainNodes         bool         `json:"drainNodes"`
    PromptOnError      bool         `json:"promptOnLock"`
    CertificateRenewal bool         `json:"certificateRenewal"`
    K8sVersion         []K8sVersion `json:"k8sVersion"`
}