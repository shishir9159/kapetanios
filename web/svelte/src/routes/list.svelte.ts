<script>
    let items = [];

async function fetchList() {
    try {
        const response = await fetch("http://k.com/list");
        if (!response.ok) throw new Error("Failed to fetch");
        items = await response.json();
    } catch (error) {
        console.error(error);
    }
}
</script>

<button on:click={fetchList}>Fetch List</button>

{#each items as item}
<button>{item}</button>
{/each}
