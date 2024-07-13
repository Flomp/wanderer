<script lang="ts">
    import { afterNavigate, goto } from "$app/navigation";
    import LogoText from "$lib/components/logo/logo_text.svelte";
    import { theme, toggleTheme } from "$lib/stores/theme_store";
    import { currentUser, logout } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import { backInOut, cubicOut } from "svelte/easing";
    import { tweened } from "svelte/motion";
    import Drawer from "./base/drawer.svelte";
    import Dropdown from "./base/dropdown.svelte";
    import LogoTextLight from "./logo/logo_text_light.svelte";
    import { _, format } from "svelte-i18n";

    let navBarItems = [
        { text: "Home", value: "/" },
        { text: $_("trail", { values: { n: 2 } }), value: "/trails" },
        { text: $_("map"), value: "/map" },
    ];

    const dropdownItems = [
        { text: $_("settings"), value: "settings", icon: "cog" },
        { text: $_("logout"), value: "logout", icon: "right-from-bracket" },
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
            window.location.href = "/";
        } else if (item.value == "settings") {
            goto("/settings/account");
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
        {#if $currentUser}
            <a
                class="font-semibold text-xl"
                href="/lists"
                data-sveltekit-preload-data="off"
                >{$_("list", { values: { n: 2 } })}</a
            >
        {/if}
    </div>
    <hr class="my-6 border-input-border" />
    <div class="flex flex-col basis-full">
        {#if $currentUser}
            <a
                class="btn-primary btn-large text-center mx-4"
                href="/trail/edit/new"
                ><i class="fa fa-plus mr-2"></i>{$_("new-trail")}</a
            >
            <div class="basis-full"></div>
            <hr class="border-input-border" />
            <div class="flex gap-4 items-center m-4">
                <a href="/profile">
                    <img
                        class="rounded-full w-10 aspect-square"
                        src={getFileURL($currentUser, $currentUser.avatar) ||
                            `https://api.dicebear.com/7.x/initials/svg?seed=${$currentUser.username}&backgroundType=gradientLinear`}
                        alt="avatar"
                    />
                </a>
                <a href="/profile">
                    <p class="text-sm">{$currentUser.username}</p>
                    <p class="text-sm text-gray-500">
                        {$currentUser.email}
                    </p>
                </a>
                <button
                    on:click={() => {
                        logout();
                        window.location.href = "/";
                    }}
                    class="btn-icon"
                    ><i class="fa-solid fa-arrow-right-from-bracket"
                    ></i></button
                >
            </div>
        {:else}
            <a class="btn-primary btn-large text-center mx-4" href="/login"
                >{$_("login")}</a
            >
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
    <menu id="nav-bar-links" class="hidden lg:flex gap-8 relative py-1 px-2">
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
        {#if $currentUser}
            <a
                class="font-semibold z-10"
                href="/lists"
                data-sveltekit-preload-data="off"
                >{$_("list", { values: { n: 2 } })}</a
            >
        {/if}
    </menu>
    {#if $currentUser}
        <div class="hidden lg:flex gap-6 items-center">
            <button
                class="btn-icon fa-regular fa-{$theme === 'light'
                    ? 'sun'
                    : 'moon'}"
                on:click={() => toggleTheme()}
            ></button>
            <a class="btn-primary btn-large" href="/trail/edit/new"
                ><i class="fa fa-plus mr-2"></i>{$_("new-trail")}</a
            >
            <Dropdown
                items={dropdownItems}
                on:change={(e) => handleDropdownClick(e.detail)}
                let:toggleMenu={openDropdown}
            >
                <div class="flex items-center">
                    <button
                        class="rounded-full bg-white text-black hover:bg-gray-200 focus:ring-4 ring-gray-100/50 transition-colors h-10 aspect-square"
                        on:click={openDropdown}
                    >
                        <img
                            class="rounded-full w-full h-full"
                            src={getFileURL(
                                $currentUser,
                                $currentUser.avatar,
                            ) ||
                                `https://api.dicebear.com/7.x/initials/svg?seed=${$currentUser.username}&backgroundType=gradientLinear`}
                            alt="avatar"
                        />
                    </button>
                </div>
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
            <a class="btn-primary btn-large" href="/login">{$_("login")}</a>
        </div>
    {/if}
    <button
        class="btn-icon fa fa-bars lg:hidden"
        on:click={() => (drawerOpen = !drawerOpen)}
    ></button>
</nav>
