<script lang="ts">
    import { Waypoint, waypointSchema } from "$lib/models/waypoint";
    import { createEventDispatcher } from "svelte";

    import { createForm } from "svelte-forms-lib";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    import Textarea from "../base/textarea.svelte";
    import { waypoint } from "$lib/stores/waypoint_store";

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
    $: form.set($waypoint);
</script>

<Modal
    id="waypoint-modal"
    title="Add Waypoint"
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
                    label="Name"
                    bind:value={$form.name}
                    error={$errors.name}
                    on:change={handleChange}
                ></TextField>
            </div>

            <TextField
                label="Icon"
                bind:value={$form.icon}
                icon={$form.icon}
                error={$errors.icon}
                on:change={handleChange}
            ></TextField>
        </div>

        <Textarea
            label="Description"
            bind:value={$form.description}
            error={$errors.description}
            on:change={handleChange}
        ></Textarea>
        <div class="flex gap-4">
            <TextField
                label="Latitude"
                bind:value={$form.lat}
                error={$errors.lat}
                on:change={handleChange}
            ></TextField>
            <TextField
                label="Longitude"
                bind:value={$form.lon}
                error={$errors.lat}
                on:change={handleChange}
            ></TextField>
        </div>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}>Cancel</button>
        <button class="btn-primary" type="submit" form="waypoint-form"
            >Save</button
        >
    </div>
</Modal>
