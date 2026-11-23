import { expect, test, waitForApi } from '../../fixtures';

// Golden paths for /grocery. Add-failure modes live in tests/examples/failure-modes.spec.ts (A8).

test.describe('grocery list', { tag: ['@grocery'] }, () => {
  const input = (page: import('@playwright/test').Page) => page.getByRole('textbox', { name: 'Add grocery item' });
  const row = (page: import('@playwright/test').Page, name: string) => page.getByText(name, { exact: true });

  test('empty list shows the empty state', async ({ page }) => {
    await page.goto('/grocery');
    await expect(page.getByText('Nothing on the list')).toBeVisible();
  });

  test('add several items; they persist in order across a reload', async ({ page }) => {
    await page.goto('/grocery');
    for (const name of ['Milk', 'Eggs']) {
      const saved = waitForApi(page, 'POST /tracking/grocery/items');
      await input(page).fill(name);
      await input(page).press('Enter');
      await saved;
    }
    await page.reload();
    await expect(row(page, 'Milk')).toBeVisible();
    await expect(row(page, 'Eggs')).toBeVisible();
    await expect(page.getByRole('checkbox')).toHaveCount(2);
  });

  test('check off an item: it leaves the list and stays gone after reload', async ({ page, api, db }) => {
    const milk = await api.createGroceryItem('Milk');
    await api.createGroceryItem('Eggs');
    await page.goto('/grocery');
    const done = waitForApi(page, 'PATCH /tracking/grocery/items/:id/complete');
    await page.getByRole('checkbox', { name: 'Milk' }).check();
    await done;
    await expect(row(page, 'Milk')).toHaveCount(0); // after the strike-through delay
    await expect(row(page, 'Eggs')).toBeVisible();
    await page.reload();
    await expect(row(page, 'Milk')).toHaveCount(0);
    await expect(row(page, 'Eggs')).toBeVisible();
    // Completing keeps the row (history/suggestions), it only stamps completed_at.
    const { rows } = await db.query('SELECT completed_at FROM grocery_items WHERE id = $1', [milk.ID]);
    expect(rows[0].completed_at).not.toBeNull();
  });

  // The page has no delete control today; deletion is API-only. Pin that a deleted item is gone from
  // the UI after a load, and that the delete is soft (deleted_at set, row kept).
  test('an item deleted through the API is gone from the list', async ({ page, api, db }) => {
    const milk = await api.createGroceryItem('Milk');
    await api.createGroceryItem('Eggs');
    await api.deleteGroceryItem(milk.ID);
    await page.goto('/grocery');
    await expect(row(page, 'Eggs')).toBeVisible();
    await expect(row(page, 'Milk')).toHaveCount(0);
    const { rows } = await db.query('SELECT deleted_at FROM grocery_items WHERE id = $1', [milk.ID]);
    expect(rows[0].deleted_at).not.toBeNull();
  });
});
