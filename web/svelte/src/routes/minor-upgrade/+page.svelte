<script lang="ts">
    import { Input, ButtonGroup, Button, Modal, Label, Checkbox } from 'flowbite-svelte';

    let rhel8AvailableVersions = $state([""]);
    let selected = $state();
    let formModal = false;

    let logMessage: string = $state("didn't change");

    function connectionInitialization() {

        const socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
        // const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

        socket.onopen = async () => {
            console.log('connected to the server');
            socket.send('next');
            //     "current step"
            //     prepare response accordingly
        };

        socket.onmessage = (e) => {
            logMessage = e.data;
            // message = utf8Array;
            // const message = JSON.parse(e.data);
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
            <Button outline color="red" on:mousedown={connectionInitialization()}>upgrade</Button>
            <Button outline color="green">rollback</Button>
            <Button outline color="yellow">status</Button>
        </ButtonGroup>
    </div>
</div>

<div class="processing">
    <div class="nodes">
    </div>
    <div class="nodes">
        <p style="color: red;">{logMessage}</p>
    </div>
</div>

<Button on:click={() => (formModal = true)}>Form modal</Button>
<Modal bind:open={formModal} size="xs" autoclose={false} class="w-full">
    <form class="flex flex-col space-y-6" action="#">
        <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Sign in to our platform</h3>
        <Label class="space-y-2">
            <span>Email</span>
            <Input type="email" name="email" placeholder="name@company.com" required />
        </Label>
        <Label class="space-y-2">
            <span>Your password</span>
            <Input type="password" name="password" placeholder="•••••" required />
        </Label>
        <div class="flex items-start">
            <Checkbox>Remember me</Checkbox>
            <a href="/" class="ms-auto text-sm text-primary-700 hover:underline dark:text-primary-500"> Lost password? </a>
        </div>
        <Button type="submit" class="w-full1">Login to your account</Button>
        <div class="text-sm font-medium text-gray-500 dark:text-gray-300">
            Not registered? <a href="/" class="text-primary-700 hover:underline dark:text-primary-500"> Create account </a>
        </div>
    </form>
</Modal>

<select
        bind:value={selected}
        onchange={() => (console.log('selected'))}
>
    {#each rhel8AvailableVersions as version}
        <option value={version}>
            {version}
        </option>
    {/each}
</select>

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