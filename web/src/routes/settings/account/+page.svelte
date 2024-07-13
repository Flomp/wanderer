<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { type RadioItem } from "$lib/components/base/radio_group.svelte";
    import { type SearchItem } from "$lib/components/base/search.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import EmailModal from "$lib/components/profile/email_modal.svelte";
    import PasswordModal from "$lib/components/profile/password_modal.svelte";
    import Settings from "$lib/components/profile/settings.svelte";
    import { settings_update } from "$lib/stores/settings_store";
    import { show_toast } from "$lib/stores/toast_store";
    import {
        currentUser,
        logout,
        users_delete,
        users_update,
    } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { getFileURL } from "$lib/util/file_util";
    import { onMount } from "svelte";
    import { _, locale } from "svelte-i18n";

    const settings = $page.data.settings;

    const languages: SelectItem[] = [
        { text: $_("chinese"), value: "zh" },
        { text: $_("german"), value: "de" },
        { text: $_("english"), value: "en" },
        { text: $_("french"), value: "fr" },
        { text: $_("hungarian"), value: "hu" },
        { text: $_("italian"), value: "it" },
        { text: $_("dutch"), value: "nl" },
        { text: $_("polish"), value: "pl" },
        { text: $_("portuguese"), value: "pt" },
    ];

    const mapFocus: SelectItem[] = [
        { text: $_("trail", { values: { n: 2 } }), value: "trails" },
        { text: $_("location"), value: "location" },
    ];

    const units: RadioItem[] = [
        { text: $_("metric"), value: "metric" },
        { text: $_("imperial"), value: "imperial" },
    ];

    let selectedLanguage = "en";
    let selectedMapFocus = "trails";

    let searchDropdownItems: SearchItem[] = [];
    let citySearchQuery: string = "";

    let openConfirmModal: () => void;
    let openEmailModal: () => void;
    let openPasswordModal: () => void;

    onMount(() => {
        citySearchQuery = settings?.location?.name ?? "";
        selectedLanguage = settings?.language || "en";
        selectedMapFocus = settings?.mapFocus ?? "trails";
    });

    async function searchCities(q: string) {
        const r = await fetch("/api/v1/search/cities500", {
            method: "POST",
            body: JSON.stringify({ q: q, options: { limit: 5 } }),
        });
        const result = await r.json();
        searchDropdownItems = result.hits.map((h: Record<string, any>) => ({
            text: h.name,
            description: `${h.division ? `${h.division} | ` : ""}${
                country_codes[h["country code"] as keyof typeof country_codes]
            }`,
            value: h,
            icon: "city",
        }));
    }

    async function handleAvatarSelection() {
        const files = (
            document.getElementById("avatarInput") as HTMLInputElement
        ).files;

        if (!files || files.length == 0) {
            return;
        }

        await users_update($currentUser!, files[0]);
    }

    async function handleLanguageSelection(
        value: "en" | "de" | "fr" | "hu" | "nl" | "pl" | "pt",
    ) {
        locale.set(value);
        await settings_update({
            id: settings!.id,
            language: value,
        });
    }

    async function deleteAccount() {
        await users_delete($currentUser!);
        logout();
        goto("/");
    }

    async function updateEmail(email: string) {
        try {
            await users_update({ ...$currentUser!, email: email });
        } catch (e) {
            show_toast({
                text: "Error updating email",
                icon: "close",
                type: "error",
            });
        }
    }

    async function updatePassword(data: {
        oldPassword: string;
        password: string;
        passwordConfirm: string;
    }) {
        try {
            await users_update({ ...$currentUser!, ...data });
            show_toast({
                text: $_("password-updated"),
                icon: "check",
                type: "success",
            });
        } catch (e) {
            show_toast({
                text: $_("error-updating-password"),
                icon: "close",
                type: "error",
            });
        }
    }

    function openFileBrowser() {
        document.getElementById("avatarInput")!.click();
    }
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
<main>
    {#if $currentUser}
        <Settings selected={0}>
            <div class="space-y-6">
                <h3 class="text-2xl font-semibold">{$_("profile")}</h3>

                <div class="flex gap-6 items-center">
                    <div
                        class="rounded-full w-24 aspect-square overflow-hidden relative group"
                    >
                        <img
                            src={getFileURL(
                                $currentUser,
                                $currentUser.avatar,
                            ) ||
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
                        <h4 class="text-2xl">{$currentUser.username}</h4>
                        <h5 class="font-medium">{$currentUser.email}</h5>
                    </div>
                </div>

                <h3 class="text-2xl font-semibold">{$_("login-details")}</h3>
                <button class="btn-secondary block" on:click={openEmailModal}
                    >{$_("change-email")}</button
                >
                <button class="btn-secondary" on:click={openPasswordModal}
                    >{$_("change-password")}</button
                >

                <h3 class="text-2xl font-semibold">{$_("language")}</h3>
                <div class="block">
                    <Select
                        items={languages}
                        bind:value={selectedLanguage}
                        on:change={(e) => handleLanguageSelection(e.detail)}
                    ></Select>
                </div>
                <hr class="border-input-border" />
                <div class="space-y-4">
                    <h4 class="text-xl text-red-400">{$_("danger-zone")}</h4>
                    <button class="btn-danger" on:click={openConfirmModal}
                        >{$_("delete-account")}</button
                    >
                </div>
            </div>
            <EmailModal
                email={$currentUser.email}
                on:save={(e) => updateEmail(e.detail)}
                bind:openModal={openEmailModal}
            ></EmailModal>
            <PasswordModal
                on:save={(e) => updatePassword(e.detail)}
                bind:openModal={openPasswordModal}
            ></PasswordModal>
        </Settings>
    {/if}
</main>
<ConfirmModal
    text={$_("account-delete-confirm")}
    bind:openModal={openConfirmModal}
    on:confirm={deleteAccount}
></ConfirmModal>
