<script>
    import { flip } from 'svelte/animate';
    import { send, receive } from './transition.js';
    import { Button } from "flowbite-svelte";

    let { nodes, remove } = $props();

</script>

<ul class="nodes">
    {#each nodes as node (node.name)}
        <li
                class={{ done: node.done }}
                in:receive={{ key: node.id }}
                out:send={{ key: node.id }}
                animate:flip={{ duration: 200 }}
        >
            <label>
                <Button type="text" bind:checked={node.done}>
                    <span>{node.name}</span>
                </Button>
                <!--                TODO                -->
                <button onclick={() => remove(node)} aria-label="Remove"></button>
            </label>
        </li>
    {/each}
</ul>

<style>
    label {
        padding: 5px;
        width: 100%;
        height: 100%;
        display: flex;
    }

    span {
        flex: 1;
    }

    button {
        /*background-image: url(./remove.svg);*/
    }
</style>
