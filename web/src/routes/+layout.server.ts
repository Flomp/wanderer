// +layout.ts
import '$lib/i18n';
import type { LayoutServerLoad } from './$types';
import { env } from "$env/dynamic/private";
import type { Settings } from '$lib/models/settings';

export const load: LayoutServerLoad = async ({ locals, url }) => {	
	return { settings: locals.settings as Settings, origin: env.ORIGIN }
}