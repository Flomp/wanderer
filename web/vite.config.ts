import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
import fs from 'fs';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	test: { include: ['src/**/*.{test,spec}.{js,ts}'] },
	ssr: { noExternal: ['three'] },
	// server: {
	// 	https: {
	// 		key: fs.readFileSync('.svelte-kit/key.pem'),
	// 		cert: fs.readFileSync('.svelte-kit/cert.pem')
	// 	},
	// 	host: true, // true
	// 	port: 443 // 443
	// }
});
