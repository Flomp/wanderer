<script lang="ts">
    import { page } from "$app/state";
    import Button from "$lib/components/base/button.svelte";
    import type { SelectItem } from "$lib/components/base/select.svelte";
    import Select from "$lib/components/base/select.svelte";
    import Textarea from "$lib/components/base/textarea.svelte";
    import type { Category } from "$lib/models/category.js";
    import { settings_update } from "$lib/stores/settings_store";
    import { show_toast } from "$lib/stores/toast_store";
    import { currentUser, users_update } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";

    let { data = $bindable() } = $props();

    let selectedCategory = $state(
        page.data.settings?.category || data.categories[0].id,
    );

    let bio = $state(data.settings?.bio ?? "");

    const categoryItems: SelectItem[] = data.categories.map((c: Category) => ({
        text: $_(c.name),
        value: c.id,
    }));

    function openFileBrowser() {
        document.getElementById("avatarInput")!.click();
    }

    async function handleAvatarSelection() {
        if (!$currentUser) {
            return;
        }
        const files = (
            document.getElementById("avatarInput") as HTMLInputElement
        ).files;

        if (!files || files.length == 0) {
            return;
        }

        await users_update($currentUser!, files[0]);
    }

    async function handleBioSave() {
        if (!data.settings) {
            return;
        }
        try {
            data.settings.bio = bio;
            await settings_update(data.settings);
        } catch (e) {
            show_toast({
                type: "error",
                icon: "close",
                text: "Error saving bio",
            });
            console.error(e);
        }
    }

    async function handleCategorySelection(value: string) {
        await settings_update({
            id: page.data.settings!.id,
            category: value,
        });
    }
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
{#if $currentUser}
    <h2 class="text-2xl font-semibold">{$_("profile")}</h2>
    <hr class="mt-4 mb-6 border-input-border" />
    <div class="space-y-6">
        <div class="flex gap-6 items-center">
            <div
                class="rounded-full w-24 aspect-square overflow-hidden relative group"
            >
                <img
                    class="object-cover h-full"
                    src={getFileURL($currentUser, $currentUser.avatar) ||
                        `https://api.dicebear.com/7.x/initials/svg?seed=${$currentUser.username}&backgroundType=gradientLinear`}
                    alt="avatar"
                />
                <button
                    aria-label="Open file browser"
                    class="absolute top-0 w-24 aspect-square opacity-0 group-hover:opacity-100 flex justify-center items-center bg-black/50 focus:bg-black/60 text-white cursor-pointer transition-opacity"
                    onclick={openFileBrowser}
                >
                    <i class="fa fa-pen"></i>
                </button>
                <input
                    type="file"
                    name="avatar"
                    id="avatarInput"
                    accept="image/*"
                    style="display: none;"
                    onchange={handleAvatarSelection}
                />
            </div>
            <div>
                <h4 class="text-xl font-semibold">{$currentUser.username}</h4>
                <h5 class="font-medium">{$currentUser.email}</h5>
            </div>
        </div>
        <div>
            <h4 class="text-xl font-medium">Bio</h4>
            <Textarea bind:value={bio} rows={5}></Textarea>
            <div class="mt-3">
                <Button
                    onclick={() => handleBioSave()}
                    primary
                    disabled={data.settings.bio === bio}>{$_("save")}</Button
                >
            </div>
        </div>

        <div>
            <h4 class="text-xl font-medium mb-2">{$_("favourite-sport")}</h4>
            <Select
                items={categoryItems}
                bind:value={selectedCategory}
                onchange={handleCategorySelection}
            ></Select>
        </div>
    </div>
{/if}
