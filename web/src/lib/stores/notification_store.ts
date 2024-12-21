import type { Notification } from "$lib/models/notification";
import { ClientResponseError, type ListResult } from "pocketbase";

let notifications: Notification[] = [];

export async function notifications_index(data: { recipient: string, seen?: boolean }, page: number = 1, perPage: number = 10, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/notification?' + new URLSearchParams({
        filter: `created>=@month&&recipient='${data.recipient}'` + (data.seen !== undefined ? `&&seen=${data.seen}` : ''),
        sort: '+seen,-created',
        page: page.toString(),
        "per-page": perPage.toString()
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const fetchedNotifications: ListResult<Notification> = await r.json();

    const result = page > 1 ? [...notifications, ...fetchedNotifications.items] : fetchedNotifications.items

    notifications = result;

    return { ...fetchedNotifications, items: result };
}

export async function notifications_mark_as_seen(notification: Notification) {
    let r = await fetch('/api/v1/notification/' + notification.id, {
        method: 'POST',
        body: JSON.stringify(notification),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}