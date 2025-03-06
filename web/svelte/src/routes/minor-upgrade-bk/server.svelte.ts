type Item = string;
let items = $state(<Item[]>([]));
let loading = $state(<boolean>false);
let error = $state(<string | null>null);

async function fetchItems() {

    loading = true;
    error = null;
    const socket = new WebSocket("http://kapetanios.default.cluster.local/minor-upgrade",
        );

    try {
        // make it a local host
        const response = await fetch("", {
            method: "GET",
        });
        if (!response.ok) throw new Error("Failed to fetch items");
        const data: Item[] = await response.json();
        items = data;
    } catch (err) {
        error = (err instanceof Error ? err.message : "Unknown error");
    } finally {
        loading = false;
    }
}

