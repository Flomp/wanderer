import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import tailwind from '@astrojs/tailwind';

import svelte from "@astrojs/svelte";

// https://astro.build/config
export default defineConfig({
  integrations: [starlight({
    title: 'Docs with Tailwind',
    logo: {
      light: '/src/assets/logo_dark_text.svg',
      dark: '/src/assets/logo_light_text.svg',
      replacesTitle: true
    },
    social: {
      github: 'https://github.com/flomp/wanderer'
    },
    components: {
      Footer: './src/components/Footer.astro',
    },
    sidebar: [{
      label: 'Getting Started',
      items: [
        // Each item here is one entry in the navigation menu.
        {
          label: 'Installation',
          link: '/getting-started/installation/'
        }]
    }, {
      label: 'Reference',
      autogenerate: {
        directory: 'reference'
      }
    }],
    customCss: ['./src/custom.css', './src/tailwind.css', '@fontsource/ibm-plex-sans/400.css', '@fontsource/ibm-plex-sans/600.css', '@fontsource/ibm-plex-mono/400.css', '@fontsource/ibm-plex-mono/600.css']
  }), tailwind({
    applyBaseStyles: false
  }), svelte()]
});