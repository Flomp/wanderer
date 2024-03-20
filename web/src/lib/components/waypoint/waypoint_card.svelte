<script lang="ts">
    import type { Waypoint } from "$lib/models/waypoint";
    import { getFileURL, readAsDataURLAsync } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";
    import Dropdown from "../base/dropdown.svelte";
    import { browser } from "$app/environment";
    import PhotoGallery from "../photo_gallery.svelte";

    export let waypoint: Waypoint;
    export let mode: "show" | "edit" = "show";

    let openGallery: (idx?: number) => void;

    let imgSrc: string[] = [];
    $: if (waypoint.photos?.length) {
        imgSrc = waypoint.photos
            .filter((_, i) => i < 3)
            .map((p) => getFileURL(waypoint, p));
    } else if (waypoint._photos?.length && browser) {
        Promise.all(
            waypoint._photos
                .filter((_, i) => i < 3)
                .map(async (f) => {
                    return await readAsDataURLAsync(f);
                }),
        ).then((v) => {
            imgSrc = v;
        });
    } else {
        imgSrc = [];
    }

    const dropdownItems = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];
</script>

<div
    class="flex gap-4 p-4 border border-input-border rounded-md my-2 hover:bg-menu-item-background-hover"
>
    {#if imgSrc.length}
        {#if mode == "show"}
            <PhotoGallery
                photos={waypoint.photos.map((p) => getFileURL(waypoint, p))}
                bind:open={openGallery}
            ></PhotoGallery>
        {/if}
        <button
            class="relative basis-16 aspect-square ml-2 mb-3 shrink-0"
            type="button"
            on:click={mode == "show" ? () => openGallery() : undefined}
        >
            {#each imgSrc as img, i}
                <img
                    class="absolute h-full rounded-xl object-cover"
                    style="top: {6 * i}px; right: {6 *
                        i}px; transform: rotate(-{i * 5}deg)"
                    src={img}
                    alt="waypoint"
                />
            {/each}
        </button>
    {/if}
    <div class="basis-full">
        <div class="flex justify-between items-center mb-2">
            <h5>
                <i class="fa fa-{waypoint.icon} mr-2"></i>{waypoint.name}
            </h5>
            {#if mode == "edit"}
                <Dropdown items={dropdownItems} on:change></Dropdown>
            {/if}
        </div>

        {#if waypoint.description}
            <p>{waypoint.description}</p>
        {/if}

        <span class="text-sm text-gray-500"
            >{waypoint.lat.toFixed(5)}, {waypoint.lon.toFixed(5)}</span
        >
    </div>
</div>
