<script lang="ts">
    import { createEventDispatcher } from "svelte";

    import { WaypointCreateSchema } from "$lib/models/api/waypoint_schema";
    import { convertDMSToDD } from "$lib/models/gpx/utils";
    import { show_toast } from "$lib/stores/toast_store";
    import { waypoint } from "$lib/stores/waypoint_store";
    import { cloneDeep } from "$lib/util/deep_util";
    import { icons } from "$lib/util/icon_util";
    import EXIF from "$lib/vendor/exif-js/exif";
    import { validator } from "@felte/validator-zod";
    import { createForm } from "felte";
    import { _ } from "svelte-i18n";
    import { z } from "zod";
    import Combobox from "../base/combobox.svelte";
    import Modal from "../base/modal.svelte";
    import TextField from "../base/text_field.svelte";
    import Textarea from "../base/textarea.svelte";
    import PhotoPicker from "../trail/photo_picker.svelte";

    export let openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;

    const dispatch = createEventDispatcher();

    const ClientWaypointCreateSchema = WaypointCreateSchema.extend({
        _photos: z.array(z.instanceof(File)).optional(),
    });

    const { form, errors, data, setFields } = createForm<
        z.infer<typeof ClientWaypointCreateSchema>
    >({
        initialValues: $waypoint,
        extend: validator({ schema: ClientWaypointCreateSchema }),
        onSubmit: async (form) => {
            dispatch("save", form);

            closeModal!();
        },
        transform: (values: unknown) => {
            const v = values as any;
            return {
                ...v,
                lat: parseFloat(v.lat),
                lon: parseFloat(v.lon),
            };
        },
    });

    $: setFields(cloneDeep($waypoint));

    $: filteredIcons =
        ($data.icon?.length ?? 0) > 2
            ? icons
                  .filter((i) => i.replaceAll("-", " ").includes($data.icon?.toLowerCase() ?? ""))
                  .map((i) => ({ text: i.replaceAll("-", " "), value: i, icon: i }))
            : [];

    function getCoordinatesFromPhoto(src: string) {
        EXIF.getData({ src: src }, function (p) {
            const lat = EXIF.getTag(p, "GPSLatitude");
            const latDir = EXIF.getTag(p, "GPSLatitudeRef");
            const lon = EXIF.getTag(p, "GPSLongitude");
            const lonDir = EXIF.getTag(p, "GPSLongitudeRef");

            if (lat && lon) {
                setFields("lat", convertDMSToDD(lat, latDir));
                setFields("lon", convertDMSToDD(lon, lonDir));
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
    title={$data.id ? $_("edit-waypoint") : $_("add-waypoint")}
    let:openModal
    bind:openModal
    bind:closeModal
>
    <slot {openModal} />
    <form
        id="waypoint-form"
        slot="content"
        class="modal-content space-y-4"
        use:form
    >
        <div class="flex gap-4">
            <div class="basis-2/3">
                <TextField name="name" label={$_("name")} error={$errors.name}
                ></TextField>
            </div>

            <Combobox
                name="icon"
                icon={$data.icon}
                bind:value={$data.icon}
                items={filteredIcons}
                label={$_("icon")}
            ></Combobox>
        </div>

        <Textarea
            name="description"
            label={$_("description")}
            error={$errors.description}
        ></Textarea>
        <div class="flex gap-4">
            <TextField name="lat" label={$_("latitude")} error={$errors.lat}
            ></TextField>
            <TextField name="lon" label={$_("longitude")} error={$errors.lat}
            ></TextField>
        </div>
        <div>
            <label for="waypoint-photo-input" class="text-sm font-medium pb-1">
                {$_("photos")}
            </label>
            <PhotoPicker
                id="waypoint-photo-input"
                parent={$data}
                on:exif={(e) => getCoordinatesFromPhoto(e.detail)}
                bind:photos={$data.photos}
                bind:photoFiles={$data._photos}
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
