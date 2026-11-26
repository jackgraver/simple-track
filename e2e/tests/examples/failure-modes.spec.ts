import { expect, test, waitForApi } from '../../fixtures';

// One worked example per way of producing a failure for real (README, "Trigger the real cause").
// They double as specs for grocery's optimistic add: it shows the row at once, retries twice by
// itself (TanStack retry: 2, ~1s then ~2s backoff), and the server de-duplicates by client_id.

const CREATE = 'POST /tracking/grocery/items';

type Page = import('@playwright/test').Page;
const created = (page: Page) => waitForApi(page, CREATE);
const failedCreate = (page: Page) =>
  page.waitForResponse((r) => r.request().method() === 'POST' && r.url().includes('/api/tracking/grocery/items') && r.status() === 500);

/** Let `n` failed attempts land, skipping the retry backoff after each one. */
async function skipBackoff(page: Page, n: number) {
  for (let i = 0; i < n; i++) {
    await failedCreate(page);
    await page.clock.runFor(5_000);
  }
}

async function addItem(page: Page, name: string) {
  const input = page.getByRole('textbox', { name: 'Add grocery item' });
  await input.fill(name);
  await input.press('Enter');
}

test.describe('grocery add: failure modes', { tag: ['@examples', '@grocery'] }, () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/grocery');
    await expect(page.getByRole('textbox', { name: 'Add grocery item' })).toBeVisible();
  });

  test('golden path: the item persists across a reload', async ({ page }) => {
    const saved = created(page);
    await addItem(page, 'Milk');
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
    await saved;
    await page.reload();
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
  });

  test('DB fault, transient: one failed insert is retried and saved', async ({ page, e2e, db }) => {
    await e2e.fault({ table: 'grocery_items', op: 'create', times: 1 });
    const saved = created(page);
    const failures = skipBackoff(page, 1);
    await addItem(page, 'Milk');
    await failures;
    await saved;
    await expect(page.getByRole('button', { name: 'Retry' })).toHaveCount(0); // no error surfaced
    await page.reload();
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM grocery_items WHERE name = 'Milk'`);
    expect(rows[0].n).toBe(1);
  });

  test('DB fault, persistent: the row rolls back with Retry, and Retry saves it', async ({ page, e2e }) => {
    await e2e.fault({ table: 'grocery_items', op: 'create', times: 3 }); // the first try and both auto-retries
    const failures = skipBackoff(page, 3);
    await addItem(page, 'Milk');
    await expect(page.getByText('Milk', { exact: true })).toBeVisible(); // optimistic
    await failures;
    const toast = page.getByRole('button', { name: 'Retry' });
    await expect(toast).toBeVisible();
    await expect(page.getByText('Internal server error', { exact: true })).toBeVisible(); // the server's own message
    await expect(page.getByText('Milk', { exact: true })).toHaveCount(0); // rolled back

    const saved = created(page);
    await toast.click();
    await saved;
    await page.reload();
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
  });

  test('hold before: the row shows while the request is in flight', async ({ page, e2e }) => {
    const gate = await e2e.hold(CREATE);
    await addItem(page, 'Milk');
    await gate.arrived();
    await expect(page.getByText('Milk', { exact: true })).toBeVisible(); // before the server answered
    const saved = created(page);
    await gate.release();
    await saved;
    await page.reload();
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
  });

  test('hold after + drop: saved but unacknowledged, the retry does not duplicate it', async ({ page, e2e, db }) => {
    const gate = await e2e.hold(CREATE, { phase: 'after' });
    await addItem(page, 'Milk');
    await gate.arrived(); // the insert has committed; the response is held
    const saved = created(page);
    await gate.drop(); // the browser gets a network error and retries with the same client_id
    await page.clock.fastForward(5_000);
    await saved;
    await page.reload();
    await expect(page.getByText('Milk', { exact: true })).toHaveCount(1);
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM grocery_items WHERE name = 'Milk'`);
    expect(rows[0].n).toBe(1);
  });

  test('offline: the add waits and is sent when the connection returns', async ({ page, context }) => {
    await context.setOffline(true);
    await addItem(page, 'Milk');
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
    const saved = created(page);
    await context.setOffline(false);
    await saved;
    await page.reload();
    await expect(page.getByText('Milk', { exact: true })).toBeVisible();
  });
});
