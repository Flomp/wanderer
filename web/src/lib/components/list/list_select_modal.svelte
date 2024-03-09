<script lang="ts">
    import { createEventDispatcher } from "svelte";

    import type { List } from "$lib/models/list";
    import { lists } from "$lib/stores/list_store";
    import { trail } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";
    import Modal from "../base/modal.svelte";
    import { getFileURL } from "$lib/util/file_util";

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
    title={$_("select-list")}
    size="max-w-sm"
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
                        src={getFileURL(list, list.avatar)}
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
