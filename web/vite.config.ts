import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	build: {
        sourcemap: true, // Active les sourcemaps pour les builds
    },
	test: {
		include: ['src/**/*.{test,spec}.{js,ts}']
	},
	ssr: {
		noExternal: ['three']
	},
	server: {
    watch: {
      usePolling: true,
      interval: 1000, // Vérification toutes les 1000 ms
    },
    host: '0.0.0.0', // Permet l'accès depuis l'extérieur du conteneur
  }
});
