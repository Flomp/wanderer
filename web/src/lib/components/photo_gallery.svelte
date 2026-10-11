<script lang="ts">
    import { isVideoURL } from "$lib/util/file_util";
    import PhotoSwipeVideoPlugin from "$lib/vendor/photo-swipe-video-plugin";
    import type { DataSource } from "photoswipe";
    import PhotoSwipeLightbox from "photoswipe/lightbox";
    import { pushState } from "$app/navigation";
    import { page } from "$app/state";
    import { onMount, onDestroy } from "svelte";

    interface Props {
        photos: string[];
        open?: (idx: number) => void;
    }

    type SlideItem = {
        src?: string;
        type?: string;
        videoSrc?: string;
        width?: number;
        height?: number;
    };

    let { photos }: Props = $props();

    let lightbox: PhotoSwipeLightbox;
    let lightboxDataSource: DataSource;

    // An open lightbox is a shallow history entry, so Back closes it and
    // Forward reopens it. The id keeps several galleries on a page apart.
    const galleryId = $props.id();
    let ownsHistoryEntry = false;

    $effect(() => {
        const entry = page.state.lightbox;
        if (!lightbox) {
            return;
        }
        if (entry?.id === galleryId) {
            if (!lightbox.pswp) {
                openGallery(entry.index);
            }
        } else if (lightbox.pswp) {
            ownsHistoryEntry = false;
            lightbox.pswp.close();
        }
    });

    function hasDimensions(item: SlideItem) {
        return (
            typeof item.width === "number" &&
            item.width > 0 &&
            typeof item.height === "number" &&
            item.height > 0
        );
    }

    function loadItemDimensions(item: SlideItem): Promise<void> {
        if (hasDimensions(item)) {
            return Promise.resolve();
        }

        if (item.type === "video") {
            return new Promise((resolve) => {
                const v = document.createElement("video");
                const finish = () => {
                    if (v.videoWidth > 0 && v.videoHeight > 0) {
                        item.width = v.videoWidth;
                        item.height = v.videoHeight;
                    }
                    resolve();
                };
                v.addEventListener("loadedmetadata", finish, { once: true });
                v.addEventListener("error", () => resolve(), { once: true });
                v.preload = "metadata";
                v.src = item.videoSrc as string;
            });
        }

        return new Promise((resolve) => {
            const img = new Image();
            img.onload = () => {
                item.width = img.naturalWidth;
                item.height = img.naturalHeight;
                resolve();
            };
            img.onerror = () => resolve();
            img.src = item.src as string;
        });
    }

    function ensureSlideDimensions(
        ds: SlideItem[],
        preferIdx?: number,
        refresh?: ((idx: number) => void) | null,
    ) {
        const order = [...ds.keys()];
        if (
            typeof preferIdx === "number" &&
            preferIdx >= 0 &&
            preferIdx < order.length
        ) {
            order.splice(preferIdx, 1);
            order.unshift(preferIdx);
        }

        for (const idx of order) {
            const item = ds[idx];
            void loadItemDimensions(item).then(() => {
                if (hasDimensions(item) && refresh) {
                    refresh(idx);
                }
            });
        }
    }

    export async function openGallery(idx: number = 0) {
        if (!lightbox || !Array.isArray(lightboxDataSource)) {
            return;
        }

        const items = lightboxDataSource as SlideItem[];
        const current = items[idx];
        if (current) {
            await loadItemDimensions(current);
        }

        // Warm the rest in the background so swipe does not stretch.
        ensureSlideDimensions(items, idx);

        lightbox.loadAndOpen(idx, lightboxDataSource);
    }

    onMount(() => {
        lightboxDataSource = photos.map((p) => {
            if (isVideoURL(p)) {
                return {
                    type: "video",
                    videoSrc: p,
                };
            }
            return {
                src: p,
            };
        });
        lightbox = new PhotoSwipeLightbox({
            dataSource: lightboxDataSource,
            pswpModule: async () => await import("photoswipe"),
            padding: { top: 20, bottom: 40, left: 20, right: 20 },
        });
        const videoPlugin = new PhotoSwipeVideoPlugin(lightbox);

        lightbox.on("beforeOpen", () => {
            const pswp = lightbox.pswp;

            if (page.state.lightbox?.id !== galleryId) {
                pushState("", {
                    lightbox: { id: galleryId, index: pswp?.options.index ?? 0 },
                });
            }
            ownsHistoryEntry = true;

            const ds = pswp?.options?.dataSource;

            if (Array.isArray(ds)) {
                ensureSlideDimensions(
                    ds as SlideItem[],
                    pswp?.currIndex,
                    (i) => pswp?.refreshSlideContent(i),
                );
            }
        });

        // Closed from the UI: drop the entry it added. Closing via Back has
        // already removed it.
        lightbox.on("close", () => {
            if (ownsHistoryEntry && page.state.lightbox?.id === galleryId) {
                history.back();
            }
            ownsHistoryEntry = false;
        });

        lightbox.init();
    });

    onDestroy(() => {
        // Unmounting must not navigate, so give up the entry before closing.
        ownsHistoryEntry = false;
        lightbox?.destroy();
    });
</script>
