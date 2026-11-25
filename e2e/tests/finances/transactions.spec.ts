import { expect, test, waitForApi } from '../../fixtures';

// B4 Finances. There is no "edit transaction" in the product (only add and delete), so the
// edit-shaped paths are: record a balance over an existing day, and edit the account.

test.describe('finances', { tag: ['@finances'] }, () => {
  test('add an account through the UI; it persists across reload', async ({ page }) => {
    await page.goto('/finances');
    await page.getByRole('button', { name: 'Add your first account' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Name').fill('Rainy day');
    await dialog.getByRole('textbox', { name: /balance/i }).fill('250');
    const saved = waitForApi(page, 'POST /money/accounts');
    await dialog.getByRole('button', { name: 'Add account' }).click();
    await saved;
    await expect(page).toHaveURL(/\/finances\/accounts\/\d+$/); // the app opens the new account
    await page.reload();
    await expect(page.getByText('$250.00').first()).toBeVisible();
    await page.goto('/finances');
    await expect(page.getByRole('link', { name: /Rainy day/ })).toBeVisible();
  });

  test('deposit with "also update balance": row and balance persist across reload', async ({ page, api, db }) => {
    const acct = await api.createAccount({ name: 'Savings', opening_balance: 100 });
    await page.goto(`/finances/accounts/${acct.ID}`);
    await expect(page.getByText('$100.00').first()).toBeVisible();

    await page.getByRole('textbox', { name: 'Amount' }).fill('40');
    await page.getByRole('textbox', { name: 'Note' }).fill('Birthday');
    const saved = waitForApi(page, `POST /money/accounts/${acct.ID}/transactions`);
    await page.getByRole('button', { name: 'Add deposit' }).click();
    await saved;
    await expect(page.getByText('Birthday')).toBeVisible();
    await expect(page.getByText('$140.00').first()).toBeVisible();

    await page.reload();
    await expect(page.getByText('Birthday')).toBeVisible();
    await expect(page.getByRole('listitem').filter({ hasText: 'Birthday' })).toBeVisible();
    await expect(page.getByText('$140.00').first()).toBeVisible();
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM investment_deposits WHERE account_id = $1`, [acct.ID]);
    expect(rows[0].n).toBe(1);
  });

  test('delete a transaction removes only the record; the balance stays', async ({ page, api }) => {
    const acct = await api.createAccount({ name: 'Savings', opening_balance: 100 });
    await api.addTransaction(acct.ID, { kind: 'deposit', amount: 25, date: '2026-01-12', note: 'Gift' });
    await page.goto(`/finances/accounts/${acct.ID}`);
    await expect(page.getByText('Gift')).toBeVisible();

    const gone = waitForApi(page, 'DELETE /money/accounts/:id/transactions/:tid');
    await page.getByRole('button', { name: 'Delete transaction' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Delete' }).click();
    await gone;
    await expect(page.getByText('Gift')).toHaveCount(0);

    await page.reload();
    await expect(page.getByText('No deposits or withdrawals recorded')).toBeVisible();
    await expect(page.getByText('$100.00').first()).toBeVisible();
  });

  test('edit the account name; persists across reload', async ({ page, api }) => {
    const acct = await api.createAccount({ name: 'Old name', opening_balance: 10 });
    await page.goto(`/finances/accounts/${acct.ID}`);
    await page.getByRole('navigation', { name: 'Page actions' }).getByRole('button', { name: 'Edit' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Name').fill('New name');
    const saved = waitForApi(page, 'PUT /money/accounts/:id');
    await dialog.getByRole('button', { name: 'Save' }).click();
    await saved;
    await page.goto('/finances');
    await expect(page.getByRole('link', { name: /New name/ })).toBeVisible();
    await expect(page.getByRole('link', { name: /Old name/ })).toHaveCount(0);
  });

  // The deposit and the new balance are written in one request, so if the balance write fails the
  // deposit is rolled back too and a retry cannot double-count it.
  test('deposit whose balance update fails rolls the transaction back (atomic)', async ({ page, api, e2e, db }) => {
    const acct = await api.createAccount({ name: 'Savings', opening_balance: 100 });
    await e2e.fault({ table: 'account_balances', op: 'create', times: 1 });
    await page.goto(`/finances/accounts/${acct.ID}`);
    await page.getByRole('textbox', { name: 'Amount' }).fill('40');
    await page.getByRole('button', { name: 'Add deposit' }).click();
    await expect(page.getByText(/e2e fault: create account_balances/)).toBeVisible();
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM investment_deposits WHERE account_id = $1`, [acct.ID]);
    expect(rows[0].n).toBe(0);
  });
});
