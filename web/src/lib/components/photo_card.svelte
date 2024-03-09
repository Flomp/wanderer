<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";

    export let src: string;
    export let isThumbnail: boolean = false;

    const dispatch = createEventDispatcher();

    function handleThumbnailClick() {
        dispatch("thumbnail", src);
    }

    function handleDeleteClick() {
        dispatch("delete", src);
    }
</script>

<div
    class="group relative h-32 w-32 rounded-xl bg-cover bg-no-repeat"
    style="background-image: url({src});"
>
    {#if isThumbnail}
        <i
            class="fa fa-file-image absolute top-2 right-2 text-primary bg-white rounded-full px-[10px] py-2 shadow-lg"
        ></i>
    {/if}
    <div
        class="flex opacity-0 group-hover:opacity-100 absolute top-0 w-full h-full bg-white/75 rounded-xl items-center justify-center gap-6 transition-all"
    >
        <button
            type="button"
            class="tooltip"
            data-title={$_("make-thumbnail")}
            on:click={handleThumbnailClick}
            ><i class="fa fa-file-image text-primary"></i></button
        >
        <button
            class="tooltip"
            data-title={$_("delete")}
            on:click={handleDeleteClick}
            type="button"><i class="fa fa-trash text-red-500"></i></button
        >
    </div>
</div>
