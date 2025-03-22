<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import { trails_search_filter } from "$lib/stores/trail_store";
    import type { Snapshot } from "@sveltejs/kit";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    let filterExpanded: boolean = $state(true);

    let loading: boolean = $state(true);

    let filter: TrailFilter = $state(page.data.filter);
    const pagination: { page: number; totalPages: number } = $state({
        page: page.url.searchParams.has("page")
            ? parseInt(page.url.searchParams.get("page")!)
            : 1,
        totalPages: 1,
    });
    let trails: Trail[] = $state([]);

    export const snapshot: Snapshot<TrailFilter> = {
        capture: () => filter,
        restore: (value) => {
            filter = value;
            handleFilterUpdate()
        },
    };

    onMount(() => {
        if (window.innerWidth < 768) {
            filterExpanded = false;
        }
    });

    async function handleFilterUpdate() {
        loading = true;
        const response = await trails_search_filter(filter, pagination.page);
        trails = response.items;
        pagination.page = response.page;
        pagination.totalPages = response.totalPages;
        loading = false;
    }

    async function paginate(newPage: number) {
        pagination.page = newPage;
        const response = await trails_search_filter(filter, newPage);
        trails = response.items;
        page.url.searchParams.set("page", newPage.toString());
        goto(`?${page.url.searchParams.toString()}`);
    }
</script>

<svelte:head>
    <title>{$_("trail", { values: { n: 2 } })} | wanderer</title>
</svelte:head>

<main
    class="grid grid-cols-1 md:grid-cols-[300px_1fr] items-start gap-8 max-w-7xl mx-6 md:mx-auto"
>
    <TrailFilterPanel
        categories={page.data.categories}
        bind:filter
        {filterExpanded}
        onupdate={handleFilterUpdate}
    ></TrailFilterPanel>
    <TrailList
        bind:filter
        {loading}
        {trails}
        {pagination}
        onupdate={handleFilterUpdate}
        onpagination={paginate}
    ></TrailList>
</main>
