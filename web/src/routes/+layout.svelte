<script lang="ts">
    import { browser } from "$app/environment";
    import { beforeNavigate, goto } from "$app/navigation";
    import { page } from "$app/stores";
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
    import { onMount } from "svelte";

    beforeNavigate((n) => {
        if (!$currentUser && isRouteProtected(n.to?.url?.pathname ?? "")) {
            n.cancel();
            goto("/login?r=" + n.to?.url?.pathname);
        }
    });

    onMount(() => {
        if ($page.data.origin != location.origin) {
            showWarning = true;
        }
    });

    let hideDemoHint = false;
    let showWarning = false;
</script>

{#if env.PUBLIC_IS_DEMO === "true" && !hideDemoHint}
    <div
        class="flex items-center justify-between bg-amber-200 text-center p-4 text-sm text-black"
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

{#if showWarning}
    <div
        class="flex items-center justify-between bg-red-200 text-center p-4 text-sm text-black"
        out:slide
    >
        <div></div>
        <p>
            You are accessing wanderer from <span class="font-mono bg-gray-100"
                >{location.origin}</span
            >
            but your ORIGIN environment variable is set to
            <span class="font-mono bg-gray-100">{$page.data.origin}</span>. This
            may cause errors.
        </p>
        <button class="btn-icon self-end" on:click={() => (showWarning = false)}
            ><i class="fa fa-close"></i></button
        >
    </div>
{/if}

<NavBar></NavBar>
<Toast></Toast>
<slot />

<Footer></Footer>
