<script lang="ts">
    import Step from './Step.svelte';
    import { K8sVersion } from './ambient.d.ts';
    import { progressInfo } from "./progress.svelte.ts";
    import { Button , ButtonGroup , Checkbox , Modal , Label , Select } from 'flowbite-svelte';
    import type { K8sPackageInfo , NodesStatus , ProgressDetail , SelectedConfig , UpgradeProgression} from './ambient';

    let config: SelectedConfig;
    let upgradeConfig: SelectedConfig = $state(config);
    let clusterStatus: NodesStatus;
    let nodeStatus: NodesStatus = $state(clusterStatus);
    let progressDetails: [ProgressDetail] = [];
    let certificateRenewal: boolean = $state(false);
    let drainNodes: boolean = $state(false);
    let promptOnError: boolean = $state(false);
    let s: string = $state('');
    let selected: Map<string, string> = $state(new Map<string, string>());
    let packagesInfo: [K8sPackageInfo]= $state([]);
    let formModal = $state(false);

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
                // this.sendMessage({ type: 'ping', data: 'health check beat' });
            };

            this.socket.onmessage = (event) => {
                try {
                    console.log("event data ",event.data);
                    const upgradeProgression: UpgradeProgression = JSON.parse(event.data);
                    // this.handleMessage(event);
                    if (upgradeProgression.stateName == 'upgrade config selection') {
                        formModal = true;
                        console.log(upgradeProgression);
                        if ('k8sPackagesInfo' in upgradeProgression.state) {
                            console.log('packageInfo');
                            packagesInfo = upgradeProgression.state.k8sPackagesInfo;
                        }
                    } else if (upgradeProgression.stateName == 'upgrade progression') {
                        if (
                            'selectedConfig' in upgradeProgression.state &&
                            'nodesStatus' in upgradeProgression.state &&
                            'progressDetails' in upgradeProgression.state
                        ) {
                            upgradeConfig = upgradeProgression.state.selectedConfig;
                            clusterStatus = upgradeProgression.state.nodesStatus;
                            progressDetails = upgradeProgression.state.progressDetails;
                            console.log("upgrade progression state", upgradeProgression.state);
                        } else {
                            // TODO:
                            //   sent error reply and expect another payload
                        }

                        for (let progressDetail of progressDetails) {
                            progressInfo.forEach((step) => {
                                if (step.stepCount == progressDetail.stepCount) {
                                    step.log = progressDetail.log;
                                }
                            });
                        }
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

        private handleMessage(event: MessageEvent): void {
            try {
                const message = JSON.parse(event.data);
                console.log('Received message:', message);
            } catch (error) {
                console.error('Error parsing message:', error);
            }
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

        public close(): void {
            this.shouldReconnect = false;
            this.socket?.close();
        }
    }

    const wsClient = new WebSocketClient('ws://kapetanios.com/minor-upgrade');
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
        // const payLoad = JSON.stringify();
        wsClient.sendJson(upgradeConfig);
        // socket.send(payLoad);

        formModal = false
        // } /else {
        // // TODO: throw error
        // }/
    }
</script>

<ButtonGroup>
    <Button outline color="red" on:mousedown={() => wsClient.sendMessage("retryCurrentStep")}>retry current step</Button>
    <Button outline color="green" on:mousedown={() => wsClient.sendMessage("terminateApplication")}>skip current node</Button>
    <Button outline color="yellow" on:mousedown={() => wsClient.sendMessage("proceedNextStep")}>proceed next step</Button>
</ButtonGroup>

<div class="action">
    <div class="button-group">
        <ButtonGroup>
            <Button outline color="red" on:mousedown={() => wsClient.connect()}>upgrade</Button>
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