<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import EmailModal from "$lib/components/settings/email_modal.svelte";
    import PasswordModal from "$lib/components/settings/password_modal.svelte";
    import { show_toast } from "$lib/stores/toast_store";
    import {
        currentUser,
        logout,
        users_delete,
        users_update,
    } from "$lib/stores/user_store";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    const settings = $page.data.settings;

    let selectedLanguage = "en";
    let selectedMapFocus = "trails";

    let citySearchQuery: string = "";

    let openConfirmModal: () => void;
    let openEmailModal: () => void;
    let openPasswordModal: () => void;

    onMount(() => {
        citySearchQuery = settings?.location?.name ?? "";
        selectedLanguage = settings?.language || "en";
        selectedMapFocus = settings?.mapFocus ?? "trails";
    });

    async function deleteAccount() {
        await users_delete($currentUser!);
        logout();
        goto("/");
    }

    async function updateEmail(email: string) {
        try {
            await users_update({ ...$currentUser!, email: email });
            show_toast({
                text: $_("email-updated"),
                icon: "check",
                type: "success",
            });
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
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
{#if $currentUser}
    <h2 class="text-2xl font-semibold">{$_("my-account")}</h2>
    <hr class="mt-4 mb-6 border-input-border" />
    <div class="space-y-6">
        <h4 class="text-xl font-medium">{$_("login-details")}</h4>
        <button class="btn-secondary block" on:click={openEmailModal}
            >{$_("change-email")}</button
        >
        <button class="btn-secondary" on:click={openPasswordModal}
            >{$_("change-password")}</button
        >
        <div class="space-y-4">
            <h4 class="text-xl text-red-400 font-medium">
                {$_("danger-zone")}
            </h4>
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
{/if}
<ConfirmModal
    text={$_("account-delete-confirm")}
    bind:openModal={openConfirmModal}
    on:confirm={deleteAccount}
></ConfirmModal>
