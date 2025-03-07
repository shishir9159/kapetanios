<script lang="ts">
    import { connectionInitialization } from './server.svelte.ts';
</script>

<button on:mousedown={connectionInitialization}> minor-upgrade </button>