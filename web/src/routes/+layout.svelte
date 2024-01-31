<script>
    import Toast from "$lib/components/base/toast.svelte";
    import Footer from "$lib/components/footer.svelte";
import NavBar from "$lib/components/nav_bar.svelte";
    import "../app.css";
    import "../components.css"
    import '@fortawesome/fontawesome-free/css/all.min.css'
</script>

<NavBar></NavBar>
<Toast></Toast>
<slot />
<Footer></Footer>