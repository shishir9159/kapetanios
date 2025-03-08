<script lang="ts">
    import { page } from '$app/state';
    import { Navbar, NavBrand, NavLi, NavUl, NavHamburger } from 'flowbite-svelte';
    let activeUrl = page.url.pathname;
</script>

<Navbar>
    <NavBrand>Kapetanios</NavBrand>
    <NavUl>
        <NavLi href="/minor-upgrade">minor upgrade</NavLi>
        <NavLi href="/certificate">certificate</NavLi>
    </NavUl>
</Navbar>


<style>
    ul {
        list-style-type: none;
        margin: 0;
        padding: 0;
    }
</style>