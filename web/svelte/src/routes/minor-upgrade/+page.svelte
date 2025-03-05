<script lang="ts">
    import { loading } from './server.svelte.js';
</script>

<main>
    <button on:click={fetchItems} disabled={$loading}>
        {#if $loading} Loading... {/if}
        {#if !$loading} Fetch Items {/if}
    </button>

    {#if $error}
        <p style="color: red;">Error: {$error}</p>
    {/if}

    {#each $items as item}
        <button>{item}</button>
    {/each}
</main>