<script lang="ts">
    import { page } from "$app/stores";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";

    import TextField from "$lib/components/base/text_field.svelte";
    import { settings_update } from "$lib/stores/settings_store";
    import { currentUser } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    $: settings = $page.data.settings;

    const mapFocus: SelectItem[] = [
        { text: $_("trail", { values: { n: 2 } }), value: "trails" },
        { text: $_("location"), value: "location" },
    ];

    let selectedLanguage = "en";
    let selectedMapFocus = "trails";

    let searchDropdownItems: SearchItem[] = [];
    let citySearchQuery: string = "";

    let customTilesetName: string = "";
    let customTilesetURL: string = "";
    let terrainURL: string = "";
    let hillshadingURL: string = "";

    onMount(() => {
        citySearchQuery = settings?.location?.name ?? "";
        selectedLanguage = settings?.language || "en";
        selectedMapFocus = settings?.mapFocus ?? "trails";

        terrainURL = settings?.terrain?.terrain ?? "";
        hillshadingURL = settings?.terrain?.hillshading ?? "";
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

    async function handleMapFocusSelection(value: "trails" | "location") {
        await settings_update({
            id: settings!.id,
            mapFocus: value,
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

    async function handleTerrainAdd() {
        await settings_update({
            id: settings!.id,
            terrain: { terrain: terrainURL, hillshading: hillshadingURL },
        });
    }

    $: terrainSaveEnabled =
        terrainURL !== settings?.terrain?.terrain ||
        hillshadingURL !== settings?.terrain?.hillshading;
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
<h2 class="text-2xl font-semibold">{$_("map")}</h2>
<hr class="mt-4 mb-6 border-input-border" />
{#if $currentUser}
    <div class="space-y-12">
        <div>
            <h4 class="text-xl font-medium mb-2">{$_("focus-map-on")}</h4>
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
                        clearAfterSelect={false}
                        bind:value={citySearchQuery}
                        on:update={(e) => searchCities(e.detail)}
                        on:click={(e) => handleSearchClick(e.detail)}
                    ></Search>
                </div>
            {/if}
        </div>
        <div>
            <h4 class="text-xl font-medium mb-2">{$_("tilesets")}</h4>

            {#each settings.tilesets ?? [] as tileset, i}
                <div
                    class="flex items-center justify-between px-4 py-2 border border-input-border rounded-xl mb-2"
                >
                    <div>
                        <p>{tileset.name}</p>
                        <p class="text-sm text-gray-500">{tileset.url}</p>
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
                    placeholder={$_("name")}
                ></TextField>
                <div class="flex items-center basis-full gap-2">
                    <div class="flex-grow">
                        <TextField
                            label="URL"
                            bind:value={customTilesetURL}
                            placeholder="https://.../style.json"
                        ></TextField>
                    </div>
                    <button class="btn-icon mt-6" on:click={handleTilesetAdd}
                        ><i class="fa fa-plus"></i></button
                    >
                </div>
            </div>
        </div>

        <div>
            <h4 class="text-xl font-medium mb-2">{$_("Terrain")}</h4>
            <div class="flex items-center gap-2">
                <div class="basis-full">
                    <TextField
                        label="Terrain URL"
                        bind:value={terrainURL}
                        placeholder="https://.../tiles.json"
                    ></TextField>
                </div>
                <div class="basis-full">
                    <TextField
                        label="Hillshading URL"
                        bind:value={hillshadingURL}
                        placeholder="https://.../tiles.json"
                    ></TextField>
                </div>
                <button
                    disabled={!terrainSaveEnabled}
                    class="btn-icon mt-6"
                    class:hover:!bg-background={!terrainSaveEnabled}
                    on:click={handleTerrainAdd}
                    class:text-gray-500={!terrainSaveEnabled}
                    ><i class="fa fa-save"></i></button
                >
            </div>
        </div>
    </div>
{/if}
