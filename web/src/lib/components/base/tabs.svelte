<script lang="ts">
    import { onMount } from "svelte";
    import { cubicOut } from "svelte/easing";
    import { Tween } from "svelte/motion";

    interface Props {
        tabs: string[];
        activeTab: number;
        extraClasses?: string;
    }

    let { tabs, activeTab = $bindable(), extraClasses = "" }: Props = $props();

    const indicatorPosition = new Tween(0, {
        duration: 300,
        easing: cubicOut,
    });

    const indicatorWidth = new Tween(0, {
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
        class="tab-indicator absolute h-full rounded-t-lg top-0 z-0 border-b-2 border-content"
        style="width: {indicatorWidth.current}px; left: {indicatorPosition.current}px;"
    ></div>
    {#each tabs as tab, i}
        <button
            class="tab z-10"
            class:tab-active={activeTab == i}
            onclick={() => switchTabs(i)}>{tab}</button
        >
    {/each}
</div>
