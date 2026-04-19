// @ts-check
import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';
import AstroPWA from '@vite-pwa/astro';

const SITE_URL = process.env.SITE_URL || 'https://example.com';

// https://astro.build/config
export default defineConfig({
  site: SITE_URL,
  // The dev toolbar overlays every page with a high-z-index host that swallows
  // pointer events on transparent regions — clicks on the navbar search and
  // theme toggle sporadically land on the toolbar instead. We don't need the
  // toolbar for this project's workflow, so keep it off in dev.
  devToolbar: { enabled: false },
  integrations: [
    sitemap(),
    AstroPWA({
      registerType: 'autoUpdate',
      manifest: {
        name: 'Marvelous journal of a wanderer',
        short_name: 'Wanderer',
        description: "Photographies de rue, poésie et carnets d'errance.",
        theme_color: '#050505',
        background_color: '#050505',
        display: 'standalone',
        lang: 'fr',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
        ],
      },
      workbox: {
        navigateFallback: '/',
        globPatterns: ['**/*.{css,js,html,svg,png,ico,webmanifest}'],
        runtimeCaching: [
          {
            urlPattern: /^\/_vercel\/image/,
            handler: 'StaleWhileRevalidate',
            options: { cacheName: 'vercel-images', expiration: { maxEntries: 100 } },
          },
        ],
      },
    }),
  ],
  image: {
    remotePatterns: [
      { protocol: "https", hostname: "**.infomaniak.com" },
      { protocol: "https", hostname: "**.kdrive.infomaniak.com" },
      { protocol: "https", hostname: "images.unsplash.com" },
      { protocol: "https", hostname: "cdn.pixabay.com" },
    ],
  },
});

