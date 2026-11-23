import { expect, test } from '../../fixtures';

// Every top-level page renders for a signed-in user, with no console errors, page errors or 5xx
// (the fixtures' console guard fails the test on any of those).
// Every route without params (redirect-only ones are covered by their targets). When adding a route
// to frontend/src/router/index.ts, add it here: a page that crashes on load is otherwise only caught
// if some deeper spec happens to open it.
const routes = [
  '/home',
  '/gym',
  '/gym/plans',
  '/gym/logging',
  '/gym/logging/cardio',
  '/gym/progression',
  '/gym/weight',
  '/gym/steps',
  '/diet',
  '/diet/log',
  '/diet/edit-planned',
  '/diet/targets',
  '/diet/water-presets',
  '/grocery',
  '/finances',
  '/finances/rules',
  '/media',
  '/media/new',
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
