<script lang="ts">
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import { ms } from "$lib/meilisearch";
    import { currentUser } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import RadioGroup, {
        type RadioItem,
    } from "$lib/components/base/radio_group.svelte";

    const languages: SelectItem[] = [
        { text: "English", value: "en" },
        { text: "German", value: "de" },
    ];

    const units: RadioItem[] = [
        { text: "Metric", value: "metric" },
        { text: "Imperial", value: "imperial" },
    ];

    let selectedLanguage = "en";

    let searchDropdownItems: SearchItem[] = [];
    let citySearchQuery: string = "";

    async function searchCities(q: string) {
        const result = await ms.index("cities500").search(q, { limit: 5 });
        searchDropdownItems = result.hits.map((h) => ({
            text: h.name,
            description:
                country_codes[h["country code"] as keyof typeof country_codes],
            value: h,
            icon: "city",
        }));
    }

    async function handleSearchClick(item: SearchItem) {}
</script>

<main
    class="grid grid-cols-1 md:grid-cols-[356px_1fr] max-w-4xl mx-4 md:mx-auto gap-x-8 items-start"
>
    {#if $currentUser}
        <div
            class="flex flex-col items-center rounded-xl border border-input-border p-4"
        >
            <img
                class="rounded-full w-32 aspect-square"
                src={$currentUser.avatar ||
                    `https://api.dicebear.com/7.x/initials/svg?seed=${$currentUser.username}&backgroundType=gradientLinear`}
                alt=""
            />
            <h3 class="text-2xl mt-4 font-semibold">{$currentUser.username}</h3>
            <h5 class="font-medium">{$currentUser.email}</h5>
        </div>
        <div class="space-y-6 rounded-xl p-4">
            <h3 class="text-2xl font-semibold">Settings</h3>

            <div>
                <h5 class="font-medium mb-1">Default Location</h5>
                <Search
                    items={searchDropdownItems}
                    placeholder="Search cities..."
                    bind:value={citySearchQuery}
                    on:update={(e) => searchCities(e.detail)}
                    on:click={(e) => handleSearchClick(e.detail)}
                ></Search>
            </div>
            <div>
                <h5 class="font-medium mb-1">Language</h5>
                <Select items={languages} bind:value={selectedLanguage}
                ></Select>
            </div>
            <div>
                <h5 class="font-medium mb-1">Units</h5>
                <RadioGroup
                    name="unit"
                    items={units}
                    selected={0}
                    on:change={() => {}}
                ></RadioGroup>
            </div>
            <div class="space-y-2">
                <h5 class="font-medium mb-1">Set new password</h5>
                <TextField placeholder="New password" label="" type="password"
                ></TextField>
                <TextField placeholder="Repeat password" type="password"
                ></TextField>
            </div>
            <hr class="border-input-border" />
            <div class="space-y-4">
                <h4 class="text-xl text-red-400">Danger zone</h4>
                <button class="btn-danger">Delete Account</button>
            </div>
        </div>
    {/if}
</main>
