<script lang="ts">
    import { page } from "$app/stores";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
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
    import { _ } from "svelte-i18n";

    let openListModal: () => void;
    let openConfirmModal: () => void;

    let filter: TrailFilter = $page.data.filter;

    let listToBeDeleted: List | null = null;

    function beforeListModalOpen() {
        list.set(new List("", []));
        openListModal();
    }

    async function saveList(e: CustomEvent<{ list: List; avatar?: File }>) {
        const result = e.detail;
        if (result.list.id) {
            await lists_update(result.list, result.avatar);
        } else {
            await lists_create(result.list, result.avatar);
        }
        await lists_index();
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
            openConfirmModal();
            listToBeDeleted = currentList;
        }
    }

    async function deleteList() {
        if (!listToBeDeleted) {
            return;
        }
        await lists_delete(listToBeDeleted);
        await lists_index();
        listToBeDeleted = null;
    }
</script>

<svelte:head>
    <title>{$_("list", { values: { n: 2 } })} | wanderer</title>
</svelte:head>
<main
    class="grid grid-cols-1 md:grid-cols-[400px_1fr] gap-8 max-w-7xl mx-4 md:mx-auto"
    style="min-height: calc(100vh - 124px)"
>
    <ul
        class="list-list max-w-xl mx-2 md:mx-auto mt-8 rounded-xl border border-input-border p-4"
    >
        <button
            class="flex w-full items-center gap-4 hover:bg-menu-item-background-hover transition-colors rounded-xl p-4 cursor-pointer"
            on:click={beforeListModalOpen}
        >
            <i class="fa fa-plus text-xl aspect-square"></i>
            <h5 id="create-list-button" class="text-xl font-semibold">
                {$_("create-new-list")}
            </h5>
        </button>
        <hr class="border-separator my-2" />
        {#each $lists as item, i}
            <li class="list-list-item" on:click={() => list.set(item)} role="presentation">
                <ListCard
                    list={item}
                    on:change={(e) => handleDropdownClick(e, item)}
                    active={item.id === $list.id}
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
    <ConfirmModal
        text={$_("delete-list-confirm")}
        bind:openModal={openConfirmModal}
        on:confirm={deleteList}
    ></ConfirmModal>
</main>
