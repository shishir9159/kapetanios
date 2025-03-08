<script lang="ts">
	import '../app.css';
    import { page } from '$app/state';
    import { AccordionItem, Accordion, Navbar, NavBrand, NavLi, NavUl } from "flowbite-svelte";
	let { children } = $props();
    let activeUrl = $derived(page.url.pathname)
</script>

<Navbar>
    <NavBrand>Kapetanios</NavBrand>
    <NavUl {activeUrl}>
        <NavLi href="/minor-upgrade">minor upgrade</NavLi>
        <NavLi href="/certificate">certificate</NavLi>
    </NavUl>
</Navbar>

{@render children()}