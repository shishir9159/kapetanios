<script lang="ts">
    import { ButtonGroup, Button } from 'flowbite-svelte';
    import { message, connectionInitialization } from './server.svelte.ts';
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
        <p>Processing...</p>
        <p>{message}</p>
    </div>
</div>


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
        margin: 100px auto;
        border: 1px solid #000000;
    }
    .nodes {
        display: flex;
        flex: 1;
        padding: 50px;
        text-align: center;
    }
</style>