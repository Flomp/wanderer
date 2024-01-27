<script lang="ts">
    import type { Waypoint } from "$lib/models/waypoint";

    export let waypoint: Waypoint;
</script>

<div class="p-4 border rounded-md my-2 cursor-pointer hover:bg-gray-100">
    <h5 class="mb-2"><i class="fa fa-{waypoint.icon} mr-2"></i>{waypoint.name}</h5>
    <p>{waypoint.description}</p>
    <span class="text-sm text-gray-500">{waypoint.lat}, {waypoint.lon}</span>
</div>
