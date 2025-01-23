<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        src: string;
        isThumbnail?: boolean;
        showThumbnailControls?: boolean;
        showExifControls?: boolean;
    }

    let {
        src,
        isThumbnail = false,
        showThumbnailControls = true,
        showExifControls = false,
    }: Props = $props();

    const dispatch = createEventDispatcher();

    function handleThumbnailClick() {
        dispatch("thumbnail", src);
    }

    function handleExifClick() {
        dispatch("exif", src);
    }

    function handleDeleteClick() {
        dispatch("delete", src);
    }
</script>

<div
    class="group relative h-32 w-32 rounded-xl bg-cover bg-no-repeat"
    style="background-image: url({src});"
>
    {#if isThumbnail && showThumbnailControls}
        <i
            class="fa fa-file-image absolute top-2 right-2 text-primary bg-white rounded-full px-[10px] py-2 shadow-lg"
        ></i>
    {/if}
    <div
        class="flex opacity-0 group-hover:opacity-100 absolute top-0 w-full h-full bg-secondary-hover/75 rounded-xl items-center justify-center gap-6 transition-all"
    >
        {#if showThumbnailControls}
            <button
                aria-label="Make thumbnail"
                type="button"
                class="tooltip"
                data-title={$_("make-thumbnail")}
                onclick={handleThumbnailClick}
                ><i class="fa fa-file-image text-content"></i></button
            >
        {/if}
        {#if showExifControls}
            <button
                aria-label="Get EXIF coordinates"
                type="button"
                class="tooltip"
                data-title={$_("get-position-from-exif")}
                onclick={handleExifClick}
                ><i class="fa fa-magnifying-glass-location text-content"
                ></i></button
            >
        {/if}
        <button
            aria-label="Delete photo"
            class="tooltip"
            data-title={$_("delete")}
            onclick={handleDeleteClick}
            type="button"><i class="fa fa-trash text-red-500"></i></button
        >
    </div>
</div>
