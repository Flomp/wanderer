<script lang="ts">
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import type { Trail, TrailFilter } from "$lib/models/trail.js";
    import { trails_search_filter } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";

    let { data } = $props();

    let loading = $state(true);

    let pagination = $state({
        page: data.trails.page,
        totalPages: data.trails.totalPages,
    });

    let trails = $state(data.trails);

    let filter: TrailFilter = $state(data.filter);

    async function handleFilterUpdate() {
        loading = true;
        trails = await trails_search_filter(filter, pagination.page);

        loading = false;
    }

    async function paginate(page: number) {
        pagination.page = page;
        trails = await trails_search_filter(filter, page);
    }
</script>

<svelte:head>
    <title>{$_("profile")} | wanderer</title>
</svelte:head>
<TrailList
    {pagination}
    {loading}
    fullWidthCards={true}
    trails={trails.items}
    {filter}
    onupdate={handleFilterUpdate}
    onpagination={paginate}
></TrailList>
