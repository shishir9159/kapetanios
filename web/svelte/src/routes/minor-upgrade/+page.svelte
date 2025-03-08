<script lang="ts">
    import { connectionInitialization } from './server.svelte.ts';
    import { Button } from 'flowbite-svelte';
</script>

<div class="action">
    <Button on:mousedown={connectionInitialization}>minor-upgrade</Button>
</div>
