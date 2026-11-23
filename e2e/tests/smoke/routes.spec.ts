import { expect, test } from '../../fixtures';

// Every top-level page renders for a signed-in user, with no console errors, page errors or 5xx
// (the fixtures' console guard fails the test on any of those).
const routes = [
  '/home',
  '/gym',
  '/gym/plans',
  '/gym/progression',
  '/gym/weight',
  '/gym/steps',
  '/diet',
  '/diet/targets',
  '/grocery',
  '/finances',
  '/media',
  '/settings/profile',
];

for (const route of routes) {
  test(`${route} renders`, { tag: '@smoke' }, async ({ page }) => {
    await page.goto(route);
    await expect(page.getByRole('main')).toBeVisible();
    await page.waitForLoadState('networkidle'); // let the page's queries land so a 5xx is seen
    await expect(page).toHaveURL(new RegExp(route.replace(/\//g, '\\/') + '(\\?|$)'));
  });
}

test('/ redirects to the gym', { tag: '@smoke' }, async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveURL(/\/gym/);
});

test('unknown routes show not-found', { tag: '@smoke' }, async ({ page }) => {
  await page.goto('/definitely-not-a-page');
  await expect(page.getByRole('main')).toBeVisible();
});
