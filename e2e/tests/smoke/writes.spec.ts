import type { Page } from '@playwright/test';
import { expect, test } from '../../fixtures';

// B1: one minimal write per module, verified after a reload. Deep scenarios live in tests/<module>/.
// Each write waits for its own successful response (by URL suffix) before reloading, because
// optimistic rows show before the request is sent.
const saved = (page: Page, method: string, pathRe: RegExp) =>
  page.waitForResponse(
    (r) => r.request().method() === method && pathRe.test(new URL(r.url()).pathname) && r.ok(),
  );

test('gym: log steps', { tag: ['@smoke', '@gym'] }, async ({ page, api }) => {
  await page.goto('/gym/steps');
  await page.getByLabel(/^Steps \(/).fill('8420');
  const done = saved(page, 'POST', /\/steps$/);
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await done;
  await page.reload(); // the form isn't prefilled, so check persistence through the API
  expect(JSON.stringify(await api.get('/tracking/steps'))).toContain('8420');
});

test('diet: log a custom water amount', { tag: ['@smoke', '@diet'] }, async ({ page }) => {
  await page.goto('/diet');
  await page.getByRole('button', { name: 'Other' }).or(page.getByRole('button', { name: 'Log water' })).click();
  await page.getByLabel(/^Custom amount/).fill('12');
  const done = saved(page, 'POST', /\/water$/);
  await page.getByRole('button', { name: 'Log custom amount' }).click();
  await done;
  await page.reload();
  await page.getByRole('button', { name: /^Water entries for/ }).click();
  await expect(page.getByRole('dialog')).toContainText('12');
});

test('finances: add an account', { tag: ['@smoke', '@finances'] }, async ({ page }) => {
  await page.goto('/finances');
  await page.getByRole('button', { name: /Add (your first )?account/i }).first().click();
  await page.getByLabel('Name').fill('Smoke chequing');
  const done = saved(page, 'POST', /\/accounts$/);
  await page.getByRole('dialog').getByRole('button', { name: /^(Add|Save|Create)/ }).click();
  await done;
  await page.goto('/finances'); // creating navigates to the account page; check the list after a fresh load
  await expect(page.getByText('Smoke chequing').locator('visible=true').first()).toBeVisible();
});

test('media: add an item', { tag: ['@smoke', '@media'] }, async ({ page }) => {
  await page.goto('/media/new');
  await page.getByLabel('Title').fill('Smoke Movie');
  const done = saved(page, 'POST', /\/media$/);
  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await done;
  await page.goto('/media');
  await expect(page.getByText('Smoke Movie').first()).toBeVisible();
});

test('grocery: add an item', { tag: ['@smoke', '@grocery'] }, async ({ page }) => {
  await page.goto('/grocery');
  const input = page.getByRole('textbox', { name: 'Add grocery item' });
  const done = saved(page, 'POST', /\/grocery\/items$/);
  await input.fill('Smoke eggs');
  await input.press('Enter');
  await done;
  await page.reload();
  await expect(page.getByText('Smoke eggs', { exact: true })).toBeVisible();
});

test('settings: save the body profile', { tag: ['@smoke', '@settings'] }, async ({ page }) => {
  await page.goto('/settings/profile');
  await page.getByLabel('Age').fill('41');
  const done = saved(page, 'PUT', /\/profile$/);
  await page.getByRole('button', { name: 'Save profile' }).click();
  await done;
  await page.reload();
  await expect(page.getByLabel('Age')).toBeVisible({ timeout: 15_000 });
  await expect(page.getByLabel('Age')).toHaveValue('41');
});
