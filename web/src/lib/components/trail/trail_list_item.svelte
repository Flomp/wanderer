<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { getFileURL } from "$lib/util/file_util";
    import { formatDistance, formatElevation, formatTimeHHMM } from "$lib/util/format_util";

    export let trail: Trail;
</script>

<li
    class="flex gap-8 p-4 rounded-xl border border-input-border cursor-pointer hover:bg-secondary-hover transition-colors"
>
    <div class="shrink-0">
        <img
            class="h-28 w-28 object-cover rounded-xl"
            src={getFileURL(trail, trail.thumbnail)}
            alt=""
        />
    </div>
    <div class="min-w-0 basis-full">
        <h4 class="font-semibold text-lg">{trail.name}</h4>
        <h5><i class="fa fa-location-dot mr-3"></i>{trail.location}</h5>
        <div class="flex mt-2 gap-4 text-sm text-gray-500">
            <span
                ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                    trail.distance,
                )}</span
            >
            <span
                ><i class="fa fa-up-down mr-2"></i>{formatElevation(
                    trail.elevation_gain,
                )}</span
            >
            <span
                ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                    trail.duration,
                )}</span
            >
        </div>
        <p
            class="mt-3 text-sm whitespace-nowrap min-w-0 max-w-full overflow-hidden text-ellipsis"
        >
            {trail.description}
        </p>
    </div>
</li>
