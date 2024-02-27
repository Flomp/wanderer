<script lang="ts">
    import { afterNavigate, goto } from "$app/navigation";
    import LogoText from "$lib/components/logo/logo_text.svelte";
    import { theme, toggleTheme } from "$lib/stores/theme_store";
    import { currentUser, logout } from "$lib/stores/user_store";
    import { backInOut, cubicOut } from "svelte/easing";
    import { tweened } from "svelte/motion";
    import Dropdown from "./base/dropdown.svelte";
    import LogoTextLight from "./logo/logo_text_light.svelte";
    import Drawer from "./base/drawer.svelte";

    const navBarItems = [
        { text: "Home", value: "/" },
        { text: "Trails", value: "/trails" },
        { text: "Map", value: "/map" },
        { text: "Lists", value: "/lists" },
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

    const indicatorScale = tweened(0, {
        duration: 600,
        easing: backInOut,
    });

    let drawerOpen: boolean = false;

    afterNavigate((e) => {
        const routeId = e.to?.route.id;
        const navBarLinks = document.getElementById("nav-bar-links");
        let childPosition = -1;
        switch (routeId) {
            case "/":
                childPosition = 1;
                break;
            case "/trails":
                childPosition = 2;
                break;
            case "/map":
                childPosition = 3;
                break;
            case "/lists":
                childPosition = 4;
                break;
            default:
                break;
        }

        if (childPosition !== -1) {
            const childElement = navBarLinks?.children[
                childPosition
            ] as HTMLElement;
            const newWidth = childElement?.getBoundingClientRect().width ?? 0;
            const newPosition = childElement.offsetLeft;
            const padding = 16;
            indicatorScale.set(1);
            indicatorWidth.set(newWidth + padding);
            indicatorPosition.set(newPosition - padding / 2);
        } else {
            indicatorScale.set(0);
        }
    });

    function handleDropdownClick(item: { text: string; value: any }) {
        if (item.value == "logout") {
            logout();
            goto("/");
        }
    }
</script>

<Drawer bind:open={drawerOpen}>
    <div class="flex gap-4 items-center m-4">
        <div class="basis-full"></div>
        <button
            class="btn-icon fa-regular fa-{$theme === 'light' ? 'sun' : 'moon'}"
            on:click={() => toggleTheme()}
        ></button>
        <button
            class="btn-icon block fa fa-close float-right"
            on:click={() => (drawerOpen = false)}
        ></button>
    </div>
    <div class="flex flex-col px-12 gap-8">
        {#each navBarItems as item}
            <a
                class="font-semibold text-xl"
                href={item.value}
                data-sveltekit-preload-data="off">{item.text}</a
            >
        {/each}
    </div>
    <hr class="my-6 border-input-border" />
    <div class="flex flex-col basis-full gap-y-3">
        <a class="btn-primary btn-large text-center mx-4" href="/trail/edit/new"
            ><i class="fa fa-plus mr-2"></i>New Trail</a
        >
        {#if $currentUser}
            <div class="basis-full"></div>
            <hr class="border-input-border" />
            <div class="flex gap-4 items-center mx-4">
                <img
                    class="rounded-full w-8 aspect-square"
                    src="https://api.dicebear.com/7.x/initials/svg?seed={$currentUser.username}&backgroundType=gradientLinear"
                    alt=""
                />
                <div>
                    <p class="text-sm">{$currentUser.username}</p>
                    <p class="text-sm text-gray-500">{$currentUser.email}</p>
                </div>
                <button class="btn-icon tooltip" data-title="Logout"
                    ><i class="fa-solid fa-arrow-right-from-bracket"
                    ></i></button
                >
            </div>
        {/if}
    </div>
</Drawer>

<nav class="flex justify-between items-center p-6">
    <a href="/" data-sveltekit-preload-data="off">
        {#if $theme == "light"}
            <LogoText></LogoText>
        {:else}
            <LogoTextLight></LogoTextLight>
        {/if}
    </a>
    <menu id="nav-bar-links" class="hidden md:flex gap-8 relative py-1 px-2">
        <div
            class="absolute h-full w-16 bg-menu-item-background-hover rounded-xl top-0 z-0"
            style="width: {$indicatorWidth}px; left: {$indicatorPosition}px; scale: {$indicatorScale}"
        ></div>
        {#each navBarItems as item}
            <a
                class="font-semibold z-10"
                href={item.value}
                data-sveltekit-preload-data="off">{item.text}</a
            >
        {/each}
    </menu>
    {#if $currentUser}
        <div class="hidden md:flex gap-6 items-center">
            <button
                class="btn-icon fa-regular fa-{$theme === 'light'
                    ? 'sun'
                    : 'moon'}"
                on:click={() => toggleTheme()}
            ></button>
            <a class="btn-primary btn-large" href="/trail/edit/new"
                ><i class="fa fa-plus mr-2"></i>New Trail</a
            >
            <Dropdown
                items={dropdownItems}
                on:change={(e) => handleDropdownClick(e.detail)}
            >
                <img
                    class="rounded-full"
                    src="https://api.dicebear.com/7.x/initials/svg?seed={$currentUser.username}&backgroundType=gradientLinear"
                    alt=""
                />
            </Dropdown>
        </div>
    {:else}
        <div class="hidden md:flex items-center gap-8">
            <button
                class="btn-icon fa-regular fa-{$theme === 'light'
                    ? 'sun'
                    : 'moon'}"
                on:click={() => toggleTheme()}
            ></button>
            <a class="btn-primary btn-large" href="/login">Login</a>
        </div>
    {/if}
    <button
        class="btn-icon fa fa-bars md:hidden"
        on:click={() => (drawerOpen = !drawerOpen)}
    ></button>
</nav>
