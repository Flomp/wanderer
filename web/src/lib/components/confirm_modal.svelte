<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";

    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    export let title: string = "Confirm Deletion";
    export let text: string;

    const dispatch = createEventDispatcher()

    function confirm() {
        dispatch("confirm");
        closeModal!();
    }
</script>

<Modal id="confirm-modal" {title} bind:openModal bind:closeModal>
    <p slot="content">{text}</p>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}>Cancel</button>
        <button
            class="btn-danger"
            type="button"
            on:click={confirm}>Delete</button
        >
    </div></Modal
>
