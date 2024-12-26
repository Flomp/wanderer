<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import { trails_search_filter } from "$lib/stores/trail_store";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    let filterExpanded: boolean = true;

    let loading: boolean = false;

    const filter: TrailFilter = $page.data.filter;
    const pagination: { page: number; totalPages: number } = {
        page: 1,
        totalPages: 1,
    };
    let trails: Trail[] = [];

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

    async function paginate(page: number) {
        pagination.page = page;
        const response = await trails_search_filter(filter, page);
        trails = response.items;
        $page.url.searchParams.set("page", page.toString());
        goto(`?${$page.url.searchParams.toString()}`);
    }
</script>

<svelte:head>
    <title>{$_("trail", { values: { n: 2 } })} | wanderer</title>
</svelte:head>

<main
    class="grid grid-cols-1 md:grid-cols-[300px_1fr] items-start gap-8 max-w-7xl mx-6 md:mx-auto"
>
    <TrailFilterPanel
        categories={$page.data.categories}
        {filter}
        {filterExpanded}
        on:update={() => handleFilterUpdate()}
    ></TrailFilterPanel>
    <TrailList
        {filter}
        {loading}
        {trails}
        {pagination}
        on:update={() => handleFilterUpdate()}
        on:pagination={(e) => paginate(e.detail)}
    ></TrailList>
</main>
