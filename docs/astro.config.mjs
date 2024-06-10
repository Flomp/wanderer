import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import tailwind from '@astrojs/tailwind';

import node from "@astrojs/node";

// https://astro.build/config
export default defineConfig({
  integrations: [starlight({
    title: 'wanderer Documentation',
    logo: {
      light: '/src/assets/logo_dark_text.svg',
      dark: '/src/assets/logo_light_text.svg',
      replacesTitle: true
    },
    social: {
      github: 'https://github.com/flomp/wanderer'
    },
    components: {
      Footer: './src/components/footer.astro'
    },
    sidebar: [{
      label: 'Getting Started',
      items: [{
        label: 'Installation',
        link: '/getting-started/installation/'
      }, {
        label: 'Configuration',
        link: '/getting-started/configuration/'
      }, {
        label: 'Local development',
        link: '/getting-started/local-development/'
      }, {
        label: 'Changelog',
        link: '/getting-started/changelog/'
      }]
    }, {
      label: 'Guides',
      items: [{
        label: 'Authentication',
        link: '/guides/authentication/'
      }, {
        label: 'Create a trail',
        link: '/guides/create-a-trail/'
      }, {
        label: 'Share trails',
        link: '/guides/share-trails/'
      }, {
        label: 'Lists',
        link: '/guides/lists/'
      }, {
        label: 'Import/Export',
        link: '/guides/import-export/'
      }, {
        label: 'API',
        link: '/guides/api/'
      }]
    }, {
      label: 'API Reference',
      autogenerate: {
        directory: 'api-reference'
      }
    }],
    customCss: ['./src/custom.css', './src/tailwind.css', '@fontsource/ibm-plex-sans/400.css', '@fontsource/ibm-plex-sans/600.css', '@fontsource/ibm-plex-mono/400.css', '@fontsource/ibm-plex-mono/600.css']
  }), tailwind({
    applyBaseStyles: false
  })],
  output: "server",
  adapter: node({
    mode: "standalone"
  })
});