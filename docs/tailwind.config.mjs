import colors from 'tailwindcss/colors';
import starlightPlugin from '@astrojs/starlight-tailwind';

const accent = { 200: '#b0c8fd', 600: '#2a56f1', 900: '#152b6d', 950: '#112149' };

/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
	theme: {
		extend: {
			colors: {
				primary: "#242734",
				// Your preferred accent color. Indigo is closest to Starlight’s defaults.
				accent: accent,
				// Your preferred gray scale. Zinc is closest to Starlight’s defaults.
				gray: colors.gray,
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
