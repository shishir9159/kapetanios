<script lang="ts">
    let { children } = $props();
</script>

<style>
    .action {
        display: flex;
        align-items: center;
        justify-content: flex-end;
    }
</style>

{@render children()}