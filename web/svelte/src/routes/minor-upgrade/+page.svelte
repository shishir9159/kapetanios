<script lang="ts">
    import { connectionInitialization } from './server.svelte.ts';
    import { Button } from 'flowbite-svelte';
</script>

<!--<body>-->
    <div class="action">
        <Button on:mousedown={connectionInitialization}>minor-upgrade</Button>
    </div>
<!--</body>-->