<script lang="ts">
    import { page } from "$app/stores";
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

    export let data;

    let selectedCategory =
        $page.data.settings?.category || data.categories[0].id;

    let bio = $currentUser?.bio ?? "";

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
        if (!$currentUser) {
            return;
        }
        try {
            $currentUser.bio = bio;
            await users_update($currentUser);
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
            id: $page.data.settings!.id,
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
                    class="absolute top-0 w-24 aspect-square opacity-0 group-hover:opacity-100 flex justify-center items-center bg-black/50 focus:bg-black/60 text-white cursor-pointer transition-opacity"
                    on:click={openFileBrowser}
                >
                    <i class="fa fa-pen"></i>
                </button>
                <input
                    type="file"
                    name="avatar"
                    id="avatarInput"
                    accept="image/*"
                    style="display: none;"
                    on:change={handleAvatarSelection}
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
                    on:click={() => handleBioSave()}
                    primary
                    disabled={$currentUser.bio === bio}>{$_("save")}</Button
                >
            </div>
        </div>

        <div>
            <h4 class="text-xl font-medium mb-2">{$_("favourite-sport")}</h4>
            <Select
                items={categoryItems}
                bind:value={selectedCategory}
                on:change={(e) => handleCategorySelection(e.detail)}
            ></Select>
        </div>
    </div>
{/if}
