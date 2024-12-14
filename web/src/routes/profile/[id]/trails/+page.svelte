<script lang="ts">
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import { trails_search_filter } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";

    export let data;

    let loading = true;

    $: pagination = {
        page: data.trails.page,
        totalPages: data.trails.totalPages,
    };

    async function handleFilterUpdate() {
        loading = true;
        data.trails = await trails_search_filter(data.filter, pagination.page);
        
        loading = false;
    }

    async function paginate(page: number) {
        pagination.page = page;
        data.trails = await trails_search_filter(data.filter, page);
    }
</script>
<svelte:head>
    <title>{$_("profile")} | wanderer</title>
</svelte:head>
<TrailList
    {pagination}
    {loading}
    fullWidthCards={true}
    trails={data.trails.items}
    filter={data.filter}
    on:update={() => handleFilterUpdate()}
    on:pagination={(e) => paginate(e.detail)}
></TrailList>
