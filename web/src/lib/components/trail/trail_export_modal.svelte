<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import Select from "../base/select.svelte";

    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;
    export let title: string = $_("export");

    const dispatch = createEventDispatcher();

    const fileFormats = [
        { text: "GPX", value: "gpx" },
        { text: "GeoJSON", value: "json" },
    ];

    const exportSettings = {
        fileFormat: "gpx",
        photos: false,
        summitLog: false,
    };

    function exportTrail() {
        dispatch("export", exportSettings);
        closeModal!();
    }
</script>

<Modal id="export-modal" {title} size="max-w-sm" bind:openModal bind:closeModal>
    <div slot="content">
        <div class="mb-3">
            <Select
                bind:value={exportSettings.fileFormat}
                items={fileFormats}
                label={$_("file-format")}
            ></Select>
        </div>
        <div class="mb-2">
            <input
                id="include-photos-checkbox"
                type="checkbox"
                bind:checked={exportSettings.photos}
                class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
            />
            <label for="include-photos-checkbox" class="ms-2 text-sm"
                >{$_("photos")}</label
            >
        </div>
        <div class="mb-2">
            <input
                id="include-summit-log-checkbox"
                type="checkbox"
                bind:checked={exportSettings.summitLog}
                class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
            />
            <label for="include-summit-log-checkbox" class="ms-2 text-sm"
                >{$_("summit-book")}</label
            >
        </div>
    </div>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button
            class="btn-primary"
            type="button"
            on:click={exportTrail}
            name="save">{$_("export")}</button
        >
    </div></Modal
>
