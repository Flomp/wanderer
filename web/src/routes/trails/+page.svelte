<script lang="ts">
    import { page } from "$app/stores";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import type { TrailFilter } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import { trails, trails_search_filter } from "$lib/stores/trail_store";
    import { onMount } from "svelte";


    let filterExpanded: boolean = true;

    const filter: TrailFilter = $page.data.filter;

    onMount(() => {
        if (window.innerWidth < 768) {
            filterExpanded = false;
        }
    });

    async function handleFilterUpdate() {
        await trails_search_filter(filter);
    }
</script>

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
        on:update={async () => await trails_search_filter(filter)}
    ></TrailList>
</main>
