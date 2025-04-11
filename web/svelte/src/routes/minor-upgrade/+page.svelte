<script lang="ts">
    import Step from './Step.svelte';
    import NodeList from './Nodes.svelte';
    import { K8sVersion } from './ambient.d.ts';
    import { progressInfo } from "./progress.svelte.ts";
    import { Button , ButtonGroup , Checkbox , Modal , Label , Select } from 'flowbite-svelte';
    import type { K8sPackageInfo , NodesStatus , ProgressDetail , SelectedConfig , UpgradeProgression } from './ambient';

    let config: SelectedConfig;
    let upgradeConfig: SelectedConfig = $state(config);
    let clusterStatus: NodesStatus;
    let nodeStatus: NodesStatus = $state(clusterStatus);
    let progressDetails: [ProgressDetail] = [];
    let drainNodes: boolean = $state(false);
    let certificateRenewal: boolean = $state(false);
    let promptOnError: boolean = $state(true);
    let selected: Map<string, string> = $state(new Map<string, string>());
    let packagesInfo: [K8sPackageInfo]= $state([]);
    let formModal = $state(false);

    let scrollingModal = $state(false);
    let finalizerModal = $state(false);
    let errorForScrollingModal = $state("");
    let logForScrollingModal = $state("");

    let selectedVersion: string;
    let toggleAccordion: boolean[] = $state([false, false, false, false, false, false, false, false, false, false]);
    let nodes = $state([
        // { id: 1, done: false, name: 'node one' },
        // { id: 2, done: true, name: 'node two' },
    ]);

    class WebSocketClient {
        private socket: WebSocket | null = null;
        private readonly url: string;
        private shouldReconnect = true;
        private reconnectInterval = 10000;

        constructor(url: string) {
            this.url = url;
        }

        public connect(): void {
            this.socket = new WebSocket(this.url);

            this.socket.onopen = () => {
                console.log('connected to the server');
                this.sendMessage("connection initialized");
            };

            this.socket.onmessage = (event) => {
                try {
                    console.log("event data ", event.data);
                    const upgradeProgression: UpgradeProgression = JSON.parse(event.data);
                    if (upgradeProgression.stateName == 'upgrade config selection') {
                        formModal = true;
                        console.log(upgradeProgression);
                        if ('k8sPackagesInfo' in upgradeProgression.state) {
                            packagesInfo = upgradeProgression.state.k8sPackagesInfo;
                        }
                    } else if (upgradeProgression.stateName == 'upgrade progression') {
                        if (
                            'selectedConfig' in upgradeProgression.state &&
                            'nodesStatus' in upgradeProgression.state &&
                            'progressDetails' in upgradeProgression.state
                        ) {
                            upgradeConfig = upgradeProgression.state.selectedConfig;
                            nodeStatus = upgradeProgression.state.nodesStatus;
                            progressDetails = upgradeProgression.state.progressDetails;
                        } else {
                            this.sendMessage("repeat the payload");
                        }

                        for (let progressDetail of progressDetails) {
                            progressInfo.forEach((step) => {
                                if (step.stepCount == progressDetail.stepCount) {
                                    step.log = progressDetail.log;
                                    if (step.error != "") {

                                    }
                                }
                            });
                        }

                        let lastStep = progressDetails.length - 1;

                        toggleAccordion = [...new Array(10)].map(() => false);
                        toggleAccordion[lastStep] = true;

                        if (progressDetails[lastStep].err || !upgradeConfig.promptOnError) {
                            scrollingModal = true
                            console.log("lastStep", lastStep);
                            console.log("progressDetails[lastStep].error", progressDetails[lastStep].err);
                            errorForScrollingModal = progressDetails[lastStep].err
                            logForScrollingModal = progressDetails[lastStep].log
                        }

                        let uid = 1;
                        let nodesToBeUpgraded = [];
                        nodeStatus.nodesToBeUpgraded.forEach(node => {
                            nodesToBeUpgraded.push({id: uid++, done: false, name: node});
                        })
                        nodesToBeUpgraded.sort((a, b) => (a.name > b.name ? 1 : -1));

                        let nodesUpgraded = [];
                        nodeStatus.nodesUpgraded.forEach(node => {
                            nodesUpgraded.push({id: uid++, done: true, name: node});
                        })
                        nodesUpgraded.sort((a, b) => (a.name > b.name ? 1 : -1));
                        nodes = nodesToBeUpgraded.concat(nodesUpgraded);
                    } else if (upgradeProgression.stateName == 'finalizer') {
                        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                            this.socket.send("finalizer received");
                        }
                        this.shouldReconnect = false;

                        if ('nodesStatus' in upgradeProgression.state) {
                            nodeStatus = upgradeProgression.state.nodesStatus;
                            finalizerModal = true;
                        }

                        let uid = 1;
                        let nodesToBeUpgraded = [];
                        nodeStatus.nodesToBeUpgraded.forEach(node => {
                            nodesToBeUpgraded.push({id: uid++, done: false, name: node});
                        })
                        nodesToBeUpgraded.sort((a, b) => (a.name > b.name ? 1 : -1));

                        let nodesUpgraded = [];
                        nodeStatus.nodesUpgraded.forEach(node => {
                            nodesUpgraded.push({id: uid++, done: true, name: node});
                        })
                        nodesUpgraded.sort((a, b) => (a.name > b.name ? 1 : -1));
                        nodes = nodesToBeUpgraded.concat(nodesUpgraded);
                    }
                } catch (error) {
                    console.error('error parsing message:', error);
                }
            };

            this.socket.onerror = (error) => {
                console.error('webSocket error:', error);
            };

            this.socket.onclose = (event) => {
                console.warn(`webSocket closed: ${event.code} - ${event.reason}`);
                if (this.shouldReconnect) {
                    setTimeout(() => this.connect(), this.reconnectInterval);
                }
            };
        }

        public sendMessage(message: string): void {
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(message);
            } else {
                console.warn('webSocket connection is not open');
            }
        }

        public sendJson(message: object): void {
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify(message));
            } else {
                console.warn('webSocket connection is not open');
            }
        }
    }

    const wsClient = new WebSocketClient('ws://kapetanios.com/minor-upgrade');
    function onsubmit() {
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
        wsClient.sendJson(upgradeConfig);

        selectedVersion = []

        formModal = false
    }

    function remove(node) {
        const index = nodes.indexOf(node);
        nodes.splice(index, 1);
    }
