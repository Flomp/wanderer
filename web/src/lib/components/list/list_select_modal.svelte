<script lang="ts">
    import { createEventDispatcher, type Snippet } from "svelte";

    import type { List } from "$lib/models/list";
    import { trail } from "$lib/stores/trail_store";
    import { getFileURL } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";
    import Modal from "../base/modal.svelte";
    import { theme } from "$lib/stores/theme_store";
    import emptyStateTrailDark from "$lib/assets/svgs/empty_states/empty_state_trail_dark.svg";
    import emptyStateTrailLight from "$lib/assets/svgs/empty_states/empty_state_trail_light.svg";

    interface Props {
        lists: List[];
        children?: Snippet<[any]>;
    }

    let { lists, children }: Props = $props();

    let modal: Modal;

    export function openModal() {
        modal.openModal();
    }

    const dispatch = createEventDispatcher();

    function handleSelect(list: List) {
        dispatch("change", list);
        modal.closeModal!();
    }

    function listContainsCurrentTrail(list: List) {
        return list.trails?.includes($trail.id!);
    }

    const children_render = $derived(children);
</script>

<Modal
    id="list-modal"
    title={$_("select-list")}
    size="max-w-sm"
    bind:this={modal}
>
    {#snippet children({ openModal })}
        {@render children_render?.({ openModal })}
    {/snippet}
    {#snippet content()}
        <ul>
            {#each lists as list}
                <li
                    class="flex gap-4 items-center p-4 hover:bg-menu-item-background-hover rounded-xl transition-colors cursor-pointer"
                    onclick={() => handleSelect(list)}
                    role="presentation"
                >
                    <img
                        class="w-12 aspect-square rounded-full"
                        src={list.avatar
                            ? getFileURL(list, list.avatar)
                            : $theme === "light"
                              ? emptyStateTrailLight
                              : emptyStateTrailDark}
                        alt="avatar"
                    />

                    <h5 class="text-md font-semibold">{list.name}</h5>

                    <i
                        class="fa fa-{listContainsCurrentTrail(list)
                            ? 'minus'
                            : 'plus'} rounded-full border border-input-border p-2"
                    ></i>
                </li>
            {/each}
        </ul>
    {/snippet}
</Modal>
