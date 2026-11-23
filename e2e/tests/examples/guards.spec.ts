import { expect, test } from '../../fixtures';

// Self-tests for the suite's guards: each `test.fail()` case must be failed by the fixtures, so a
// guard that silently stops working turns this file red.

test.describe('fixture guards', { tag: '@infra' }, () => {
  test('an armed fault that never fires fails the test', async ({ e2e }) => {
    test.fail();
    await e2e.fault({ table: 'grocery_items', op: 'create' });
  });

  test('an optional fault may stay unused', async ({ e2e }) => {
    await e2e.fault({ table: 'grocery_items', op: 'create', optional: true });
  });

  test('an unreleased hold fails the test', async ({ e2e }) => {
    test.fail();
    await e2e.hold('POST /tracking/grocery/items');
  });

  test('faults and holds are validated when armed', async ({ e2e }) => {
    await expect(e2e.fault({ table: 'grocery_item', op: 'create' })).rejects.toThrow(/no table .*grocery_item/);
    await expect(e2e.hold('POST /tracking/grocery/item')).rejects.toThrow(/no route matches/);
  });

  test('a console error fails the test', async ({ page }) => {
    test.fail();
    await page.goto('/home');
    await page.evaluate(() => console.error('boom'));
  });

  test.describe('with allowedErrors', () => {
    test.use({ allowedErrors: [/boom/] });
    test('an allowed console error passes', async ({ page }) => {
      await page.goto('/home');
      await page.evaluate(() => console.error('boom'));
    });
  });

  test('server and browser clocks both read the epoch', async ({ page, api }) => {
    await page.goto('/home');
    const browserNow = await page.evaluate(() => new Date().toISOString());
    expect(browserNow.slice(0, 13)).toBe('2026-01-15T17'); // noon EST
    const me = await api.get<{ user?: unknown }>('/auth/me');
    expect(me).toBeTruthy();
  });

  test('every test starts from the seed', async ({ db }) => {
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM grocery_items`);
    expect(rows[0].n).toBe(0);
  });
});
