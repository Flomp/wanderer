<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";

    import { lists, lists_index } from "$lib/stores/list_store";
    import Modal from "../base/modal.svelte";
    import type { List } from "$lib/models/list";
    import { trail } from "$lib/stores/trail_store";

    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    function handleSelect(list: List) {
        dispatch("change", list);
        closeModal!();
    }

    function listContainsCurrentTrail(list: List) {
        return list.trails?.includes($trail.id!);
    }
</script>

<Modal
    id="list-modal"
    title="Select List"
    size="md"
    let:openModal
    bind:openModal
    bind:closeModal
>
    <slot {openModal} />

    <ul slot="content">
        {#each $lists as list}
            <li
                class="flex gap-4 items-center p-4 hover:bg-menu-item-background-hover rounded-xl transition-colors cursor-pointer"
                on:click={() => handleSelect(list)}
                role="presentation"
            >
                {#if list.avatar}
                    <img
                        class="w-12 aspect-square rounded-full"
                        src={list.avatar}
                        alt="avatar"
                    />
                {:else}
                    <div
                        class="flex w-12 aspect-square shrink-0 items-center justify-center"
                    >
                        <i class="fa fa-table-list text-5xl"></i>
                    </div>
                {/if}
                <h5 class="text-md font-semibold">{list.name}</h5>

                <i
                    class="fa fa-{listContainsCurrentTrail(list)
                        ? 'minus'
                        : 'plus'} rounded-full border border-input-border p-2"
                ></i>
            </li>
        {/each}
    </ul>
</Modal>
