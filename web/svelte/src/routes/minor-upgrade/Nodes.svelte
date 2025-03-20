<script>
    import { flip } from 'svelte/animate';
    import { send, receive } from './transition.js';

    let { nodes, remove } = $props();
</script>

<ul class="nodes">
    {#each nodes as node (node.id)}
        <li
                class={{ done: node.done }}
                in:receive={{ key: node.id }}
                out:send={{ key: node.id }}
                animate:flip={{ duration: 200 }}
        >
            <label>
                <input type="checkbox" bind:checked={node.done}/>
                <span>{node.description}</span>
                <button onclick={() => remove(node)} aria-label="Remove"></button>
            </label>
        </li>
    {/each}
</ul>

<style>
    label {
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
