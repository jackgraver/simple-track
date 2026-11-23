import { defineConfig, devices } from '@playwright/test';
import { TZ } from './infra/paths';

const CI = !!process.env.CI;

export default defineConfig({
  testDir: './tests',
  globalSetup: './global-setup.ts',
  fullyParallel: true,
  forbidOnly: CI,
  workers: Number(process.env.E2E_WORKERS) || (CI ? 2 : 6),
  retries: CI ? 1 : 0,
  failOnFlakyTests: CI,
  timeout: 30_000,
  expect: { timeout: 5_000 },
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    // baseURL is set per worker by the fixtures (the front door starts in global setup).
    timezoneId: TZ,
    locale: 'en-CA',
    serviceWorkers: 'block',
    contextOptions: { reducedMotion: 'reduce' },
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['Pixel 7'] } },
  ],
});
