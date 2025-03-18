<script lang="ts">
    import { Button , Checkbox , ButtonGroup , Modal , Label , Select } from 'flowbite-svelte';
    import { progresses } from './progress.svelte.ts';
    import { message } from './upgrading.svelte.ts';
    import Step from './Step.svelte';
    import type {k8sVersion, nodeStatus, selectedConfig} from './ambient.d.ts';
    import type { k8sPackageInfo, progressDetails , UpgradeProgression} from './ambient';

    let progression: UpgradeProgression;
    let u: UpgradeProgression = $state(progression);
    let certificateRenewal: boolean = $state(false);
    let drainNodes: boolean = $state(false);
    let promptOnError: boolean = $state(true);
    let s: string = $state('');
    let selected: Map<string, string> = $state(new Map<string, string>());
    let packagesInfo: [k8sPackageInfo]= $state([]);
    let upgradeConfig = $state(selectedConfig);
    let clusterStatus = $state(nodeStatus);
    let progressionDetails = $state([progressDetails]);
    let formModal = $state(false);

    // TODO:
    //  display selectedConfig

    let socket: WebSocket;

    function connectionInitialization() {
        socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
        // const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

        socket.onopen = async () => {
            socket.send('connected to the server');
        };

        socket.onmessage = (e) => {
            u = JSON.parse(e.data);
            if (u.stateName == "upgrade config selection") {
                formModal = true;
                if ("k8sPackageInfo" in u.state) {
                    packagesInfo = u.state.k8sPackageInfo;
                }
            } else if (u.stateName == "upgrade progression") {
                if ("selectedConfig" in u.state && "nodeStatus" in u.state && "progressDetails" in u.state) {
                    upgradeConfig = u.state.selectedConfig;
                    clusterStatus = u.state.nodeStatus;
                    progressionDetails = u.state.progressDetails;
                } else {
                // TODO:
                //   sent error reply and expect another payload
                }

                for (let progressDetail of progressionDetails) {
                    // TODO: make progresses a map
                    // progresses[progressDetail.stepCount] = ;
                }

                socket.send('received');
            }
        };

        socket.onerror = (error) => {
            console.error('websocket error:', error);
        };

        socket.onclose = async () => {
            console.log('websocket closed');
        };
    }

    function onsubmit() {
        // TODO: check connection
        // if (socket.readyState == "connected") {
            let k8sConfig = new Array<k8sVersion>();
            upgradeConfig = selectedConfig
            {
                certificateRenewal
                progression
                drainNodes
                promptOnError
            }
            console.log("selected", selected);
            selected.forEach((value:string, key: string) => {
                const packageInfo = new k8sVersion();
                packageInfo.distroPrettyName = key;
                packageInfo.targetK8sVersion = value;
                k8sConfig.push(packageInfo);
            });
            upgradeConfig.k8sVersion = k8sConfig;
            const payLoad = JSON.stringify(upgradeConfig);
            socket.send(payLoad);
        //  TODO:
        //   check error
            formModal = false;

        // } /else {
        // // TODO: throw error
        // }/
    }
</script>

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
<!--        <p style="color: red;">{message.logMessage}</p>-->
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