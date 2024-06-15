// +layout.ts
import '$lib/i18n';
import type { LayoutServerLoad } from './$types';
import { env } from "$env/dynamic/private";

export const load: LayoutServerLoad = async ({ locals, url }) => {
	const warnings = []
	if (env.ORIGIN != url.origin) {
		warnings.push(`You are accessing wanderer from <span class="font-mono bg-gray-100">${url.origin}</span> but your ORIGIN environment variable is set to <span class="font-mono bg-gray-100">${env.ORIGIN}</span>. This may cause errors.`)
	}
	return { settings: locals.settings, warnings: warnings }
}