// +layout.ts
import '$lib/i18n';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	return { settings: locals.settings }
}