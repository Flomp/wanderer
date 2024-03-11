<script lang="ts">
    import { createEventDispatcher } from "svelte";

    import { browser } from "$app/environment";
    import { listSchema, type List } from "$lib/models/list";
    import { list } from "$lib/stores/list_store";
    import { createForm } from "$lib/vendor/svelte-form-lib/index";
    import { util } from "$lib/vendor/svelte-form-lib/util";
    import { _ } from "svelte-i18n";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    import Textarea from "../base/textarea.svelte";
    import { getFileURL } from "$lib/util/file_util";
    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    let previewURL = "";

    const { form, errors, handleChange, handleSubmit } = createForm<List>({
        initialValues: $list,
        validationSchema: listSchema,
        onSubmit: async (submittedList) => {
            dispatch("save", { list: submittedList, avatar: (document.getElementById("avatar") as HTMLInputElement).files![0] });
            (document.getElementById("avatar") as HTMLInputElement).value = "";
            closeModal!();
        },
    });

    function openAvatarBrowser() {
        document.getElementById("avatar")!.click();
    }

    function handleAvatarSelection() {
        const files = (document.getElementById("avatar") as HTMLInputElement)
            .files;

        if (!files) {
            return;
        }

        previewURL = URL.createObjectURL(files[0]);
    }
    $: if (browser) {
        form.set(util.cloneDeep($list));
        previewURL = getFileURL($list, $list.avatar) ?? "";
    }
</script>

<Modal
    id="list-modal"
    title={$form.id ? $_("edit-list") : $_("new-list")}
    let:openModal
    bind:openModal
    bind:closeModal
>
    <slot {openModal} />
    <form
        id="list-form"
        slot="content"
        class="modal-content space-y-4"
        on:submit={handleSubmit}
    >
        <label for="avatar" class="text-sm font-medium block"> {$_('avatar')} </label>
        <input
            name="avatar"
            type="file"
            id="avatar"
            accept="image/*"
            style="display: none;"
            on:change={handleAvatarSelection}
        />
        <div class="flex items-center gap-4">
            {#if previewURL.length > 0}
                <img
                    class="w-32 aspect-square rounded-full object-cover"
                    alt="avatar"
                    src={previewURL}
                />
            {/if}
            <button
                class="btn-secondary"
                type="button"
                on:click={openAvatarBrowser}>{$_('change')}...</button
            >
        </div>

        <TextField
            name="name"
            label={$_("name")}
            bind:value={$form.name}
            error={$errors.name}
            on:change={handleChange}
        ></TextField>

        <Textarea
            name="description"
            label={$_("description")}
            bind:value={$form.description}
            error={$errors.description}
            on:change={handleChange}
        ></Textarea>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button class="btn-primary" type="submit" form="list-form" name="save"
            >{$_("save")}</button
        >
    </div>
</Modal>
