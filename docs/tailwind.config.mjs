import colors from 'tailwindcss/colors';
import starlightPlugin from '@astrojs/starlight-tailwind';

/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
	theme: {
		extend: {
			colors: {
				// Your preferred accent color. Indigo is closest to Starlight’s defaults.
				accent: colors.blue,
				// Your preferred gray scale. Zinc is closest to Starlight’s defaults.
				gray: "#242734",
			},
			fontFamily: {
				// Deine bevorzugte Schriftart. Starlight verwendet standardmäßig eine Systemschriftart.
				sans: ['"IBM Plex Sans"'],
				// Deine bevorzugte Code-Schriftart. Starlight verwendet standardmäßig die Systemschriftart Monospace.
				mono: ['"IBM Plex Mono"'],
			},
		},
	},
	plugins: [starlightPlugin()],
};