</script>

<div class="action">
    <div class="button-group">
        <ButtonGroup>
            <Button outline color="red" on:mousedown={() => wsClient.connect()}>upgrade</Button>
            <Button outline color="green">rollback</Button>
            <Button outline color="yellow">status</Button>
        </ButtonGroup>
    </div>
</div>

<div class="action">
    <div class="button-group">
        <ButtonGroup>
            <Button outline color="red"
                    on:mousedown={() => {
                        wsClient.sendMessage("retryCurrentStep");
                        scrollingModal= false;
                    }}>retry current step</Button>
            <Button outline color="green" on:mousedown={() => wsClient.sendMessage("terminateApplication")}>skip current node</Button>
            <Button outline color="yellow" on:mousedown={() => {wsClient.sendMessage("proceedNextStep"); scrollingModal= false;}}>proceed next step</Button>
        </ButtonGroup>
    </div>
</div>

<div class="processing">
    <div class="nodes">

        <div class="nodeList">
            <h2 class="text-black dark:text-white">nodes to be upgraded</h2>
            <NodeList nodes={nodes.filter((t) => !t.done)} />
<!--            <NodeList nodes={nodes.filter((t) => !t.done)} {remove} />-->
        </div>

        <div class="done">
            <h2 class="text-black dark:text-white">upgraded nodes</h2>
            <NodeList nodes={nodes.filter((t) => t.done)} />
<!--            <NodeList nodes={nodes.filter((t) => t.done)} {remove} />-->
        </div>

    </div>
    <div class="nodes">
        <Step {toggleAccordion} />
    </div>
</div>

<Modal bind:open={formModal} size="xs" autoclose={false} class="w-full">
    <Label>
        Select K8s version for each OS
        <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Select K8s version for Different OS Image</h3>
        {#each packagesInfo as packageInfo, i}
            <p>{packageInfo.distroPrettyName}</p>
            <Select class="mt-2"
                    items={packageInfo.availableVersions.filter(v=>v).map(value => ({ value: value, name: value }))}
                    bind:value={selectedVersion}
                    onchange={() => selected.set(packageInfo.distroPrettyName, selectedVersion)}
            />
        {/each}
    </Label>
    <div class="flex items-start">
        <Checkbox bind:checked={certificateRenewal}>Certificate Renewal</Checkbox>
    </div>
    <div class="flex items-start">
        <Checkbox bind:checked={drainNodes}>Node Draining before Upgrade</Checkbox>
    </div>
    <div class="flex items-start">
        <Checkbox bind:checked={promptOnError}>Only Prompt on Error</Checkbox>
    </div>
    <!--    TODO: disabled(until)={all packages are selected}  -->
    <Button type="submit" class="w-full1" on:click={() => (onsubmit())}>Submit</Button>
</Modal>

<!--<Modal title={upgradeConfig.promptOnError ? "Prompt on error catch" : "Progress Information"} bind:open={scrollingModal}>-->
<Modal title="Prompt on error catch" bind:open={scrollingModal}>
<!--    <summary></summary> >-->
    <p class="text-base leading-relaxed text-gray-500 dark:text-gray-400">{errorForScrollingModal}</p>
    <p class="text-base leading-relaxed text-gray-500 dark:text-gray-400">{logForScrollingModal}</p>
    <svelte:fragment slot="footer">
            <ButtonGroup>
                <Button pill color="red" on:mousedown={() => wsClient.sendMessage("retryCurrentStep")}>retry current step</Button>
                <Button pill color="green" on:mousedown={() => wsClient.sendMessage("terminateApplication")}>skip current node</Button>
                <Button pill color="yellow" on:mousedown={() => wsClient.sendMessage("proceedNextStep")}>proceed next step</Button>
            </ButtonGroup>
<!--        <Button on:click={() => alert('Handle "success"')}>I accept</Button>-->
<!--        <Button color="alternative">Decline</Button>-->
    </svelte:fragment>
</Modal>

<Modal title="Minor Upgrade Successful" bind:open={finalizerModal} color="green" autoclose>
    <div class="text-base leading-relaxed">Upgrade Report:</div>
    <svelte:fragment slot="footer">
        <!--        TODO:        -->
        <Button on:click={() => console.log('finalizer received')} color="green">Homepage</Button>
        <Button color="alternative">Okay</Button>
    </svelte:fragment>
</Modal>

<!-- TODO: error Modal -->

<style>
    .action {
        display: flex;
        /*flex: 0 0 100%;*/
        /*height: 100px;*/
        /*margin: 100px;*/
        width: 100vw;
        padding: 0;
        /*margin: 10px 150px 100px 80px;*/
    }
    .button-group {
        margin-left: auto;
        padding-right: 50px;
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
        /*display: flex;*/
        flex: 1;
        padding: 50px;
        text-align: justify-all;
        align-items: flex-start;
        white-space: normal;
        word-wrap: break-word;
    }

    h2 {
        font-size: 2em;
        font-weight: 200;
    }
</style>