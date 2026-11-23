import { expect, test } from '../../fixtures';

// Every top-level page still renders, without errors, against two years of history and awkward
// values (seed profile "large"). A first line of defence for the performance work; timing budgets
// can be added here once the performance project sets them.
test.use({ seed: 'large' });

const routes = ['/gym', '/gym/progression', '/gym/weight', '/gym/steps', '/diet', '/grocery', '/finances', '/media'];

for (const route of routes) {
  test(`${route} renders with the large profile`, { tag: ['@scale'] }, async ({ page }) => {
    await page.goto(route);
    await expect(page.getByRole('main')).toBeVisible();
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(new RegExp(route.replace(/\//g, '\\/') + '(\\?|$)'));
  });
}

test('the large profile loads the expected volume', { tag: ['@scale'] }, async ({ db }) => {
  const { rows } = await db.query(`SELECT
    (SELECT count(*) FROM logged_sets)::int AS sets,
    (SELECT count(*) FROM days)::int AS days,
    (SELECT count(*) FROM media_items)::int AS media`);
  expect(rows[0].sets).toBeGreaterThan(3000);
  expect(rows[0].days).toBe(730);
  expect(rows[0].media).toBe(300);
});
