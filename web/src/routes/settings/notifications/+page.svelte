<script lang="ts">
    import { page } from "$app/state";
    import Toggle from "$lib/components/base/toggle.svelte";
    import { NotificationType } from "$lib/models/notification";
    import type { Settings } from "$lib/models/settings";
    import { settings_update } from "$lib/stores/settings_store";
    import { _ } from "svelte-i18n";

    const notifications = $state((page.data.settings as Settings)?.notifications ?? {
        list_create: {
            web: true,
            email: true,
        },
        list_share: {
            web: true,
            email: true,
        },
        trail_create: {
            web: true,
            email: true,
        },
        trail_share: {
            web: true,
            email: true,
        },
        new_follower: {
            web: true,
            email: true,
        },
        trail_comment: {
            web: true,
            email: true,
        },
    });

    const notificationItems: { text: string; key: NotificationType }[] = [
        {
            text: $_("settings-notification-trail-comment"),
            key: NotificationType.trailComment,
        },
        {
            text: $_("settings-notification-new-follower"),
            key: NotificationType.newFollower,
        },
        {
            text: $_("settings-notification-trail-create"),
            key: NotificationType.trailCreate,
        },
        {
            text: $_("settings-notification-trail-share"),
            key: NotificationType.trailShare,
        },
        {
            text: $_("settings-notification-list-create"),
            key: NotificationType.listCreate,
        },
        {
            text: $_("settings-notification-list-share"),
            key: NotificationType.listShare,
        },
    ];

    async function updateNotificationSettings() {
        await settings_update({
            id: page.data.settings!.id,
            notifications,
        });
    }
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
<h2 class="text-2xl font-semibold">{$_("notifications")}</h2>
<hr class="mt-4 mb-6 border-input-border" />

<div
    class="grid gap-4"
    style="grid-template-columns: 1fr min-content min-content;"
>
    <div></div>
    <span class="text-sm font-medium">Web</span>
    <span class="text-sm font-medium">Email</span>
    {#each notificationItems as item}
        <p>{item.text}</p>
        <div>
            <Toggle
                onchange={updateNotificationSettings}
                bind:value={notifications[item.key].web}
            ></Toggle>
        </div>
        <div>
            <Toggle
                onchange={updateNotificationSettings}
                bind:value={notifications[item.key].email}
            ></Toggle>
        </div>
    {/each}
</div>
