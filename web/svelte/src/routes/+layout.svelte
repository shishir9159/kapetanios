<script lang="ts">
	import '../app.css';
    import { page } from '$app/state';
    import {Navbar, NavBrand, NavLi, NavUl} from "flowbite-svelte";
	let { children } = $props();
</script>

<header>
<Navbar>
    <NavBrand>Kapetanios</NavBrand>
    <NavUl>
        <NavLi href="/minor-upgrade" aria-current={page.url.pathname === '/minor-upgrade'}>minor upgrade</NavLi>
        <NavLi href="/certificate">certificate</NavLi>
    </NavUl>
</Navbar>
</header>

{@render children()}