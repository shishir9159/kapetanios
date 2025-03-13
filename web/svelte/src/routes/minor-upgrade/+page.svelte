<script lang="ts">
    import { Button, Checkbox, ButtonGroup , Modal, Label , Select } from 'flowbite-svelte';
    import { message } from './upgrading.svelte.ts';
    import type { k8sPackageInfo , upgradeProgression} from './ambient.d.ts';

    let rhel8AvailableVersions = $state([{}]);
    // let packageInfo = $state([{
    //
    // }]);
    let packagesInfo= $state([k8sPackageInfo]);

    let selected = $state();
    let formModal = $state(false);

    function connectionInitialization() {
        const socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
        // const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

        socket.onopen = async () => {
            socket.send('connected to the server');
        };

        socket.onmessage = (e) => {
            let progression: upgradeProgression = JSON.parse(e.data);
            if (progression.stateName == "upgrade config selection") {
                formModal = true;
                if ("k8sPackageInfo" in progression.state) {
                    packagesInfo = progression.state.k8sPackageInfo;
                }

                console.log($state.snapshot(packagesInfo));
            }

            socket.send('current state');
        };

        socket.onerror = (e) => {
            console.error('websocket error:', e);
        };

        socket.onclose = async () => {
            console.log('websocket closed');
        };
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

<Button on:click={() => (formModal = true)}>Form modal</Button>
<Modal bind:open={formModal} size="xs" autoclose={false} class="w-full">

    <Label>
        Select K8s version for each OS
    <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Select K8s version for Different OS Image</h3>
    {#each packagesInfo as packageInfo}
        <Select class="mt-2"
                items={packageInfo.availableVersions}
                bind:value={selected}
                onchange={() => (console.log('selected'))}
        />
    {/each}
    </Label>
        <div class="flex items-start">
            <Checkbox>Certificate Renewal</Checkbox>
        </div>
        <div class="flex items-start">
            <Checkbox>Node Draining before Upgrade</Checkbox>
        </div>
        <div class="flex items-start">
            <Checkbox>Allow different repo for the same OS</Checkbox>
        </div>
        <Button type="submit" class="w-full1">Submit</Button>
</Modal>

<!--<button disabled={!answer} type="submit">-->
<!--    Submit-->
<!--</button>-->

<!-- TODO: fix the height -->
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