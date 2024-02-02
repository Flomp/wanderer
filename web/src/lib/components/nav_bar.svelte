<script lang="ts">
    import { goto } from "$app/navigation";
    import LogoText from "$lib/components/logo/logo_text.svelte";
    import { logout, currentUser } from "$lib/stores/user_store";
    import Dropdown from "./base/dropdown.svelte";

    const dropdownItems = [
        { text: "Profile", value: "profile", icon: "user" },
        { text: "Logout", value: "logout", icon: "right-from-bracket" },
    ];

    function handleDropdownClick(item: { text: string; value: any }) {
        if (item.value == "logout") {
            logout();
            goto("/login");
        }
    }
</script>

<nav class="flex justify-between items-center p-6">
    <a href="/"><LogoText></LogoText></a>
    {#if $currentUser}
        <menu class="flex gap-8">
            <a class="font-semibold" href="">Trails</a>
            <a class="font-semibold" href="">Map</a>
            <a class="font-semibold" href="">Categories</a>
            <a class="font-semibold" href="">Favorites</a>
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
