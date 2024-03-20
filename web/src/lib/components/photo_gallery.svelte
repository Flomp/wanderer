<script lang="ts">
    import type { DataSource } from "photoswipe";
    import PhotoSwipeLightbox from "photoswipe/lightbox";
    import { onMount } from "svelte";

    export let photos: string[];
    export function open(idx: number = 0) {
        lightbox.loadAndOpen(idx, lightboxDataSource)
    }
    let lightbox: PhotoSwipeLightbox;
    let lightboxDataSource: DataSource;

    onMount(() => {
        lightboxDataSource = photos.map((p) => ({
            src: p,
        }));
        lightbox = new PhotoSwipeLightbox({
            dataSource: lightboxDataSource,
            pswpModule: async () => await import("photoswipe"),
        });
        lightbox.init();

        lightbox.on("beforeOpen", () => {
            const pswp = lightbox.pswp;
            const ds = pswp?.options?.dataSource;
            if (Array.isArray(ds)) {
                for (let idx = 0, len = ds.length; idx < len; idx++) {
                    const item = ds[idx];
                    const img = new Image();
                    img.onload = () => {
                        item.width = img.naturalWidth;
                        item.height = img.naturalHeight;
                        pswp?.refreshSlideContent(idx);
                    };
                    img.src = item.src as string;
                }
            }
        });
    });
</script>
