<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    
    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    export let title: string = $_('confirm-deletion');
    export let text: string;
    export let action: string = "delete";
    export let id: string = "confirm-modal";

    const dispatch = createEventDispatcher();

    function confirm() {
        dispatch("confirm");
        closeModal!();
    }
</script>

<Modal {id} {title} bind:openModal bind:closeModal>
    <p slot="content">{text}</p>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button id="confirm" class={action === "delete" ? "btn-danger" : "btn-primary"} type="button" on:click={confirm} name="delete"
            >{$_(action)}</button
        >
    </div></Modal
>
