<script lang="ts">
    import { page } from "$app/state";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import type { Trail, TrailFilter } from "$lib/models/trail.js";
    import { profile_trails_index } from "$lib/stores/profile_store.js";
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
        trails = await profile_trails_index(
            page.params.handle,
            filter,
            pagination.page,
            12,
            fetch,
        );

        loading = false;
    }

    async function paginate(newPage: number) {
        pagination.page = newPage;
        trails = await profile_trails_index(
            page.params.handle,
            filter,
            newPage,
            12,
            fetch,
        );
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
