<script lang="ts">
    import { progressInfo } from './progress.svelte.ts';
    import { AccordionItem, Accordion, Button, ButtonGroup} from "flowbite-svelte";
    let { items = $bindable([false, false, false, false, false, false, false, false, false, false]) } = $props();
    // let { items = $bindable(new Array<boolean>(10).fill(false)) } = $props();
</script>

<Accordion activeClass="bg-blue-100 dark:bg-gray-800 text-blue-600 dark:text-white focus:ring-4 focus:ring-blue-200 dark:focus:ring-blue-800" inactiveClass="text-gray-500 dark:text-gray-400 hover:bg-blue-100 dark:hover:bg-gray-800">
    {#each progressInfo as step, i}
        <AccordionItem bind:open={items[i]}>
        <span slot="header">{step.title}</span>
        <p class="mb-2 text-gray-500 dark:text-gray-400">{step.log}</p>
        </AccordionItem>
    {/each}
</Accordion>