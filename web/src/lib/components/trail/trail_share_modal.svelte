<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import type { Trail } from "$lib/models/trail";
    import { TrailShare } from "$lib/models/trail_share";
    import {
        shares,
        trail_share_create,
        trail_share_delete,
        trail_share_index,
        trail_share_update,
    } from "$lib/stores/trail_share_store";
    import { getFileURL } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";
    import Button from "../base/button.svelte";
    import type { SelectItem } from "../base/select.svelte";
    import Select from "../base/select.svelte";
    import UserSearch from "../user_search.svelte";

    interface Props {
        trail: Trail;
        onsave?: () => void;
    }

    let { trail, onsave }: Props = $props();

    let modal: Modal;

    export function openModal() {
        openShareModalLocal();
    }
    async function openShareModalLocal() {
        modal.openModal();
        await fetchShares();
    }

    let copyButtonText = $state($_("copy-link"));

    let sharesLoading: boolean = false;

    const permissionSelectItems: SelectItem[] = [
        { text: $_("view"), value: "view" },
        { text: $_("edit"), value: "edit" },
    ];

    function copyURLToClipboard() {
        navigator.clipboard.writeText(window.location.href);

        copyButtonText = $_("link-copied");
        setTimeout(() => (copyButtonText = $_("copy-link")), 3000);
    }

    function close() {
        onsave?.();
        modal.closeModal!();
    }

    async function shareTrail(item: SelectItem) {
        const share = new TrailShare(item.value.id, trail.id!, "view");
        await trail_share_create(share);
        fetchShares();
    }

    async function updateSharePermission(
        share: TrailShare,
        permission: "view" | "edit",
    ) {
        share.permission = permission;
        await trail_share_update(share);
    }

    async function deleteShare(share: TrailShare) {
        await trail_share_delete(share);
        fetchShares();
    }

    async function fetchShares() {
        sharesLoading = true;
        await trail_share_index({ trail: trail.id! });
        sharesLoading = false;
    }
</script>

<Modal
    id="share-modal"
    title={$_("share-this-trail")}
    size="max-w-sm overflow-visible"
    bind:this={modal}
>
    {#snippet content()}
        <div>
            <UserSearch includeSelf={false} onclick={(item) => shareTrail(item)}
            ></UserSearch>
            <h4 class="font-semibold mt-4">{$_("shared-with")}</h4>

            {#if $shares.length == 0}
                <p class="text-gray-500 text-center mt-2 text-sm">
                    {$_("trail-not-shared")}
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
                                    value={share.permission}
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
