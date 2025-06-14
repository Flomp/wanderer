import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

import node from "@astrojs/node";

import starlightOpenAPI, { openAPISidebarGroups } from 'starlight-openapi'
import tailwindcss from "@tailwindcss/vite";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: 'wanderer Documentation',
      logo: {
        light: '/src/assets/logo_text_dark.svg',
        dark: '/src/assets/logo_text_light.svg',
        replacesTitle: true
      },
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/flomp/wanderer' },
      ],
      components: {
        Footer: './src/components/footer.astro'
      },
      plugins: [
        starlightOpenAPI([
          {
            base: 'api-reference',
            label: 'API Reference',
            schema: 'wanderer.openapi.yaml',
          },
        ]),
      ],
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
        },
        {
          label: 'Statistics',
          link: '/guides/statistics/'
        },
        {
          label: 'Custom categories',
          link: '/guides/custom-categories/'
        },
        {
          label: 'Customize the map',
          link: '/guides/customize-map/'
        },
        {
          label: 'Import/Export',
          link: '/guides/import-export/'
        },
        {
          label: 'Integrations',
          link: '/guides/integrations/'
        },
        {
          label: 'API',
          link: '/guides/api/'
        }]
      },
      ...openAPISidebarGroups,],
      customCss: ['./src/custom.css', './src/tailwind.css', '@fontsource/ibm-plex-sans/400.css', '@fontsource/ibm-plex-sans/600.css', '@fontsource/ibm-plex-mono/400.css', '@fontsource/ibm-plex-mono/600.css']
    })],
  output: "server",
  vite: { plugins: [tailwindcss()] },
  adapter: node({
    mode: "standalone"
  })
});