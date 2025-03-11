interface upgradeConfig {
    drainNodes: boolean;
    certificateRenewal: boolean;
    promptOnError: boolean;
    NodesUpgraded: string;
    NodesToBeUpgraded: string;
    UbuntuK8sVersion: string;
    UbuntuK8sRepoName: string;
    UbuntuK8sRepoBaseURL: string;
    Redhat8K8sVersion: string;
    Redhat8K8sRepoName: string;
    Redhat8K8sRepoBaseURL: string;
    Redhat9K8sVersion: string;
    Redhat9K8sRepoName: string;
    Redhat9K8sRepoBaseURL: string;
    UbuntuTargetK8sVersion: string;
    Rhel8TargetK8sVersion: string;
    Rhel9TargetK8sVersion: string;
    UbuntuAvailableK8sVersion: [""];
    Rhel8AvailableK8sVersion: [""];
    Rhel9AvailableK8sVersion: [""];
}


interface PrerequisitesCheck {
    state: "prerequisites"
    log: ""
}

interface UpdateVersionSelection {
    state: "updateVersionSelection"
    log: ""
}

interface Request {
    id: number;
    state: PrerequisitesCheck | UpdateVersionSelection;
}