<script lang="ts">
    export let tabs: string[];
    export let activeTab: number;
</script>

<div class="flex gap-2 overflow-x-auto">
    {#each tabs as tab, i}
        <button class="tab" class:tab-active={activeTab == i} on:click={() => activeTab = i}>{tab}</button>
    {/each}
</div>
