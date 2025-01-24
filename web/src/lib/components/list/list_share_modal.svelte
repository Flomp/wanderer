<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import type { List } from "$lib/models/list";
    import { ListShare } from "$lib/models/list_share";
    import { TrailShare } from "$lib/models/trail_share";
    import {
        list_share_create,
        list_share_delete,
        list_share_index,
        list_share_update,
        shares,
    } from "$lib/stores/list_share_store";
    import {
        trail_share_create,
        trail_share_delete,
        trail_share_index,
    } from "$lib/stores/trail_share_store";
    import { getFileURL } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";
    import Button from "../base/button.svelte";
    import type { SelectItem } from "../base/select.svelte";
    import Select from "../base/select.svelte";
    import UserSearch from "../user_search.svelte";

    interface Props {
        list: List;
        onsave?: () =>  void,
        onupdate?: (list: List) => void
    }

    let {
        list,
        onsave,
        onupdate
    }: Props = $props();

    let modal: Modal;

    export function openModal() {
        openShareModal();
    }

    async function openShareModal() {
        await fetchShares();
        modal.openModal();
    }

    let copyButtonText = $state($_("copy-link"));

    let sharesLoading: boolean = false;

    const permissionSelectItems: SelectItem[] = [
        { text: $_("view"), value: "view" },
        { text: $_("edit"), value: "edit" },
    ];

    function copyURLToClipboard() {
        navigator.clipboard.writeText(
            window.location.href.split("?")[0] + "?list=" + list.id,
        );

        copyButtonText = $_("link-copied");
        setTimeout(() => (copyButtonText = $_("copy-link")), 3000);
    }

    function close() {
        onsave?.()
        modal.closeModal!();
    }

    async function shareTrails(userId: string) {
        const existingTrailShares = await trail_share_index({ user: userId });
        const trailIds = existingTrailShares.map((s) => s.trail);
        for (const trailId of list.trails ?? []) {
            if (!trailIds.includes(trailId)) {
                const share = new TrailShare(userId, trailId, "view");
                await trail_share_create(share);
            }
        }
    }

    async function shareList(item: SelectItem) {
        const share = new ListShare(item.value.id, list.id!, "view");
        await list_share_create(share);
        await shareTrails(item.value.id);
        fetchShares();
    }

    async function updateSharePermission(
        share: ListShare,
        permission: "view" | "edit",
    ) {
        share.permission = permission;
        await list_share_update(share);
    }

    async function deleteTrailShares(userId: string) {
        const existingTrailShares = await trail_share_index({ user: userId });
        for (const trailId of list.trails ?? []) {
            const shareToDelete = existingTrailShares.find(
                (s) => s.trail == trailId,
            );
            if (shareToDelete) {
                await trail_share_delete(shareToDelete);
            }
        }
    }

    async function deleteShare(share: ListShare) {
        await list_share_delete(share);
        // await deleteTrailShares(share.user);
        fetchShares();
    }

    async function fetchShares() {
        sharesLoading = true;
        const fetchedShares = await list_share_index(list.id!);
        list.expand = {
            trails: list.expand?.trails ?? [],
            list_share_via_list: fetchedShares.items,
        };
        sharesLoading = false;
        onupdate?.(list);
    }
</script>

<Modal
    id="share-modal"
    title={$_("share-this-list")}
    size="max-w-sm overflow-visible"
    bind:this={modal}
>
    {#snippet content()}
        <div>
            <p class="p-4 bg-amber-100 rounded-xl mb-4 text-sm text-gray-500">
                {$_("list-share-warning")}
            </p>
            <UserSearch
                includeSelf={false}
                onclick={(item) => shareList(item)}
            ></UserSearch>
            <h4 class="font-semibold mt-4">{$_("shared-with")}</h4>

            {#if $shares.length == 0}
                <p class="text-gray-500 text-center mt-2 text-sm">
                    {$_("list-not-shared")}
                </p>
            {:else}
                {#each $shares as share}
                    {#if share.expand}
                        <div class="flex items-center gap-x-2 p-2">
                            <img
                                class="rounded-full w-8 aspect-square mr-2"
                                src={getFileURL(
                                    share.expand.user,
                                    share.expand.user.avatar,
                                ) ||
                                    `https://api.dicebear.com/7.x/initials/svg?seed=${share.expand.user.username}&backgroundType=gradientLinear`}
                                alt="avatar"
                            />
                            <p>{share.expand.user.username}</p>
                            <span
                                class="basis-full text-sm text-center text-gray-500"
                                >{$_("can")}</span
                            >
                            <div class="shrink-0">
                                <Select
                                    bind:value={share.permission}
                                    items={permissionSelectItems}
                                    onchange={(value) =>
                                        updateSharePermission(share, value)}
                                ></Select>
                            </div>

                            <button
                                aria-label="Delete share"
                                class="btn-icon text-red-500"
                                onclick={() => deleteShare(share)}
                                ><i class="fa fa-trash"></i></button
                            >
                        </div>
                    {/if}
                {/each}
            {/if}
        </div>
    {/snippet}
    {#snippet footer()}
        <div class="flex justify-between items-center gap-4">
            <Button
                secondary={true}
                disabled={copyButtonText == $_("link-copied")}
                onclick={copyURLToClipboard}
            >
                <i class="fa fa-link mr-2"></i>
                {copyButtonText}
            </Button>
            <button class="btn-primary" onclick={close}>{$_("close")}</button>
        </div>
    {/snippet}</Modal
>
