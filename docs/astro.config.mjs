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
      sidebar: [
        {
          label: 'Welcome to wanderer',
          link: '/welcome'
        },
        {
          label: 'Using wanderer',
          items: [{
            label: 'Authentication',
            link: '/use/authentication/'
          }, {
            label: 'Create a trail',
            link: '/use/create-a-trail/'
          },
          {
            label: 'Summit logs',
            link: '/use/summit-logs/'
          },
          {
            label: 'Interact with the community',
            link: '/use/community-interaction/'
          },
          {
            label: 'Share trails',
            link: '/use/share-trails/'
          }, {
            label: 'Lists',
            link: '/use/lists/'
          },
          {
            label: 'Statistics',
            link: '/use/statistics/'
          },
          {
            label: 'Customize the map',
            link: '/use/customize-map/'
          },
          {
            label: 'Import/Export',
            link: '/use/import-export/'
          },
          {
            label: 'Integrations',
            link: '/use/integrations/'
          },
          ]
        },
        {
          label: 'Running wanderer',
          items: [
            {
              label: 'Installation',
              link: '/run/installation/'
            },
            {
              label: 'Environment configuration',
              link: '/run/environment-configuration/'
            },
            {
              label: 'Backend configuration',
              link: '/run/backend-configuration/'
            },
            {
              label: 'Custom categories',
              link: '/run/custom-categories/'
            },
            {
              label: 'Backing up your server',
              link: '/run/backup-server/'
            },
            {
              label: 'Changelog',
              link: '/run/changelog/'
            }]
        }, {
          label: 'Develop wanderer',
          items: [
            {
              label: 'Local development',
              link: '/develop/local-development/'
            },
            {
              label: 'API',
              link: '/develop/api/'
            },
            {
              label: 'Federation',
              link: '/develop/federation/'
            },
          ]
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