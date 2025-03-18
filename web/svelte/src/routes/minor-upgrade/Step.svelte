<script lang="ts">
    import { progresses } from './progress.svelte.ts';
    import { AccordionItem, Accordion } from "flowbite-svelte";
</script>

<Accordion activeClass="bg-blue-100 dark:bg-gray-800 text-blue-600 dark:text-white focus:ring-4 focus:ring-blue-200 dark:focus:ring-blue-800" inactiveClass="text-gray-500 dark:text-gray-400 hover:bg-blue-100 dark:hover:bg-gray-800">
    {#each progresses as progress, i}
        <AccordionItem class="">
        <span slot="header">{progress.title}</span>
        <p class="mb-2 text-gray-500 dark:text-gray-400">{progress.log}</p>
        </AccordionItem>
    {/each}
</Accordion>