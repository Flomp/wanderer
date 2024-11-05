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
    import Textarea from "../base/textarea.svelte";
    import TrailPicker from "../trail/trail_picker.svelte";
    import GPX from "$lib/models/gpx/gpx";
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
            if (!$form.expand.gpx_data) {
                $form.gpx = "";
            }
            dispatch("save", submittedValues);
            closeModal!();
        },
    });
    $: form.set(util.cloneDeep($summitLog));

    $: if ($summitLog._gpx) {
        $form._gpx = $summitLog._gpx;
    }

    async function handleTrailSelection(trailData: string | null) {
        if (!trailData) {
            $form.duration = undefined;
            $form.elevation_gain = undefined;
            $form.elevation_loss = undefined;
            $form.distance = undefined;
            return;
        }
        const gpxObject = await GPX.parse(trailData);
        if (gpxObject instanceof Error) {
            throw gpxObject;
        }

        const totals = gpxObject.getTotals();

        $form.duration = totals.duration / 1000;
        $form.elevation_gain = totals.elevationGain;
        $form.elevation_loss = totals.elevationLoss;
        $form.distance = totals.distance;
    }
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
        <div class="flex">
            <Datepicker
                name="date"
                label={$_("date")}
                bind:value={$form.date}
                error={$errors.date}
                on:change={handleChange}
            ></Datepicker>
        </div>
        <div class="flex gap-4">
            <TrailPicker
                bind:trailFile={$form._gpx}
                bind:trailData={$form.expand.gpx_data}
                label={$_("trail", { values: { n: 1 } })}
                on:change={(e) => handleTrailSelection(e.detail)}
            ></TrailPicker>
            <div class="basis-full">
                <Textarea
                    name="text"
                    extraClasses="h-28"
                    label={$_("text")}
                    bind:value={$form.text}
                    error={$errors.text}
                    on:change={handleChange}
                ></Textarea>
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
