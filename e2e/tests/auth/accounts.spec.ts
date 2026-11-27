import { expect, test } from '../../fixtures';

// Accounts (multi-user lane C6): the admin creates an account with a username only, texts the
// one-time link, the person sets a password through it and is signed in, and the link is dead
// afterwards. The seeded dev user is the admin (signed in through the bypass cookie).
// A dead link answers 410, which the browser logs as a failed resource.
// One regex: a two-element array would read as Playwright's [value, options] fixture tuple.
test.use({ allowedErrors: [/status of (401|410)/] });

test('create user -> open link -> set password -> signed in -> link no longer works', { tag: ['@auth', '@smoke'] }, async ({ page, context, db }) => {
  await page.goto('/settings/users');
  await expect(page.getByRole('heading', { name: 'Users' }).first()).toBeVisible();

  await page.getByLabel('Username').fill('Mia');
  await page.getByRole('button', { name: 'Create and get link' }).click();
  const dialog = page.getByRole('dialog', { name: 'Setup link for mia' });
  await expect(dialog).toBeVisible();
  const url = await dialog.getByLabel('Setup link').inputValue();
  expect(url).toMatch(/\/setup\/[A-Za-z0-9_-]{43}$/);
  await page.keyboard.press('Escape');
  await expect(page.getByRole('listitem').filter({ hasText: 'mia' })).toContainText('Setup link open until');

  // Only a hash is stored.
  const token = url.split('/setup/')[1];
  const { rows } = await db.query('SELECT token_hash FROM user_setup_tokens');
  expect(rows).toHaveLength(1);
  expect(rows[0].token_hash).not.toContain(token);

  // Mia's phone: nobody signed in.
  await context.clearCookies({ name: 'auth_token' });
  await page.goto(new URL(url).pathname);
  await expect(page.getByRole('heading', { name: 'Set up your account' })).toBeVisible();
  await expect(page.getByText('Choose a password for mia')).toBeVisible();

  await page.getByLabel('New password').fill('short');
  await page.getByLabel('Repeat password').fill('short');
  await page.getByRole('button', { name: 'Set password and sign in' }).click();
  await expect(page.getByText('Use at least 10 characters.')).toBeVisible();

  await page.getByLabel('New password').fill('mia-password-1');
  await page.getByLabel('Repeat password').fill('mia-password-1');
  const setup = page.waitForResponse((r) => r.url().includes('/api/auth/setup') && r.request().method() === 'POST' && r.ok());
  await page.getByRole('button', { name: 'Set password and sign in' }).click();
  await setup;
  await expect(page).not.toHaveURL(/\/setup\//);
  await expect(page.getByText("Welcome! You're signed in.").first()).toBeVisible();
  const me = await page.request.get('/api/auth/me');
  expect((await me.json()).username).toBe('mia');

  // The link is used up.
  await page.goto(new URL(url).pathname);
  await expect(page.getByText('This link no longer works')).toBeVisible();

  const { rows: events } = await db.query(
    `SELECT event FROM auth_events e JOIN users u ON u.id = e.user_id WHERE u.username = 'mia' ORDER BY e.id`,
  );
  expect(events.map((e) => e.event)).toEqual(['user_created', 'setup_link_created', 'setup_link_used']);
});

test('the admin can disable an account, which signs it out', { tag: '@auth' }, async ({ page, api, db }) => {
  // Arrange: an account with a password, through the API's setup flow.
  const created = await api.post<{ link: { path: string } }>('/admin/users', { username: 'leo' });
  const token = created.link.path.split('/setup/')[1];
  await api.post('/auth/setup', { token, password: 'leo-password-1' });

  await page.goto('/settings/users');
  const row = page.getByRole('listitem').filter({ hasText: 'leo' });
  await row.getByRole('button', { name: 'Actions for leo' }).click();
  await page.getByRole('menuitem', { name: 'Disable and sign out' }).click();
  await expect(page.getByText('leo is disabled and signed out').first()).toBeVisible();
  await expect(row.getByText('Disabled', { exact: true })).toBeVisible();

  const { rows } = await db.query(`SELECT disabled_at, token_version FROM users WHERE username = 'leo'`);
  expect(rows[0].disabled_at).not.toBeNull();
  expect(Number(rows[0].token_version)).toBe(3); // 1, +1 for the password, +1 for the disable

  // The audit view shows it.
  await expect(page.getByRole('listitem').filter({ hasText: 'Account disabled · leo' })).toBeVisible();
});

test.describe('signed out', () => {
  test.use({ authenticated: false });

  test('a made-up setup link shows that it no longer works', { tag: '@auth' }, async ({ page }) => {
    await page.goto('/setup/not-a-real-token');
    await expect(page.getByText('This link no longer works')).toBeVisible();
    await page.getByRole('link', { name: 'Go to sign in' }).click();
    await expect(page).toHaveURL(/\/auth\/signin/);
  });
});
