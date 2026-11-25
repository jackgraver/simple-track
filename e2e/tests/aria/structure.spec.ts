import { expect, test } from '../../fixtures';

// Structure snapshots (TASKS C4): the accessibility tree of a few regions that must not silently
// lose items, not pixels. Kept deliberately small because the UI overhaul churns markup. After an
// intended change, review and accept with:
//   npx playwright test tests/aria --update-snapshots
// Files: tests/aria/structure.spec.ts-snapshots/<name>-<project>.aria.yml (desktop and mobile differ).

test.describe('structure', { tag: ['@aria'] }, () => {
  test('main navigation', async ({ page, isMobile }) => {
    await page.goto('/gym');
    if (isMobile) {
      // phones: the nav lives in the drawer. Retry the tap: an early click can land before the app hydrates.
      await expect(async () => {
        await page.getByRole('button', { name: 'Open menu' }).click();
        await expect(page.getByRole('dialog')).toBeVisible({ timeout: 1_000 });
      }).toPass();
    }
    const nav = page.getByRole('navigation', { name: 'Main' }).filter({ visible: true });
    await expect(nav).toMatchAriaSnapshot({ name: 'main-nav.aria.yml' });
  });

  test('gym week strip at the epoch', async ({ page }) => {
    await page.goto('/gym');
    await expect(page.getByRole('list', { name: 'Days this week' })).toMatchAriaSnapshot({ name: 'gym-week.aria.yml' });
  });
});
