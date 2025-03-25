<script lang="ts">
	import '../app.css';
    import { page } from '$app/state';
    import { DarkMode } from 'flowbite-svelte';
    import { Navbar, NavBrand, NavLi, NavUl } from "flowbite-svelte";
	let { children } = $props();
    let activeUrl = $derived(page.url.pathname)

    let btnClass = 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg text-xl p-2';
</script>

<Navbar>
    <NavBrand>Kapetanios</NavBrand>
    <NavUl {activeUrl}>
        <NavLi href="/minor-upgrade">minor upgrade</NavLi>
        <NavLi href="/certificate">certificate</NavLi>
        <DarkMode {btnClass} />
    </NavUl>
</Navbar>

{@render children()}