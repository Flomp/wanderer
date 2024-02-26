<script>
    import Toast from "$lib/components/base/toast.svelte";
    import Footer from "$lib/components/footer.svelte";
    import NavBar from "$lib/components/nav_bar.svelte";
    import "@fortawesome/fontawesome-free/css/all.min.css";
    import "../css/app.css";
    import "../css/components.css";
    import "../css/theme.css";
</script>

<svelte:head>
    <title>| wanderer</title>
</svelte:head>
<NavBar></NavBar>
<Toast></Toast>
<slot />

<Footer></Footer>
