<script lang="ts">

    type Item = string;

    let items = $state(<Item[]>([]));
    let loading = $state(<boolean>(false));
    let error = $state(<string | null>(null));

    async function fetchItems() {
        loading = true;
        error = null;
        try {
            // make it a kubernetes service
            const response = await fetch("http://k.com/list");
            if (!response.ok) throw new Error("Failed to fetch items");
            const data: Item[] = await response.json();
            items = data;
        } catch (err) {
            error.set(err instanceof Error ? err.message : "Unknown error");
        } finally {
            loading = false;
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
