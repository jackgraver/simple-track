import { expect, test, waitForApi } from '../../fixtures';
import { LAST_WEEK, LEGS, SETS, TODAY, expectReps, logButton, openExercise, saveButton, setReps, setRow, setsList } from './logger';

// Seed: "Legs" is today's plan (the epoch is a Thursday), Back Squat is exercise 9. Weight starts
// at 0 and the slider's reps tick buttons are the only way to set reps (helpers in ./logger).
const SQUAT = LEGS.squat.id;
const open = (page: import('@playwright/test').Page) => openExercise(page, LEGS.squat);

test.describe('gym logging', { tag: ['@gym'] }, () => {
  test('log a set: it shows at once and survives a reload', async ({ page, db }) => {
    await open(page);
    await setReps(page, 8);
    const saved = waitForApi(page, SETS);
    await logButton(page, 1).click();
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
    await expect(saveButton(page, 1)).toBeVisible(); // edit state settled
    await expectReps(page, 8);
    await setReps(page, 9);
    const saved = waitForApi(page, 'PATCH /workout/sets/:id');
    await saveButton(page, 1).click();
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
    const log = logButton(page, 1);
    await expect(log).toBeInViewport({ ratio: 1 });
    const tabbar = await page.getByRole('navigation', { name: /main/i }).last().boundingBox();
    const box = await log.boundingBox();
    if (tabbar && box) expect(box.y + box.height).toBeLessThanOrEqual(tabbar.y + 1);
  });

  test('dropped response: the retry does not duplicate the set', async ({ page, e2e, db }) => {
    await open(page);
    await setReps(page, 8);
    const gate = await e2e.hold(SETS, { phase: 'after' });
    await logButton(page, 1).click();
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
