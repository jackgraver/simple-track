import { expect, test } from '../../fixtures';

// Optimistic UI: hold the request, assert the instant state, release, then assert the state
// after a reload. The hold makes "before the server answered" deterministic (no sleeps).
// Pending grocery rows carry aria-busy="true" (ListRow `pending`), which also tells us when the
// write has settled without a network wait.

test.describe('optimistic UI', { tag: ['@optimistic'] }, () => {
  test.describe('grocery', { tag: ['@grocery'] }, () => {
    test('add: row shows busy while held, settles on release, persists', async ({ page, e2e }) => {
      await page.goto('/grocery');
      const gate = await e2e.hold('POST /tracking/grocery/items');
      const input = page.getByRole('textbox', { name: 'Add grocery item' });
      await input.fill('Milk');
      await input.press('Enter');
      await gate.arrived();

      const row = page.getByRole('listitem').filter({ hasText: 'Milk' });
      await expect(row).toBeVisible();
      await expect(row).toHaveAttribute('aria-busy', 'true');

      await gate.release();
      await expect(row).not.toHaveAttribute('aria-busy', 'true'); // settled: safe to reload
      await page.reload();
      await expect(page.getByRole('listitem').filter({ hasText: 'Milk' })).toBeVisible();
    });

    test('complete: checked at once, item gone after reload', async ({ page, api, e2e }) => {
      const item = await api.createGroceryItem('Eggs');
      await page.goto('/grocery');
      const gate = await e2e.hold(`PATCH /tracking/grocery/items/:id/complete`);
      const box = page.getByRole('checkbox', { name: 'Eggs' });
      await box.check();
      await gate.arrived();
      await expect(box).toBeChecked(); // before the server answered
      await expect(page.getByRole('listitem').filter({ hasText: 'Eggs' })).toHaveAttribute('aria-busy', 'true');

      await gate.release();
      await page.clock.runFor(2_000); // past the strike-through delay: the row leaves the list
      await expect(page.getByText('Eggs', { exact: true })).toHaveCount(0);
      await page.reload();
      await expect(page.getByText('Eggs', { exact: true })).toHaveCount(0);
    });
  });

  test.describe('weight', { tag: ['@gym'] }, () => {
    // The cache is the only instant signal here (the chart is a canvas): saving a second time
    // asks to overwrite, and the dialog quotes the value the optimistic write put in the cache.
    test('save: the overwrite prompt knows the value before the server answered, and after reload', async ({ page, api, e2e }) => {
      await api.saveWeight('2026-01-10', 180); // gives the slider a starting value of 180
      await page.goto('/gym/weight');
      const save = page.getByRole('button', { name: 'Save', exact: true });
      const overwrite = page.getByRole('dialog').getByText(/You already logged 180 lbs/);

      const gate = await e2e.hold('POST /tracking/weight');
      await save.click();
      await gate.arrived();
      await save.click();
      await expect(overwrite).toBeVisible(); // today's log exists only optimistically
      await page.getByRole('dialog').getByRole('button', { name: 'Cancel' }).click();

      await gate.release();
      await expect(page.getByText('Saved', { exact: true })).toBeVisible();
      await page.reload();
      await save.click();
      await expect(overwrite).toBeVisible(); // now from the server
      await page.getByRole('dialog').getByRole('button', { name: 'Cancel' }).click();
    });
  });
});
