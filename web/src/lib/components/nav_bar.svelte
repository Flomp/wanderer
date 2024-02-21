<script lang="ts">
    import { goto } from "$app/navigation";
    import LogoText from "$lib/components/logo/logo_text.svelte";
    import { currentUser, logout } from "$lib/stores/user_store";
    import { tweened } from "svelte/motion";
    import Dropdown from "./base/dropdown.svelte";
    import { cubicOut } from "svelte/easing";

    const navBarItems = [
        { text: "Home", value: "/" },
        { text: "Trails", value: "/trails" },
        { text: "Map", value: "/map" },
        { text: "Category", value: "/" },
        { text: "Favorites", value: "/" },
    ];

    const dropdownItems = [
        { text: "Profile", value: "profile", icon: "user" },
        { text: "Logout", value: "logout", icon: "right-from-bracket" },
    ];

    const indicatorPosition = tweened(0, {
        duration: 300,
        easing: cubicOut,
    });

    const indicatorWidth = tweened(0, {
        duration: 300,
        easing: cubicOut,
    });

    function handleDropdownClick(item: { text: string; value: any }) {
        if (item.value == "logout") {
            logout();
            goto("/login");
        }
    }

    function setActiveItem(target: EventTarget | null) {
        if (!target) {
            return;
        }
        const element = target as HTMLElement;
        const rect = element.getBoundingClientRect();
        const padding = 16;
        indicatorWidth.set(rect.width + padding);
        indicatorPosition.set(element.offsetLeft - padding / 2);
    }
</script>

<nav class="flex justify-between items-center p-6">
    <a href="/" data-sveltekit-preload-data="tap"><LogoText></LogoText></a>
    {#if $currentUser}
        <menu class="flex gap-8 relative py-1 px-2">
            <div
                class="absolute h-full w-16 bg-slate-400 rounded-xl opacity-20 top-0"
                style="width: {$indicatorWidth}px; left: {$indicatorPosition}px"
            ></div>
            {#each navBarItems as item}
                <a
                    class="font-semibold"
                    on:click={(e) => setActiveItem(e.target)}
                    href={item.value}
                    data-sveltekit-preload-data="tap">{item.text}</a
                >
            {/each}
        </menu>
        <div class="flex gap-6 items-center">
            <a class="btn-primary btn-large" href="/trail/edit/new"
                ><i class="fa fa-plus mr-2"></i>New Trail</a
            >
            <Dropdown
                items={dropdownItems}
                on:change={(e) => handleDropdownClick(e.detail)}
            >
                <img
                    class="w-12 h-h12 rounded-full"
                    src="https://api.dicebear.com/7.x/initials/svg?seed={$currentUser.username}&backgroundType=gradientLinear"
                    alt=""
                />
            </Dropdown>
        </div>
    {:else}
        <a class="btn-primary btn-large" href="/login">Login</a>
    {/if}
</nav>
