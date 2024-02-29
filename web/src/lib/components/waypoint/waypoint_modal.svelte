<script lang="ts">
    import { Waypoint, waypointSchema } from "$lib/models/waypoint";
    import { createEventDispatcher } from "svelte";

    import { waypoint } from "$lib/stores/waypoint_store";
    import { createForm } from "$lib/vendor/svelte-form-lib/index";
    import { util } from "$lib/vendor/svelte-form-lib/util";
    import { _ } from "svelte-i18n";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    import Textarea from "../base/textarea.svelte";
    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    const { form, errors, handleChange, handleSubmit } = createForm<Waypoint>({
        initialValues: $waypoint,
        validationSchema: waypointSchema,
        onSubmit: async (submittedWaypoint) => {
            dispatch("save", submittedWaypoint);
            closeModal!();
        },
    });
    $: form.set(util.cloneDeep($waypoint));
</script>

<Modal
    id="waypoint-modal"
    title={$form.id ? $_("edit-waypoint") : $_("add-waypoint")}
    let:openModal
    bind:openModal
    bind:closeModal
>
    <slot {openModal} />
    <form
        id="waypoint-form"
        slot="content"
        class="modal-content space-y-4"
        on:submit={handleSubmit}
    >
        <div class="flex gap-4">
            <div class="basis-full">
                <TextField
                    name="name"
                    label={$_("name")}
                    bind:value={$form.name}
                    error={$errors.name}
                    on:change={handleChange}
                ></TextField>
            </div>

            <TextField
                name="icon"
                label={$_("icon")}
                bind:value={$form.icon}
                icon={$form.icon}
                error={$errors.icon}
                on:change={handleChange}
            ></TextField>
        </div>

        <Textarea
            name="description"
            label={$_("description")}
            bind:value={$form.description}
            error={$errors.description}
            on:change={handleChange}
        ></Textarea>
        <div class="flex gap-4">
            <TextField
                name="lat"
                label={$_("latitude")}
                bind:value={$form.lat}
                error={$errors.lat}
                on:change={handleChange}
            ></TextField>
            <TextField
                name="lon"
                label={$_("longitude")}
                bind:value={$form.lon}
                error={$errors.lat}
                on:change={handleChange}
            ></TextField>
        </div>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button class="btn-primary" type="submit" form="waypoint-form"
            >{$_("save")}</button
        >
    </div>
</Modal>
