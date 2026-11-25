import { E2E_EPOCH, expect, test } from '../../fixtures';

// Only this lane logs in for real. Login is rate-limited (5/user/15min; >5 usernames/IP/10min
// blocks the IP for 24h) and the limiter is per backend process, so keep the number of real logins
// per test and per worker small: one per test, all as the seeded user "dev", plus one bad password.
// Unauthenticated requests (/auth/me on load, a bad password, an expired token) are real 401s.
test.use({ authenticated: false, allowedErrors: [/status of 401/] });

const DAY_MS = 24 * 3600_000;

async function signIn(page: import('@playwright/test').Page, password = 'dev') {
  await page.getByLabel('Username').fill('dev');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
}

test('unauthenticated visit redirects to sign-in and back after login', { tag: ['@auth', '@smoke'] }, async ({ page }) => {
  await page.goto('/diet');
  await expect(page).toHaveURL(/\/auth\/signin\?redirect=\/diet/);
  const login = page.waitForResponse((r) => r.url().includes('/api/auth/login') && r.ok());
  await signIn(page);
  await login;
  await expect(page).toHaveURL(/\/diet$/);
  await expect(page.getByRole('main')).toBeVisible();

  await page.reload(); // the cookie persists the session
  await expect(page).toHaveURL(/\/diet$/);
});

test('wrong password shows an error and stays on sign-in', { tag: '@auth' }, async ({ page }) => {
  await page.goto('/auth/signin');
  await signIn(page, 'not-the-password');
  await expect(page.getByRole('alert')).toBeVisible();
  await expect(page).toHaveURL(/\/auth\/signin/);
});

test('logout ends the session', { tag: '@auth' }, async ({ page }) => {
  await page.goto('/auth/signin');
  await signIn(page);
  await expect(page).not.toHaveURL(/signin/);

  await page.getByRole('button', { name: 'Account menu for dev' }).click();
  const loggedOut = page.waitForResponse((r) => r.url().includes('/api/auth/logout') && r.ok());
  await page.getByRole('menuitem', { name: 'Sign out' }).click();
  await loggedOut;
  await expect(page).toHaveURL(/\/auth\/signin/);
  await expect(page.getByText('Signed out', { exact: true })).toBeVisible();

  await page.goto('/gym');
  await expect(page).toHaveURL(/\/auth\/signin/);
});

test('an expired session sends the user back to sign-in', { tag: '@auth' }, async ({ page, e2e }) => {
  await page.goto('/auth/signin');
  await signIn(page);
  await expect(page).not.toHaveURL(/signin/);

  // Token lifetime is 365 days on the server's clock (the browser cookie expiry uses real time).
  await e2e.setClock(new Date(E2E_EPOCH.getTime() + 366 * DAY_MS));
  await page.goto('/gym');
  await expect(page).toHaveURL(/\/auth\/signin\?redirect=\/gym/);
});

test('a session that expires mid-use redirects when the app next loads', { tag: '@auth' }, async ({ page, e2e }) => {
  await page.goto('/auth/signin');
  await signIn(page);
  await expect(page).not.toHaveURL(/signin/);
  await page.goto('/grocery');
  await expect(page.getByRole('main')).toBeVisible();

  await e2e.setClock(new Date(E2E_EPOCH.getTime() + 366 * DAY_MS));
  await page.goto('/diet'); // a full load re-checks /auth/me, which now answers 401
  await expect(page).toHaveURL(/\/auth\/signin/);
});
