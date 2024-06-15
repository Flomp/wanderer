<script lang="ts">
    import { Waypoint, waypointSchema } from "$lib/models/waypoint";
    import { createEventDispatcher } from "svelte";

    import { waypoint } from "$lib/stores/waypoint_store";
    import { icons } from "$lib/util/icon_util";
    import EXIF from "$lib/vendor/exif-js/exif";
    import { createForm } from "$lib/vendor/svelte-form-lib/index";
    import { util } from "$lib/vendor/svelte-form-lib/util";
    import { _ } from "svelte-i18n";
    import Combobox from "../base/combobox.svelte";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    import Textarea from "../base/textarea.svelte";
    import PhotoPicker from "../trail/photo_picker.svelte";
    import { convertDMSToDD } from "$lib/util/leaflet_util";
    import { show_toast } from "$lib/stores/toast_store";

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

    $: filteredIcons =
        ($form.icon?.length ?? 0) > 2
            ? icons
                  .filter((i) => i.includes($form.icon ?? ""))
                  .map((i) => ({ text: i, value: i, icon: i }))
            : [];

    function getCoordinatesFromPhoto(src: string) {
        EXIF.getData({ src: src }, function (p) {
            const lat = EXIF.getTag(p, "GPSLatitude");
            const latDir = EXIF.getTag(p, "GPSLatitudeRef");
            const lon = EXIF.getTag(p, "GPSLongitude");
            const lonDir = EXIF.getTag(p, "GPSLongitudeRef");

            if (lat && lon) {
                $form.lat = convertDMSToDD(lat, latDir);
                $form.lon = convertDMSToDD(lon, lonDir);
            } else {
                show_toast({
                    text: "No GPS data in image",
                    icon: "close",
                    type: "error",
                });
            }
        });
    }
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
            <div class="basis-2/3">
                <TextField
                    name="name"
                    label={$_("name")}
                    bind:value={$form.name}
                    error={$errors.name}
                    on:change={handleChange}
                ></TextField>
            </div>

            <Combobox
                name="icon"
                bind:value={$form.icon}
                icon={$form.icon}
                items={filteredIcons}
                label={$_("icon")}
                on:change={handleChange}
            ></Combobox>
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
        <div>
            <label for="trail-photo-input" class="text-sm font-medium pb-1">
                {$_("photos")}
            </label>
            <PhotoPicker
                id="waypoint"
                parent={$form}
                on:exif={(e) => getCoordinatesFromPhoto(e.detail)}
                bind:photos={$form.photos}
                bind:photoFiles={$form._photos}
                showThumbnailControls={false}
                showExifControls={true}
            ></PhotoPicker>
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
