<script lang="ts">
    import { listSchema, type List } from "$lib/models/list";
    import { createForm } from "$lib/vendor/svelte-form-lib/index";
    import { _ } from "svelte-i18n";

    import { page } from "$app/stores";
    import Button from "$lib/components/base/button.svelte";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Textarea from "$lib/components/base/textarea.svelte";
    import MapWithElevationMultiple from "$lib/components/trail/map_with_elevation_multiple.svelte";
    import type { Trail } from "$lib/models/trail.js";
    import { lists_create, lists_update } from "$lib/stores/list_store.js";
    import { show_toast } from "$lib/stores/toast_store.js";
    import { trails_show } from "$lib/stores/trail_store";
    import { getFileURL } from "$lib/util/file_util.js";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import {
        trail_share_create,
        trail_share_index,
    } from "$lib/stores/trail_share_store.js";
    import { TrailShare } from "$lib/models/trail_share.js";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";

    export let data;

    let previewURL = data.previewUrl ?? "";
    let searchDropdownItems: SearchItem[] = [];

    let activeTrailIndex: number | null = null;

    let map: MapWithElevationMultiple;

    let loading: boolean = false;

    let newShares: TrailShare[] = [];

    let openConfirmModal: () => void;

    const { form, errors, handleChange, handleSubmit } = createForm<List>({
        initialValues: data.list!,
        validationSchema: listSchema,
        onSubmit: async (submittedList) => {
            const avatarFile = (
                document.getElementById("avatar") as HTMLInputElement
            ).files![0];
            loading = true;
            try {
                if ($form.id) {
                    await lists_update($form, avatarFile);
                    await findNewTrailShares();
                    if (!newShares.length) {
                        show_toast({
                            type: "success",
                            icon: "check",
                            text: $_("list-saved-successfully"),
                        });
                    }
                } else {
                    await lists_create($form, avatarFile);
                    show_toast({
                        type: "success",
                        icon: "check",
                        text: $_("list-saved-successfully"),
                    });
                }
            } catch (e) {
                show_toast({
                    type: "error",
                    icon: "close",
                    text: $_("error-saving-list"),
                });
            } finally {
                loading = false;
            }
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

    async function search(q: string) {
        const r = await fetch("/api/v1/search/multi", {
            method: "POST",
            body: JSON.stringify({
                queries: [
                    {
                        indexUid: "trails",
                        q: q,
                        limit: 3,
                    },
                ],
            }),
        });

        const response = await r.json();

        searchDropdownItems = response.results[0].hits
            .filter((h: List) => !$form.trails?.includes(h.id!))
            .map((t: Record<string, any>) => ({
                text: t.name,
                description: `${t.location ?? "-"}`,
                value: t.id,
                icon: "route",
            }));
    }

    async function handleSearchClick(item: SearchItem) {
        const trail = await trails_show(item.value, true);
        $form.trails?.push(trail.id!);
        $form.expand = {
            trails: [...($form.expand?.trails ?? []), trail],
            list_share_via_list: $form.expand?.list_share_via_list,
        };
    }

    function deleteTrail(trail: Trail) {
        $form.trails = $form.trails?.filter((id) => id !== trail.id);
        $form.expand!.trails = $form.expand!.trails!.filter(
            (t) => t.id !== trail.id,
        );
    }

    async function findNewTrailShares() {
        const usersWithAccess: string[] = [
            $form.author!,
            ...($form.expand?.list_share_via_list ?? []).map((s) => s.user),
        ];
        for (const userId of usersWithAccess) {
            const existingTrailShares = await trail_share_index({
                user: userId,
            });
            for (const trail of $form.expand?.trails ?? []) {
                if (trail.author == userId) {
                    continue;
                }
                const trailShare = existingTrailShares.find(
                    (s) => s.trail == trail.id,
                );
                if (!trailShare) {
                    newShares.push(new TrailShare(userId, trail.id!, "view"));
                }
            }
        }

        if (newShares.length) {
            openConfirmModal();
        }
    }

    async function updateTrailShares() {
        for (const newShare of newShares) {
            await trail_share_create(newShare);
        }
        newShares = [];
        show_toast({
            type: "success",
            icon: "check",
            text: $_("list-saved-successfully"),
        });
    }

    function moveTrail(trail: Trail, index: number, direction: 1 | -1) {
        (document?.activeElement as HTMLElement).blur();

        if (!$form.trails || !$form.expand || !$form.expand.trails) {
            return;
        }
        const previousTrail = $form.expand?.trails?.at(index + direction);
        const previousTrailId = $form.trails?.at(index + direction);

        if (!previousTrail || !previousTrailId) {
            return;
        }

        $form.expand!.trails![index] = previousTrail;
        $form.expand!.trails![index + direction] = trail;

        $form.trails[index] = previousTrailId;
        $form.trails[index + direction] = trail.id!;
    }
</script>

<svelte:head>
    <title
        >{$form.id ? `${$form.name} | ${$_("edit")}` : $_("new-list")} | wanderer</title
    >
</svelte:head>

<main class="grid grid-cols-1 md:grid-cols-[440px_1fr]">
    <form
        id="list-form"
        class="flex flex-col gap-4 px-8 order-1 md:order-none mt-8 md:mt-0 overflow-y-scroll"
        on:submit={handleSubmit}
    >
        <h2 class="text-2xl font-semibold">
            {$page.params.id === "new" ? $_("new-list") : $_("edit-list")}
        </h2>
        <label for="avatar" class="text-sm font-medium block">
            {$_("avatar")}
        </label>
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
                    class="w-32 aspect-square rounded-full object-cover border border-gray-100"
                    alt="avatar"
                    src={previewURL}
                />
            {:else}
                <div
                    class="flex items-center justify-center w-32 aspect-square rounded-full object-cover border border-gray-200"
                >
                    <i class="fa fa-table-list text-5xl"></i>
                </div>
            {/if}
            <button
                class="btn-secondary"
                type="button"
                on:click={openAvatarBrowser}>{$_("change")}...</button
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
        <h3 class="text-xl font-semibold">
            {$_("trail", { values: { n: 2 } })}
        </h3>
        <Search
            on:update={(e) => search(e.detail)}
            on:click={(e) => handleSearchClick(e.detail)}
            placeholder="{$_('search-trails')}..."
            items={searchDropdownItems}
        ></Search>
        {#if $form.expand?.trails?.length}
            {#each $form.expand?.trails ?? [] as trail, i}
                <div
                    class="flex gap-4 p-4 rounded-xl border border-input-border cursor-pointer hover:bg-secondary-hover transition-colors items-center"
                    class:bg-secondary-hover={i == activeTrailIndex}
                    role="presentation"
                    on:mouseenter={() => map.highlightTrail(trail.id ?? "")}
                    on:mouseleave={() => map.unHighlightTrail(trail.id ?? "")}
                    on:click={() => map.selectTrail(trail.id ?? "")}
                >
                    <div class="shrink-0">
                        <img
                            class="h-12 w-12 object-cover rounded-xl"
                            src={trail.photos.length
                                ? getFileURL(
                                      trail,
                                      trail.photos[trail.thumbnail],
                                  )
                                : "/imgs/default_thumbnail.webp"}
                            alt=""
                        />
                    </div>
                    <div class="basis-full">
                        <h4 class="font-semibold text-lg">
                            {trail.name}
                        </h4>
                        <span class="text-sm"
                            ><i class="fa fa-gauge mr-2"></i>{$_(
                                trail.difficulty ?? "?",
                            )}</span
                        >
                        <div class="grid grid-cols-2 mt-1 gap-x-4 gap-y-2 text-sm text-gray-500">
                            <span
                                ><i class="fa fa-left-right mr-2"
                                ></i>{formatDistance(trail.distance)}</span
                            >
                            <span
                                ><i class="fa fa-clock mr-2"
                                ></i>{formatTimeHHMM(trail.duration)}</span
                            >
                            <span
                                ><i class="fa fa-arrow-trend-up mr-2"
                                ></i>{formatElevation(
                                    trail.elevation_gain,
                                )}</span
                            >
                            <span
                                ><i class="fa fa-arrow-trend-down mr-2"
                                ></i>{formatElevation(
                                    trail.elevation_loss,
                                )}</span
                            >
                        </div>
                    </div>
                    <div class="basis-0">
                        {#if i > 0}
                            <button
                                on:click|stopPropagation={() =>
                                    moveTrail(trail, i, -1)}
                                type="button"
                                class="btn-icon"
                            >
                                <i class="fa fa-chevron-up"></i>
                            </button>
                        {/if}
                        {#if i < $form.expand?.trails?.length - 1}
                            <button
                                on:click|stopPropagation={() =>
                                    moveTrail(trail, i, 1)}
                                type="button"
                                class="btn-icon"
                            >
                                <i class="fa fa-chevron-down"></i>
                            </button>
                        {/if}
                        <button
                            type="button"
                            class="btn-icon text-red-500"
                            on:click={() => deleteTrail(trail)}
                            ><i class="fa fa-trash"></i></button
                        >
                    </div>
                </div>
            {/each}
        {:else}
            <span class="text-center text-sm text-gray-500 my-8"
                >No routes added</span
            >
        {/if}
        <Button
            primary={true}
            large={true}
            type="submit"
            extraClasses="mb-2"
            {loading}>{$_("save-list")}</Button
        >
    </form>
    <div id="trail-map" class="max-h-full">
        <MapWithElevationMultiple
            trails={$form.expand?.trails ?? []}
            options={{ flyToBounds: true }}
            bind:activeTrailIndex
            bind:this={map}
        ></MapWithElevationMultiple>
    </div>
</main>

<ConfirmModal
    text={$_("list-share-warning-update")}
    title={$_("confirm-share")}
    action="confirm"
    bind:openModal={openConfirmModal}
    on:confirm={updateTrailShares}
></ConfirmModal>

<style>
    @media only screen and (min-width: 768px) {
        #trail-map,
        #list-form {
            height: calc(100vh - 124px);
        }
    }
</style>
