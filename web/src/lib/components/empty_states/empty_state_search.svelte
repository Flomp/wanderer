<script lang="ts">
    import { _ } from "svelte-i18n";
    export let width: number = 256;
</script>

<div class="flex flex-col justify-center items-center gap-y-4">
    <img
        style="width: {width}px"
        class="rounded-full aspect-square"
        src="/imgs/empty_state.png"
        alt="Empty State showing a wanderer going into the distance"
    />

    <h3 class="text-2xl font-semibold text-center">{$_("no-results")}</h3>
</div>
