<script lang="ts">
    import { connectionInitialization } from './server.svelte.ts';
    import { Button } from 'flowbite-svelte';
</script>

<Button on:mousedown={connectionInitialization}>minor-upgrade</Button>