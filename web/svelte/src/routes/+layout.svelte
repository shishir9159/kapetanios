<script lang="ts">
	import '../app.css';
    import {Navbar, NavBrand, NavLi, NavUl} from "flowbite-svelte";
	let { children } = $props();
</script>

<!--<svelte:head>-->
<Navbar>
    <NavBrand>Kapetanios</NavBrand>
    <NavUl>
        <NavLi href="/minor-upgrade">minor upgrade</NavLi>
        <NavLi href="/certificate">certificate</NavLi>
    </NavUl>
</Navbar>
<!--</svelte:head>-->

{@render children()}