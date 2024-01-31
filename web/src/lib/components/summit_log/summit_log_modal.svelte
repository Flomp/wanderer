<script lang="ts">
    import { createEventDispatcher } from "svelte";

    import { summitLogSchema, type SummitLog } from "$lib/models/summit_log";
    import { summitLog } from "$lib/stores/summit_log_store";
    import { createForm } from "$lib/vendor/svelte-form-lib/index";
    import Datepicker from "../base/datepicker.svelte";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    import { util } from "$lib/vendor/svelte-form-lib/util";

    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    const { form, errors, handleChange, handleSubmit } = createForm<SummitLog>({
        initialValues: $summitLog,
        validationSchema: summitLogSchema,
        onSubmit: async (submittedValues) => {           
            dispatch("save", submittedValues);
            closeModal!();
        },
    });
    $: form.set(util.cloneDeep($summitLog));
</script>

<Modal
    id="summit-log-modal"
    title="Add Entry"
    let:openModal
    bind:openModal
    bind:closeModal
>
    <slot {openModal} />
    <form
        id="summit-log-form"
        slot="content"
        class="modal-content space-y-4"
        on:submit={handleSubmit}
    >
        <div class="flex gap-4">
            <Datepicker
                name="date"
                label="Date"
                bind:value={$form.date}
                error={$errors.date}
                on:change={handleChange}
            ></Datepicker>
            <div class="basis-full">
                <TextField
                    name="text"
                    label="Text"
                    bind:value={$form.text}
                    error={$errors.text}
                    on:change={handleChange}
                ></TextField>
            </div>
        </div>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}>Cancel</button>
        <button class="btn-primary" type="submit" form="summit-log-form"
            >Save</button
        >
    </div>
</Modal>
