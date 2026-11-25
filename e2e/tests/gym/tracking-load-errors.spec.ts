import type { Page } from '@playwright/test';
import { expect, test } from '../../fixtures';

// Cold-load failures on the tracking pages. Queries retry twice on 5xx (frontend/src/api/queryClient.ts,
// shouldRetryQuery, backoff ~1s then ~2s), so a fault with `times: 3` fails every automatic attempt and
// the page's ErrorState shows; Retry then reaches the database again and succeeds.
const ATTEMPTS = 3;

/** Count failed GETs to `path` and skip the retry backoff after each one, until all attempts failed. */
async function failAllAttempts(page: Page, path: RegExp, open: () => Promise<unknown>) {
  let failures = 0;
  page.on('response', (r) => {
    if (r.request().method() === 'GET' && path.test(new URL(r.url()).pathname) && r.status() === 500) failures++;
  });
  await open();
  for (let i = 1; i <= ATTEMPTS; i++) {
    await expect.poll(() => failures).toBe(i);
    if (i < ATTEMPTS) await page.clock.runFor(5_000);
  }
}

test.describe('gym tracking: load errors recover with Retry', { tag: ['@gym'] }, () => {
  test('weight', async ({ page, api, e2e }) => {
    await api.saveWeight('2026-01-10', 180);
    await e2e.fault({ table: 'body_weight_logs', op: 'query', times: ATTEMPTS });
    await failAllAttempts(page, /\/api\/tracking\/weight$/, () => page.goto('/gym/weight'));
    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load your weight history" });
    await expect(alert).toBeVisible();
    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(alert).toHaveCount(0);
    await expect(page.getByRole('heading', { name: 'Previous' })).toBeVisible();
    await expect(page.getByText('No history through this day yet.')).toHaveCount(0); // the Jan 10 log loaded
  });

  test('steps', async ({ page, e2e }) => {
    await e2e.fault({ table: 'step_logs', op: 'query', times: ATTEMPTS });
    await failAllAttempts(page, /\/api\/tracking\/steps$/, () => page.goto('/gym/steps'));
    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load your step history" });
    await expect(alert).toBeVisible();
    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(alert).toHaveCount(0);
    await expect(page.getByRole('heading', { name: 'Previous' })).toBeVisible();
  });

  test('progression', async ({ page, e2e }) => {
    // The progression query is a gorm Scan on logged_exercises: the "row" callback.
    await e2e.fault({ table: 'logged_exercises', op: 'row', times: ATTEMPTS });
    await failAllAttempts(page, /\/api\/workout\/exercises\/progression\/\d+$/, () => page.goto('/gym/progression'));
    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load progression" });
    await expect(alert).toBeVisible();
    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(alert).toHaveCount(0);
    await expect(page.getByText('No data in the last month')).toBeVisible();
  });
});
