<script lang="ts">
    import { Button } from 'flowbite-svelte';
    import { connectionInitialization } from './server.svelte.ts';
</script>

<div class="action">
    <Button on:mousedown={connectionInitialization}>minor-upgrade</Button>
</div>
