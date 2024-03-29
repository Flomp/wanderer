<script lang="ts">
    import { createEventDispatcher } from "svelte";

    import { type SummitLog } from "$lib/models/summit_log";
    import { summitLog } from "$lib/stores/summit_log_store";
    import { createForm } from "$lib/vendor/svelte-form-lib/index";
    import { util } from "$lib/vendor/svelte-form-lib/util";
    import { _ } from "svelte-i18n";
    import { date, object, string } from "yup";
    import Datepicker from "../base/datepicker.svelte";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    const summitLogSchema = object<SummitLog>({
        id: string().optional(),
        date: date().required("Required").typeError($_("invalid-date")),
        text: string().optional(),
    });

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
    title={$form.id ? $_("edit-entry") : $_("add-entry")}
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
                label={$_("date")}
                bind:value={$form.date}
                error={$errors.date}
                on:change={handleChange}
            ></Datepicker>
            <div class="basis-full">
                <TextField
                    name="text"
                    label={$_("text")}
                    bind:value={$form.text}
                    error={$errors.text}
                    on:change={handleChange}
                ></TextField>
            </div>
        </div>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button class="btn-primary" type="submit" form="summit-log-form"
            >{$_("save")}</button
        >
    </div>
</Modal>
