import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:4321',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
      testIgnore: /gestures\.spec\.ts/,
    },
    {
      name: 'mobile',
      // Pixel 7 uses Chromium (preinstalled) and still emulates touch; avoids
      // the WebKit download that iPhone devices trigger.
      use: { ...devices['Pixel 7'] },
      testMatch: /gestures\.spec\.ts/,
    },
  ],
  webServer: {
    command: 'node scripts/generate-search-index.mjs && npm run dev',
    url: 'http://localhost:4321',
    reuseExistingServer: !process.env.CI,
    timeout: 30000,
    env: {
      ASTRO_DISABLE_DEV_TOOLBAR: 'true',
    },
  },
});
