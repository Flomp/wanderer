<script>
    import { beforeNavigate, goto } from "$app/navigation";
    import Toast from "$lib/components/base/toast.svelte";
    import Footer from "$lib/components/footer.svelte";
    import NavBar from "$lib/components/nav_bar.svelte";
    import { isRouteProtected } from "$lib/util/authorization_util";
    import "@fortawesome/fontawesome-free/css/all.min.css";
    import "../css/app.css";
    import "../css/components.css";
    import "../css/theme.css";
    import { currentUser } from "$lib/stores/user_store";

    beforeNavigate((n) => {
        if (!$currentUser && isRouteProtected(n.to?.url?.pathname ?? "")) {
            n.cancel();
            goto("/login");
        }
    });
</script>

<NavBar></NavBar>
<Toast></Toast>
<slot />

<Footer></Footer>
