<script lang="ts">
    import { page } from "$app/stores";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import type { TrailFilter } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import { trails, trails_search_filter } from "$lib/stores/trail_store";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    let filterExpanded: boolean = true;

    const filter: TrailFilter = $page.data.filter;
    const pagination: { page: number; totalPages: number } =
        $page.data.pagination;

    onMount(() => {
        if (window.innerWidth < 768) {
            filterExpanded = false;
        }
    });

    async function handleFilterUpdate() {
        const response = await trails_search_filter(filter, 1);
        pagination.page = response.page;
        pagination.totalPages = response.totalPages;
    }

    async function paginate(page: number) {
        pagination.page = page;
        const response = await trails_search_filter(filter, page);
    }
</script>

<svelte:head>
    <title>{$_("trail", { values: { n: 2 } })} | wanderer</title>
</svelte:head>

<main
    class="grid grid-cols-1 md:grid-cols-[300px_1fr] gap-8 max-w-7xl mx-6 md:mx-auto"
>
    <TrailFilterPanel
        categories={$categories}
        {filter}
        {filterExpanded}
        on:update={() => handleFilterUpdate()}
    ></TrailFilterPanel>
    <TrailList
        {filter}
        trails={$trails}
        {pagination}
        on:update={() => handleFilterUpdate()}
        on:pagination={(e) => paginate(e.detail)}
    ></TrailList>
</main>
