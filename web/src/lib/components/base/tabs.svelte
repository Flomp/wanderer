<script lang="ts">
    import { onMount } from "svelte";
    import { cubicOut } from "svelte/easing";
    import { tweened } from "svelte/motion";

    interface Props {
        tabs: string[];
        activeTab: number;
        extraClasses?: string;
    }

    let { tabs, activeTab = $bindable(), extraClasses = "" }: Props = $props();

    const indicatorPosition = tweened(0, {
        duration: 300,
        easing: cubicOut,
    });

    const indicatorWidth = tweened(0, {
        duration: 300,
        easing: cubicOut,
    });

    onMount(() => {
        switchTabs(activeTab ?? 0);
    });

    function switchTabs(index: number) {
        const tabs = document.getElementById("tabs");

        const childElement = tabs?.children[index + 1] as HTMLElement;
        const newWidth = childElement?.getBoundingClientRect().width ?? 0;
        const newPosition = childElement?.offsetLeft;
        indicatorWidth.set(newWidth);
        indicatorPosition.set(newPosition);

        activeTab = index;
    }
</script>

<div id="tabs" class="flex gap-2 overflow-x-auto relative {extraClasses}">
    <div
        class="absolute h-full bg-menu-item-background-hover rounded-t-lg top-0 z-0"
        style="width: {$indicatorWidth}px; left: {$indicatorPosition}px;"
    ></div>
    {#each tabs as tab, i}
        <button
            class="tab z-10"
            class:tab-active={activeTab == i}
            onclick={() => switchTabs(i)}>{tab}</button
        >
    {/each}
</div>
