<script lang="ts">
    import { onMount } from "svelte";
    import { writable } from "svelte/store";

    // nodeInfo struct
    type Item = string;

    let items = writable<Item[]>([]);
    let loading = writable<boolean>(false);
    let error = writable<string | null>(null);

    async function fetchItems() {
        loading.set(true);
        error.set(null);
        try {
            const response = await fetch("http://k.com/list");
            if (!response.ok) throw new Error("Failed to fetch items");
            const data: Item[] = await response.json();
            items.set(data);
        } catch (err) {
            error.set(err instanceof Error ? err.message : "Unknown error");
        } finally {
            loading.set(false);
        }
    }
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
