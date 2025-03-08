<script lang="ts">
    let { children } = $props();
</script>

<style>
    .action {
        display: flex;
        /*align-items: center;*/
        justify-content: right;
    }
</style>

{@render children()}