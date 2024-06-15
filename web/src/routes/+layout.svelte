<script lang="ts">
    import { beforeNavigate, goto } from "$app/navigation";
    import { env } from "$env/dynamic/public";
    import Toast from "$lib/components/base/toast.svelte";
    import Footer from "$lib/components/footer.svelte";
    import NavBar from "$lib/components/nav_bar.svelte";
    import { currentUser } from "$lib/stores/user_store";
    import { isRouteProtected } from "$lib/util/authorization_util";
    import "@fortawesome/fontawesome-free/css/all.min.css";
    import { slide } from "svelte/transition";
    import "../css/app.css";
    import "../css/components.css";
    import "../css/theme.css";
    import { page } from "$app/stores";
    import { onMount } from "svelte";

    
    beforeNavigate((n) => {
        if (!$currentUser && isRouteProtected(n.to?.url?.pathname ?? "")) {
            n.cancel();
            goto("/login?r=" + n.to?.url?.pathname);
        }
    });

    let hideDemoHint = false;
    let hideWarning = false;
</script>

{#if env.PUBLIC_IS_DEMO === "true" && !hideDemoHint}
    <div
        class="flex items-center justify-between bg-amber-200 text-center p-4 text-sm"
        out:slide
    >
        <div></div>
        <span
            >This is a demo instance. Do not store any relevant data here. You
            can use the user 'demo' and password 'password' to login.
        </span>
        <button class="btn-icon self-end" on:click={() => (hideDemoHint = true)}
            ><i class="fa fa-close"></i></button
        >
    </div>
{/if}

{#if $page.data.warnings?.length && !hideWarning}
    <div
        class="flex items-center justify-between bg-red-200 text-center p-4 text-sm"
        out:slide
    >
        <div></div>
        <span>{@html $page.data.warnings[0]} </span>
        <button class="btn-icon self-end" on:click={() => (hideWarning = true)}
            ><i class="fa fa-close"></i></button
        >
    </div>
{/if}

<NavBar></NavBar>
<Toast></Toast>
<slot />

<Footer></Footer>
