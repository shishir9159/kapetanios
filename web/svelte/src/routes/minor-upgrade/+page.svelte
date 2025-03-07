<script lang="ts">
    import { Button } from 'flowbite-svelte';
    import { connectionInitialization } from './server.svelte.ts';
</script>

<h1>minor-upgrade</h1>
<Button on:mousedown={connectionInitialization}> minor-upgrade </Button>