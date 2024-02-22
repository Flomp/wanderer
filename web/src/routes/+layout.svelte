<script>
    import Toast from "$lib/components/base/toast.svelte";
    import Footer from "$lib/components/footer.svelte";
    import NavBar from "$lib/components/nav_bar.svelte";
    import { theme } from "$lib/stores/theme_store";
    import "@fortawesome/fontawesome-free/css/all.min.css";
    import "../css/app.css";
    import "../css/components.css";
    import "../css/theme.css";

</script>

<div class="theme {$theme}">
    <NavBar></NavBar>
    <Toast></Toast>
    <slot />

    <Footer></Footer>
</div>
