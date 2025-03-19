<script lang="ts">
    import { Button , ButtonGroup , Checkbox , Modal , Label , Select } from 'flowbite-svelte';
    import Step from './Step.svelte';
    import { progressInfo } from "./progress.svelte.ts";
    import { K8sVersion } from './ambient.d.ts';
    import type { K8sPackageInfo , NodesStatus , ProgressDetails , SelectedConfig , UpgradeProgression} from './ambient';

    let progression: UpgradeProgression;
    let action = $state('');
    //     let upgradeProgression: UpgradeProgression = $state(UpgradeProgression);
    let upgradeProgression: UpgradeProgression = $state(progression);
    let config: SelectedConfig;
    let upgradeConfig: SelectedConfig = $state(config);
    let clusterStatus: NodesStatus;
    let nodeStatus: NodesStatus = $state(clusterStatus);
    // TODO:
    //  fix the following
    let progressDetails: [ProgressDetails] = [];
    let certificateRenewal: boolean = $state(false);
    let drainNodes: boolean = $state(false);
    let promptOnError: boolean = $state(true);
    let s: string = $state('');
    let selected: Map<string, string> = $state(new Map<string, string>());
    let packagesInfo: [K8sPackageInfo]= $state([]);
    let formModal = $state(false);

    let socket: WebSocket   ;
    const wsClient: WebSocketClient | null;

    function connect() {
        wsClient = new WebSocketClient('ws://kapetanios.com/minor-upgrade');
    }

    function connectionInitialization() {
        socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
        // const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

        socket.onopen = async () => {
            console.log('connected to the server')
        };

        socket.onmessage = (e) => {
            upgradeProgression = JSON.parse(e.data);
            if (upgradeProgression.stateName == "upgrade config selection") {
                formModal = true;
                console.log(upgradeProgression);
                if ("k8sPackagesInfo" in upgradeProgression.state) {
                    console.log("packageInfo")
                    packagesInfo = upgradeProgression.state.k8sPackagesInfo;
                }
            } else if (upgradeProgression.stateName == "upgrade progression") {
                if ("selectedConfig" in upgradeProgression.state && "nodeStatus" in upgradeProgression.state && "progressDetails" in upgradeProgression.state) {
                    upgradeConfig = upgradeProgression.state.selectedConfig;
                    clusterStatus = upgradeProgression.state.nodeStatus;
                    progressDetails = upgradeProgression.state.progressDetails;
                } else {
                    // TODO:
                    //   sent error reply and expect another payload
                }

                for (let progressDetail of progressDetails) {
                    progressInfo.forEach((step) => {
                        if (step.stepCount == progressDetail.stepCount) {
                            step.log = progressDetail.log;
                        }
                    })
                }
            }
        };

        socket.onerror = (error) => {
            console.error('websocket error:', error);
        };

        socket.onclose = async () => {
            console.log('websocket closed');
        };
    }

    $effect(() => {
        // if (socket.readyState !== WebSocket.OPEN) {
        // socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
        // }
        socket.send(action);
    })

    function onsubmit() {
        // if (socket.readyState == "connected") {
        let k8sVersion = new Array<K8sVersion>();
        selected.forEach((value:string, key: string) => {
            const packageInfo = new K8sVersion();
            packageInfo.distroPrettyName = key;
            packageInfo.targetK8sVersion = value;
            k8sVersion.push(packageInfo);
        });
        console.log("selected", selected);
        upgradeConfig = {
            drainNodes,
            promptOnError,
            certificateRenewal,
            k8sVersion
        }
        const payLoad = JSON.stringify(upgradeConfig);
        // socket.send(payLoad);

        formModal = false
        // } /else {
        // // TODO: throw error
        // }/
    }
</script>

<ButtonGroup>
    <Button outline color="red" on:mousedown={() => action="retryCurrentStep"}>retry current step</Button>
    <Button outline color="green" on:mousedown={() => action="terminateApplication"}>skip current node</Button>
    <Button outline color="yellow" on:mousedown={() => action="proceedNextStep"}>proceed next step</Button>
</ButtonGroup>


<div class="action">
    <div class="button-group">
        <ButtonGroup>
            <Button outline color="red" on:mousedown={connectionInitialization}>upgrade</Button>
            <Button outline color="green">rollback</Button>
            <Button outline color="yellow">status</Button>
        </ButtonGroup>
    </div>
</div>

<div class="processing">
    <div class="nodes">

    </div>
    <div class="nodes">
        <Step />
    </div>
</div>

<Modal bind:open={formModal} size="xs" autoclose={false} class="w-full">
    <Label>
        Select K8s version for each OS
        <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Select K8s version for Different OS Image</h3>
        {#each packagesInfo as packageInfo}
            <Select class="mt-2"
                    items={packageInfo.availableVersions.filter(v=>v).map(value => ({ value: value, name: value }))}
                    bind:value={s}
                    onchange={() => selected.set(packageInfo.distroPrettyName, s) }
            />
        {/each}
    </Label>
    <div class="flex items-start">
        <Checkbox bind:value={certificateRenewal}>Certificate Renewal</Checkbox>
    </div>
    <div class="flex items-start">
        <Checkbox bind:value={drainNodes}>Node Draining before Upgrade</Checkbox>
    </div>
    <div class="flex items-start">
        <Checkbox bind:value={promptOnError}>Allow different repo for the same OS</Checkbox>
    </div>
    <!--    TODO: disabled={all packages are selected}  -->
    <Button type="submit" class="w-full1" on:click={() => (onsubmit())}>Submit</Button>
</Modal>

<style>
    .action {
        display: flex;
        /*flex: 0 0 100%;*/
        /*height: 100px;*/
        /*margin: 100px;*/
        width: 100vw;
        padding: 50px;
        /*margin: 10px 150px 100px 80px;*/
    }
    .button-group {
        margin-left: auto;
        padding: 50px;
        flex-direction: row-reverse;
        justify-content: flex-start;
    }

    .processing {
        display: flex;
        width: 80vw;
        margin: auto;
        flex-wrap: nowrap;
        /*border: 1px solid #000000;*/
    }
    .nodes {
        display: flex;
        flex: 1;
        padding: 50px;
        text-align: justify-all;
        align-items: stretch;
        white-space: normal;
        word-wrap: break-word;
    }
</style>