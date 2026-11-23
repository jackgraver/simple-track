import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// Seed: "Legs" is today's plan (the epoch is a Thursday), Back Squat is exercise 9. Weight starts
// at 0 and the slider's reps tick buttons are the only way to set reps, so `setReps` clicks one.
const SQUAT = 9;
const TODAY = '2026-01-15';
const LAST_WEEK = '2026-01-08';
const SETS = 'POST /workout/sets';

const setsList = (page: Page) => page.getByRole('region', { name: 'Sets' });
const setRow = (page: Page, n: number) => setsList(page).getByText(new RegExp(`^Set ${n} ·`));

async function open(page: Page) {
  await page.goto(`/gym/logging/${SQUAT}`);
  await expect(page.getByRole('heading', { name: 'Back Squat' })).toBeVisible();
}
// The slider is scroll-driven: a tick click scrolls and the value follows. The primary button echoes
// the value ("Log set 1 · 8 reps"), so wait for it before moving on.
async function setReps(page: Page, reps: number) {
  await page.getByRole('region', { name: 'Reps' }).getByRole('button', { name: String(reps), exact: true }).click();
  const primary = page.getByRole('button', { name: new RegExp(`· ${reps} reps$`) });
  await expect(primary).toBeVisible();
  // Let the smooth scroll finish (no scroll events for ~300ms); mid-scroll events can still move the value.
  await page.evaluate(
    () =>
      new Promise<void>((done) => {
        let timer = setTimeout(done, 300);
        document.addEventListener(
          'scroll',
          () => {
            clearTimeout(timer);
            timer = setTimeout(done, 300);
          },
          { capture: true },
        );
      }),
  );
  await expect(primary).toBeVisible();
}

test.describe('gym logging', { tag: ['@gym'] }, () => {
  test('log a set: it shows at once and survives a reload', async ({ page, db }) => {
    await open(page);
    await setReps(page, 8);
    const saved = waitForApi(page, SETS);
    await page.getByRole('button', { name: /^Log set 1/ }).click();
    await expect(setRow(page, 1)).toContainText('× 8');
    await saved;
    await page.reload();
    await expect(setRow(page, 1)).toContainText('× 8');
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets WHERE reps = 8`);
    expect(rows[0].n).toBe(1);
  });

  test('edit: changing a saved set persists', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 100, date: TODAY });
    await open(page);
    await expect(setRow(page, 1)).toContainText('100 × 8');
    await setRow(page, 1).click();
    await expect(page.getByRole('button', { name: /^Save set 1 · 8 reps$/ })).toBeVisible(); // edit state settled
    await setReps(page, 9);
    const saved = waitForApi(page, 'PATCH /workout/sets/:id');
    await page.getByRole('button', { name: /^Save set 1/ }).click();
    await expect(setRow(page, 1)).toContainText('100 × 9');
    await saved;
    await page.reload();
    await expect(setRow(page, 1)).toContainText('100 × 9');
  });

  test('delete: only the deleted set is gone after a reload', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 100, date: TODAY });
    await api.logSet({ exerciseId: SQUAT, reps: 9, weight: 100, date: TODAY });
    await open(page);
    await expect(setRow(page, 2)).toContainText('× 9');
    const deleted = waitForApi(page, 'DELETE /workout/sets/:id');
    await page.getByRole('button', { name: 'Delete set 1' }).click();
    await deleted;
    await page.reload();
    await expect(setRow(page, 1)).toContainText('× 9');
    await expect(setRow(page, 2)).toHaveCount(0);
  });

  test('history: last session\'s sets show next to today\'s', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 100, date: LAST_WEEK });
    await api.logSet({ exerciseId: SQUAT, reps: 9, weight: 100, date: TODAY });
    await open(page);
    await expect(setRow(page, 1)).toContainText('100 × 9');
    await expect(setsList(page)).toContainText('Last 100 × 8');
  });

  test('mobile: on a short phone the Log button is pinned above the tab bar', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'sticky bar is a phone layout');
    await page.setViewportSize({ width: 390, height: 480 });
    await open(page);
    const log = page.getByRole('button', { name: /^Log set 1/ });
    await expect(log).toBeInViewport({ ratio: 1 });
    const tabbar = await page.getByRole('navigation', { name: /main/i }).last().boundingBox();
    const box = await log.boundingBox();
    if (tabbar && box) expect(box.y + box.height).toBeLessThanOrEqual(tabbar.y + 1);
  });

  test('dropped response: the retry does not duplicate the set', async ({ page, e2e, db }) => {
    await open(page);
    await setReps(page, 8);
    const gate = await e2e.hold(SETS, { phase: 'after' });
    await page.getByRole('button', { name: /^Log set 1/ }).click();
    await gate.arrived(); // committed; the response is held
    const saved = waitForApi(page, SETS);
    await gate.drop();
    await page.clock.fastForward(5_000);
    await saved;
    await page.reload();
    await expect(setRow(page, 1)).toContainText('× 8');
    await expect(setRow(page, 2)).toHaveCount(0);
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets`);
    expect(rows[0].n).toBe(1);
  });
});
