// +layout.ts
import '$lib/i18n';
import type { LayoutServerLoad } from './$types';
import { env } from "$env/dynamic/private";

export const load: LayoutServerLoad = async ({ locals, url }) => {	
	return { settings: locals.settings, origin: env.ORIGIN }
}