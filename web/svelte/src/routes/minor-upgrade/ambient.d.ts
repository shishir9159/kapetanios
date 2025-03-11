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
    Rhel8AvailableK8sVersion: string;
    Rhel9TargetK8sVersion: string;
    UbuntuAvailableK8sVersion: [""];
    Rhel8AvailableK8sVersion: [""];
    Rhel9AvailableK8sVersion: [""];

    promptOnError             bool     `yaml:"promptOnLock"`
    NodesUpgraded             string   `yaml:"nodesUpgraded"`
    NodesToBeUpgraded         string   `yaml:"nodesToBeUpgraded"`
    UbuntuK8sVersion          string   `yaml:"ubuntuK8sVersion"` // currently only works with 24.02
    UbuntuK8sRepoName         string   `yaml:"ubuntuK8sRepoName"`
    UbuntuK8sRepoBaseURL      string   `yaml:"ubuntuK8sRepoBaseURL"`
    Redhat8K8sVersion         string   `yaml:"redhat8K8sVersion"`
    Redhat8K8sRepoName        string   `yaml:"redhat8K8sRepoName"`
    Redhat8K8sRepoBaseURL     string   `yaml:"redhat8K8sRepoBaseURL"`
    Redhat9K8sVersion         string   `yaml:"redhat9K8sVersion"`
    Redhat9K8sRepoName        string   `yaml:"redhat9K8sRepoName"`
    Redhat9K8sRepoBaseURL     string   `yaml:"redhat9K8sRepoBaseURL"`
    UbuntuTargetK8sVersion    string   `yaml:"ubuntuTargetK8sVersion"`
    Rhel8TargetK8sVersion     string   `yaml:"rhel8TargetK8sVersion"`
    Rhel9TargetK8sVersion     string   `yaml:"rhel9TargetK8sVersion"`
    UbuntuAvailableK8sVersion []string `yaml:"ubuntuAvailableK8sVersion"`
    Rhel8AvailableK8sVersion  []string `yaml:"rhel8AvailableK8sVersion"`
    Rhel9AvailableK8sVersion  []string `yaml:"rhel9AvailableK8sVersion"`
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