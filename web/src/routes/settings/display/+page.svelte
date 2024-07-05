<script lang="ts">
    import { page } from "$app/stores";
    import RadioGroup, {
        type RadioItem,
    } from "$lib/components/base/radio_group.svelte";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";

    import TextField from "$lib/components/base/text_field.svelte";
    import Settings from "$lib/components/profile/settings.svelte";
    import type { Category } from "$lib/models/category.js";
    import { settings_update } from "$lib/stores/settings_store";
    import { currentUser } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    export let data;

    $: settings = $page.data.settings;

    const mapFocus: SelectItem[] = [
        { text: $_("trail", { values: { n: 2 } }), value: "trails" },
        { text: $_("location"), value: "location" },
    ];

    const categoryItems: SelectItem[] = data.categories.map((c: Category) => ({
        text: $_(c.name),
        value: c.id,
    }));

    const units: RadioItem[] = [
        { text: $_("metric"), value: "metric" },
        { text: $_("imperial"), value: "imperial" },
    ];

    let selectedLanguage = "en";
    let selectedMapFocus = "trails";

    let selectedCategory =
        $page.data.settings?.category || data.categories[0].id;

    let searchDropdownItems: SearchItem[] = [];
    let citySearchQuery: string = "";

    let customTilesetName: string = "";
    let customTilesetURL: string = "";

    onMount(() => {
        citySearchQuery = settings?.location?.name ?? "";
        selectedLanguage = settings?.language || "en";
        selectedMapFocus = settings?.mapFocus ?? "trails";
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
        await settings_update({
            id: settings?.id!,
            location: {
                name: item.value.name,
                lat: item.value.lat,
                lon: item.value.lon,
            },
        });
    }

    async function handleUnitSelection(e: RadioItem) {
        await settings_update({
            id: settings!.id,
            unit: e.value as "imperial" | "metric",
        });
    }

    async function handleMapFocusSelection(value: "trails" | "location") {
        await settings_update({
            id: settings!.id,
            mapFocus: value,
        });
    }

    async function handleCategorySelection(value: string) {
        await settings_update({
            id: settings!.id,
            category: value,
        });
    }

    async function handleTilesetAdd() {
        if (!customTilesetName || !customTilesetURL) {
            return;
        }
        await settings_update({
            id: settings!.id,
            tilesets: [
                ...(settings.tilesets ?? []),
                { name: customTilesetName, url: customTilesetURL },
            ],
        });
        customTilesetName = "";
        customTilesetURL = "";
    }

    async function handleTilesetDelete(index: number) {
        settings.tilesets.splice(index, 1);
        await settings_update({
            id: settings!.id,
            tilesets: settings.tilesets,
        });
    }
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
<main>
    {#if $currentUser}
        <Settings selected={1}>
            <div class="space-y-6">
                <h3 class="text-2xl font-semibold">{$_("units")}</h3>
                <RadioGroup
                    name="unit"
                    items={units}
                    selected={settings?.unit == "metric" ? 0 : 1}
                    on:change={(e) => handleUnitSelection(e.detail)}
                ></RadioGroup>
                <h3 class="text-2xl font-semibold">{$_("focus-map-on")}</h3>

                <Select
                    items={mapFocus}
                    bind:value={selectedMapFocus}
                    on:change={(e) => handleMapFocusSelection(e.detail)}
                ></Select>
                {#if selectedMapFocus == "location"}
                    <div class="mt-3">
                        <Search
                            items={searchDropdownItems}
                            placeholder="{$_('search-cities')}..."
                            bind:value={citySearchQuery}
                            on:update={(e) => searchCities(e.detail)}
                            on:click={(e) => handleSearchClick(e.detail)}
                        ></Search>
                    </div>
                {/if}
                <h3 class="text-2xl font-semibold">{$_("default-category")}</h3>
                <Select
                    items={categoryItems}
                    bind:value={selectedCategory}
                    on:change={(e) => handleCategorySelection(e.detail)}
                ></Select>
                <h3 class="text-2xl font-semibold">{$_("tilesets")}</h3>

                {#each settings.tilesets ?? [] as tileset, i}
                    <div
                        class="flex items-center justify-between px-4 py-2 border border-input-border rounded-xl"
                    >
                        <div>
                            <p class="font-medium">{tileset.name}</p>
                            <p class=" text-sm">{tileset.url}</p>
                        </div>
                        <button
                            class="btn-icon"
                            on:click={() => handleTilesetDelete(i)}
                            ><i class="fa fa-trash text-red-500"></i></button
                        >
                    </div>
                {/each}

                <div class="flex gap-4 items-center">
                    <TextField
                        label={$_("name")}
                        bind:value={customTilesetName}
                        placeholder="Open Street Maps"
                    ></TextField>
                    <div class="flex items-center basis-full gap-2">
                        <div class="flex-grow">
                            <TextField
                                label="URL"
                                bind:value={customTilesetURL}
                                placeholder="https://{'{'}s{'}'}.tile.openstreetmap.org/{'{'}z{'}'}/{'{'}x{'}'}/{'{'}y{'}'}.png"
                            ></TextField>
                        </div>
                        <button
                            class="btn-icon mt-6"
                            on:click={handleTilesetAdd}
                            ><i class="fa fa-plus"></i></button
                        >
                    </div>
                </div>
            </div>
        </Settings>
    {/if}
</main>
