<script lang="ts">
    import { Button } from 'flowbite-svelte';
    import { connectionInitialization } from './+server.ts';
</script>

<h1>minor-upgrade</h1>
<button onMousedown={connectionInitialization}> minor-upgrade </button>