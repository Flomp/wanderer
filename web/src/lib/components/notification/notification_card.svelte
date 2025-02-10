<script lang="ts">
    import {
        NotificationType,
        type Notification,
    } from "$lib/models/notification";
    import { getFileURL } from "$lib/util/file_util";
    import { formatTimeSince } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";

    interface Props {
        notification: Notification;
        onclick?: (data: {notification: Notification, link: string | null}) => void
    }

    let { notification, onclick }: Props = $props();

    const avatarSrc = notification.expand?.author.avatar
        ? getFileURL(
              notification.expand.author,
              notification.expand.author.avatar,
          )
        : `https://api.dicebear.com/7.x/initials/svg?seed=${notification.expand?.author.username ?? ""}&backgroundType=gradientLinear`;

    const timeSince = formatTimeSince(new Date(notification.created ?? ""));


    function getTitle(n: Notification) {
        switch (n.type) {
            case NotificationType.listCreate:
                return $_("notification-list-create", {
                    values: { user: n.expand.author.username },
                });
            case NotificationType.listShare:
                return $_("notification-list-share", {
                    values: { user: n.expand.author.username },
                });
            case NotificationType.newFollower:
                return $_("notification-new-follower");
            case NotificationType.trailComment:
                return $_("notification-trail-comment", {
                    values: {
                        user: n.expand.author.username,
                        trail: n.metadata?.trail,
                    },
                });
            case NotificationType.trailCreate:
                return $_("notification-trail-create", {
                    values: { user: n.expand.author.username },
                });
            case NotificationType.trailShare:
                return $_("notification-trail-share", {
                    values: { user: n.expand.author.username },
                });
        }
    }

    function getDescription(n: Notification) {
        switch (n.type) {
            case NotificationType.listCreate:
                return n.metadata?.list ?? "";
            case NotificationType.listShare:
                return n.metadata?.list ?? "";
            case NotificationType.newFollower:
                return n.expand.author.username ?? "";
            case NotificationType.trailComment:
                return n.metadata?.comment ?? "";
            case NotificationType.trailCreate:
                return n.metadata?.trail ?? "";
            case NotificationType.trailShare:
                return n.metadata?.trail ?? "";
        }
    }

    function getLink(n: Notification) {
        switch (n.type) {
            case NotificationType.listCreate:
                return `/lists?list=${n.metadata?.id}`;
            case NotificationType.listShare:
                return `/lists?list=${n.metadata?.id}`;
            case NotificationType.newFollower:
                return n.expand.author.private === true
                    ? null
                    : `/profile/${n.author}`;
            case NotificationType.trailComment:
                return `/trail/view/${n.metadata?.id}?t=4`;
            case NotificationType.trailCreate:
                return `/trail/view/${n.metadata?.id}`;
            case NotificationType.trailShare:
                return `/trail/view/${n.metadata?.id}`;
        }
    }

    function handleItemClick() {
        onclick?.({ notification, link });
    }
    let title = $derived(getTitle(notification));
    let description = $derived(getDescription(notification));
    let link = $derived(getLink(notification));
</script>

<li
    class="flex items-center gap-x-3 px-3 py-2 hover:bg-menu-item-background-hover relative cursor-pointer"
    role="presentation"
    onclick={handleItemClick}
>
    <img class="rounded-full w-8 aspect-square" src={avatarSrc} alt="avatar" />
    <div>
        <p
            class="text-sm {notification.seen
                ? 'font-medium'
                : 'font-semibold'} mr-3"
        >
            {title}:
        </p>
        <p class="text-sm line-clamp-1">{description}</p>
        <p class="text-xs text-gray-500">
            {$_(`n-${timeSince.unit}-ago`, {
                values: { n: timeSince.value },
            })}
        </p>
    </div>

    {#if !notification.seen}
        <div
            class="bg-content w-[6px] aspect-square rounded-full absolute top-3 right-3"
        ></div>
    {/if}
</li>
