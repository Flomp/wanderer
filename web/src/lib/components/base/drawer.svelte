<script lang="ts">
    import { browser } from "$app/environment";
    import { beforeNavigate } from "$app/navigation";
    import { backIn, backOut } from "svelte/easing";
    import { slide } from "svelte/transition";

    export let open: boolean = false;

    $: if (open && browser) {
        document.body.style.position = "fixed";
        document.body.style.top = `-${window.scrollY}px`;
    } else if (browser) {
        const scrollY = document.body.style.top;
        document.body.style.position = "";
        document.body.style.top = "";
        window.scrollTo(0, parseInt(scrollY || "0") * -1);
    }

    beforeNavigate(() => {
        open = false;
    });
</script>

{#if open}
    <div
        class="fixed drawer-overlay h-screen w-screen backdrop-blur-md z-50 bg-gray-500/50 cursor-pointer"
        on:click={() => (open = false)}
        on:keydown={(e) => {
            if (e.key == "Escape") open = false;
        }}
        role="button"
        tabindex="0"
    ></div>
    <div
        class="absolute flex flex-col overflow-y-auto w-72 h-screen bg-background right-0 top-0 z-50"
        in:slide={{ axis: "x", easing: backOut }}
        out:slide={{ axis: "x", easing: backIn }}
    >
        <slot />
    </div>
{/if}
