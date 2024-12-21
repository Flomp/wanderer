// +layout.ts
import '$lib/i18n';
import type { LayoutServerLoad } from './$types';
import { env } from "$env/dynamic/private";
import type { Settings } from '$lib/models/settings';
import { notifications_index } from '$lib/stores/notification_store';

export const load: LayoutServerLoad = async ({ locals, url, fetch }) => {

	let notifications
	if (locals.user) {
		notifications = await notifications_index({ recipient: locals.user.id }, 1, 10, fetch);
	}
	return { settings: locals.settings as Settings, notifications, origin: env.ORIGIN }
}