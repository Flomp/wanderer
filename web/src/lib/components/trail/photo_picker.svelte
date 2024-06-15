<script lang="ts">
    import { getFileURL, readAsDataURLAsync } from "$lib/util/file_util";
    import PhotoCard from "../photo_card.svelte";

    export let id: string;
    export let photos: string[];
    export let photoFiles: File[] | undefined;
    export let parent: { [key: string]: any };
    export let thumbnail: number = 0;
    export let showThumbnailControls: boolean = true;
    export let showExifControls: boolean = false;

    let photoPreviews: string[] = [];

    $: Promise.all(
        (photoFiles ?? []).map(async (f) => {
            return await readAsDataURLAsync(f);
        }),
    ).then((v) => {
        photoPreviews = v;
    });

    let offerUpload: boolean = false;

    function handlePhotoDragOver(e: DragEvent) {
        e.preventDefault();
        offerUpload = true;
    }

    function handlePhotoDragLeave() {
        offerUpload = false;
    }

    function handlePhotoDrop(e: DragEvent) {
        e.preventDefault();

        offerUpload = false;
        handlePhotoSelection(e.dataTransfer?.files);
    }

    function openPhotoBrowser() {
        document.getElementById(`${id}-photo-input`)!.click();
    }

    function handlePhotoSelection(files?: FileList | null) {
        if (!files) {
            files = (
                document.getElementById(`${id}-photo-input`) as HTMLInputElement
            ).files;
        }

        if (!files) {
            return;
        }

        for (const file of files) {
            if (!file.type.startsWith("image")) {
                continue;
            }
            if (!photoFiles) {
                photoFiles = [];
            }
            photoFiles = [...photoFiles, file];
        }
    }

    function makePhotoThumbnail(index: number) {
        thumbnail = index;
    }

    function handlePhotoDelete(index: number) {
        if (thumbnail == index) {
            thumbnail = 0;
        }

        if (index >= photos.length) {
            if (!photoFiles) {
                photoFiles = [];
            }
            const adjustedIndex = index - photos.length;
            photoFiles.splice(adjustedIndex, 1);
            photoPreviews.splice(adjustedIndex, 1);

            photoPreviews = [...photoPreviews];
        } else {
            photos.splice(index, 1);
            photos = [...photos];
        }
    }
</script>

<div
    class="flex gap-x-4 max-w-full overflow-x-auto shrink-0 rounded-xl {offerUpload
        ? 'outline-dashed outline-input-border'
        : ''}"
    role="dialog"
    on:dragover={handlePhotoDragOver}
    on:dragleave={handlePhotoDragLeave}
    on:drop={handlePhotoDrop}
>
    <button
        class="btn-secondary h-32 w-32 m-2 shrink-0 grow-0 basis-auto"
        type="button"
        on:click={openPhotoBrowser}><i class="fa fa-plus"></i></button
    >
    <input
        type="file"
        id="{id}-photo-input"
        accept="image/*"
        multiple={true}
        style="display: none;"
        on:change={() => handlePhotoSelection()}
    />
    {#each (photos ?? []).concat(photoPreviews) as photo, i}
        <div class="shrink-0 grow-0 basis-auto m-2">
            <PhotoCard
                src={i >= photos.length ? photo : getFileURL(parent, photo)}
                on:delete={() => handlePhotoDelete(i)}
                isThumbnail={thumbnail === i}
                on:thumbnail={() => makePhotoThumbnail(i)}
                on:exif
                {showThumbnailControls}
                {showExifControls}
            ></PhotoCard>
        </div>
    {/each}
</div>
