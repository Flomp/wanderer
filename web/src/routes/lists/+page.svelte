<script lang="ts">
    import { page } from "$app/stores";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ListCard from "$lib/components/list/list_card.svelte";
    import ListModal from "$lib/components/list/list_modal.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import { List } from "$lib/models/list";
    import type { TrailFilter } from "$lib/models/trail";
    import {
        list,
        lists,
        lists_create,
        lists_delete,
        lists_index,
        lists_update,
    } from "$lib/stores/list_store";

    let openListModal: () => void;

    let filter: TrailFilter = $page.data.filter;

    function beforeListModalOpen() {
        list.set(new List("", []));
        openListModal();
    }

    async function saveList(
        e: CustomEvent<{ list: List; formData: FormData }>,
    ) {
        const result = e.detail;
        if (result.list.id) {
            await lists_update(result.list, result.formData);
            await lists_index();
        } else {
            await lists_create(result.formData);
            await lists_index();
        }
    }

    async function handleDropdownClick(
        e: CustomEvent<DropdownItem>,
        currentList: List,
    ) {
        const item = e.detail;

        if (item.value == "edit") {
            list.set(currentList);
            openListModal();
        } else if (item.value == "delete") {
            await lists_delete(currentList);
            await lists_index();
        }
    }
</script>

<main
    class="grid grid-cols-1 md:grid-cols-[400px_1fr] gap-8 max-w-7xl mx-6 md:mx-auto min-h-screen"
>
    <ul
        class="list-list max-w-xl mx-2 md:mx-auto mt-8 rounded-xl border border-input-border  p-4"
    >
        <button
            class="flex w-full items-center gap-4 hover:bg-menu-item-background-hover transition-colors rounded-xl p-4 cursor-pointer"
            on:click={beforeListModalOpen}
        >
            <i class="fa fa-plus text-xl aspect-square"></i>
            <h5 class="text-xl font-semibold">Create new list</h5>
        </button>
        <hr class="border-separator my-2" />
        {#each $lists as item, i}
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
            <li on:click={() => list.set(item)}>
                <ListCard list={item} on:change={(e) => handleDropdownClick(e, item)} active={item.id === $list.id}
                ></ListCard>
                {#if i != $lists.length - 1}
                    <hr class="border-separator my-2" />
                {/if}
            </li>
        {/each}
    </ul>
    <TrailList
        bind:filter
        trails={$list.expand?.trails ?? []}
        on:update={async () => await lists_index()}
    ></TrailList>
    <ListModal bind:openModal={openListModal} on:save={saveList}></ListModal>
</main>
