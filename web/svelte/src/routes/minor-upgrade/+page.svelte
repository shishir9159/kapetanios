<script lang="ts">
    import { Button, Checkbox , ButtonGroup , Modal, Label , Select } from 'flowbite-svelte';
    import { message } from './upgrading.svelte.ts';
    import { k8sVersion, SelectedConfig } from './ambient.d.ts';
    import type { k8sPackageInfo ,  UpgradeProgression} from './ambient';

    let progression: UpgradeProgression;
    let u: UpgradeProgression = $state(progression);
    let certificateRenewal: boolean = $state(false);
    let drainNodes: boolean = $state(false);
    let promptOnError: boolean = $state(true);
    let s: string = $state('');
    let selected: Map<string, string> = $state(new Map<string, string>());
    let packagesInfo: [k8sPackageInfo]= $state([]);
    let formModal = $state(false);
    $inspect(selected).with(console.trace);

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
        console.log($state.snapshot(selected));
        // if (socket.readyState == "connected") {
        //  TODO: fix
            let selectedConfig: SelectedConfig = new SelectedConfig();
            selectedConfig.certificateRenewal = certificateRenewal;
            selectedConfig.drainNodes = drainNodes;
            selectedConfig.promptOnError = promptOnError;
            let k8sConfig = [k8sVersion];
            console.log("selectedConfig.promptOnError", selectedConfig.promptOnError);
            selected.forEach((value:string, key: string) => {
                const packageInfo = new k8sVersion();
                packageInfo.distroPrettyName = key;
                packageInfo.targetK8sVersion = value;
                k8sConfig.push(packageInfo);
            })
            selectedConfig.k8sVersion = k8sConfig;
            console.log("selectedConfig", selectedConfig);
            console.log("selectedConfig.k8sVersion", selectedConfig.k8sVersion[0].targetK8sVersion);
            const payLoad = JSON.stringify(selectedConfig);

            socket.send(payLoad);
        // } else {
        // // TODO: throw error
        // }
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
        <p style="color: red;">{message.logMessage}</p>
    </div>
</div>

<Modal bind:open={formModal} size="xs" autoclose={false} class="w-full">
    <Label>
        Select K8s version for each OS
    <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Select K8s version for Different OS Image</h3>
    {#each packagesInfo as packageInfo, i}
        <Select class="mt-2"
                items={packageInfo.availableVersions.filter(v=>v).map(value => ({ value: value, name: value }))}
                bind:value={s}
                onselect={() => {
                    console.log('selected', selected);
                    selected[packageInfo.distroPrettyName] = s;
                }}
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