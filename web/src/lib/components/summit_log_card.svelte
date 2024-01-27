<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import { formatISODate } from "$lib/util/format_util";

    export let log: SummitLog;
</script>

<div class="p-4 my-2 border rounded-xl">
    <h5 class="font-medium mr-2">{formatISODate(log.date)}</h5>
    <span>{log.text}</span>
</div>
