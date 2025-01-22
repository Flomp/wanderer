<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        title?: string;
        text: string;
        action?: string;
        id?: string;
    }

    let {
        title = $_("confirm-deletion"),
        text,
        action = "delete",
        id = "confirm-modal",
    }: Props = $props();

    let modal: Modal;

    export function openModal() {
        modal.openModal();
    }

    const dispatch = createEventDispatcher();

    function confirm() {
        dispatch("confirm");
        modal.closeModal!();
    }
</script>

<Modal {id} {title} bind:this={modal}>
    {#snippet content()}
        <p>{text}</p>
    {/snippet}
    {#snippet footer()}
        <div class="flex items-center gap-4">
            <button class="btn-secondary" onclick={() => modal.closeModal()}
                >{$_("cancel")}</button
            >
            <button
                id="confirm"
                class={action === "delete" ? "btn-danger" : "btn-primary"}
                type="button"
                onclick={confirm}
                name="delete">{$_(action)}</button
            >
        </div>
    {/snippet}</Modal
>
