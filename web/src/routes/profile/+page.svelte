<script lang="ts">
    import { goto } from "$app/navigation";
    import RadioGroup, {
        type RadioItem,
    } from "$lib/components/base/radio_group.svelte";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import {
        currentUser,
        logout,
        users_delete,
        users_update,
    } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { getFileURL } from "$lib/util/file_util";
    import { onMount } from "svelte";
    import { _, locale } from "svelte-i18n";

    const languages: SelectItem[] = [
        { text: $_("german"), value: "de" },
        { text: $_("english"), value: "en" },
        { text: $_("french"), value: "fr" },
        { text: $_("hungarian"), value: "hu" },
        { text: $_("dutch"), value: "nl" },
        { text: $_("polish"), value: "pl" },
        { text: $_("portuguese"), value: "pt" },
    ];

    const units: RadioItem[] = [
        { text: $_("metric"), value: "metric" },
        { text: $_("imperial"), value: "imperial" },
    ];

    let selectedLanguage = "en";

    let searchDropdownItems: SearchItem[] = [];
    let citySearchQuery: string = "";

    let openConfirmModal: () => void;

    onMount(() => {
        citySearchQuery = $currentUser?.location?.name ?? "";
        selectedLanguage = $currentUser?.language || "en";
    });

    async function searchCities(q: string) {
        const r = await fetch("/api/v1/search/cities500", {
            method: "POST",
            body: JSON.stringify({ q: q, options: { limit: 5 } }),
        });
        const result = await r.json();
        searchDropdownItems = result.hits.map((h: Record<string, any>) => ({
            text: h.name,
            description: `${h.division ? `${h.division} | ` : ""}${
                country_codes[h["country code"] as keyof typeof country_codes]
            }`,
            value: h,
            icon: "city",
        }));
    }

    async function handleSearchClick(item: SearchItem) {
        citySearchQuery = item.text;
        await users_update({
            id: $currentUser!.id,
            location: {
                name: item.value.name,
                lat: item.value.lat,
                lon: item.value.lon,
            },
        });
    }

    async function handleAvatarSelection() {
        const files = (
            document.getElementById("avatarInput") as HTMLInputElement
        ).files;

        if (!files || files.length == 0) {
            return;
        }

        await users_update($currentUser!, files[0]);
    }

    async function handleUnitSelection(e: RadioItem) {
        await users_update({
            id: $currentUser!.id,
            unit: e.value as "imperial" | "metric",
        });
    }

    async function handleLanguageSelection(
        value: "en" | "de" | "fr" | "hu" | "nl" | "pl" | "pt",
    ) {
        locale.set(value);
        await users_update({
            id: $currentUser!.id,
            language: value,
        });
    }

    async function deleteAccount() {
        await users_delete($currentUser!);
        logout();
        goto("/");
    }

    function openFileBrowser() {
        document.getElementById("avatarInput")!.click();
    }
</script>

<svelte:head>
    <title>{$_("profile")} | wanderer</title>
</svelte:head>
<main
    class="grid grid-cols-1 md:grid-cols-[356px_1fr] max-w-4xl mx-4 md:mx-auto gap-x-8 items-start"
>
    {#if $currentUser}
        <div
            class="flex flex-col items-center rounded-xl border border-input-border p-4"
        >
            <div
                class="rounded-full w-32 aspect-square overflow-hidden relative group"
            >
                <img
                    src={getFileURL($currentUser, $currentUser.avatar) ||
                        `https://api.dicebear.com/7.x/initials/svg?seed=${$currentUser.username}&backgroundType=gradientLinear`}
                    alt="avatar"
                />
                <button
                    class="absolute top-0 w-32 aspect-square opacity-0 group-hover:opacity-100 flex justify-center items-center bg-black/50 focus:bg-black/60 text-white cursor-pointer transition-opacity"
                    on:click={openFileBrowser}
                >
                    <i class="fa fa-pen"></i>
                </button>
                <input
                    type="file"
                    name="avatar"
                    id="avatarInput"
                    accept="image/*"
                    style="display: none;"
                    on:change={handleAvatarSelection}
                />
            </div>

            <h3 class="text-2xl mt-4 font-semibold">{$currentUser.username}</h3>
            <h5 class="font-medium">{$currentUser.email}</h5>
        </div>
        <div class="space-y-6 rounded-xl p-4">
            <h3 class="text-2xl font-semibold">{$_("settings")}</h3>

            <div>
                <h5 class="font-medium mb-1">{$_("default-location")}</h5>
                <Search
                    items={searchDropdownItems}
                    placeholder="{$_('search-cities')}..."
                    bind:value={citySearchQuery}
                    on:update={(e) => searchCities(e.detail)}
                    on:click={(e) => handleSearchClick(e.detail)}
                ></Search>
            </div>
            <div>
                <h5 class="font-medium mb-1">{$_("language")}</h5>
                <Select
                    items={languages}
                    bind:value={selectedLanguage}
                    on:change={(e) => handleLanguageSelection(e.detail)}
                ></Select>
            </div>
            <div>
                <h5 class="font-medium mb-1">{$_("units")}</h5>
                <RadioGroup
                    name="unit"
                    items={units}
                    selected={$currentUser.unit == "metric" ? 0 : 1}
                    on:change={(e) => handleUnitSelection(e.detail)}
                ></RadioGroup>
            </div>
            <hr class="border-input-border" />
            <div class="space-y-4">
                <h4 class="text-xl text-red-400">{$_("danger-zone")}</h4>
                <button class="btn-danger" on:click={openConfirmModal}
                    >{$_("delete-account")}</button
                >
            </div>
        </div>
    {/if}
</main>
<ConfirmModal
    text={$_("account-delete-confirm")}
    bind:openModal={openConfirmModal}
    on:confirm={deleteAccount}
></ConfirmModal>
