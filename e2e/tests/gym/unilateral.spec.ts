import { expect, test, waitForApi } from '../../fixtures';
import { LAST_WEEK, LEGS, LEGS_PLAN, SETS, TODAY, expectReps, openExercise, setReps, setsList } from './logger';

// Unilateral logging: each side is its own set, a set number is a left + right pair. "Log Set 1" asks
// which side that was (preselected with last session's), then sides alternate. The rest clock starts
// on the first side (red < 0:15, yellow < 0:25, green after) and keeps counting through the second
// side, so the 2:00 rest is measured from the first side.
const SQUAT = LEGS.squat;
const makeUnilateral = (db: { query: (sql: string) => Promise<unknown> }) =>
  db.query(`UPDATE exercises SET unilateral = true WHERE id = ${SQUAT.id}`);
const logSide = (page: import('@playwright/test').Page, n: number, side: 'Left' | 'Right') =>
  page.getByRole('button', { name: `Log Set ${n} · ${side}`, exact: true });
const pairCard = (page: import('@playwright/test').Page, n: number) =>
  setsList(page).getByRole('region', { name: `Set ${n}`, exact: true });
const half = (page: import('@playwright/test').Page, n: number, side: 'left' | 'right') =>
  pairCard(page, n).getByRole('button', { name: `Edit set ${n} ${side}` });
/** Log Set 1: confirm the first side in the dialog. */
async function logFirst(page: import('@playwright/test').Page, side?: 'Left' | 'Right') {
  await page.getByRole('button', { name: 'Log Set 1', exact: true }).click();
  const dialog = page.getByRole('dialog', { name: 'Which side was that?' });
  if (side) await dialog.getByText(side, { exact: true }).click();
  await dialog.getByRole('button', { name: 'Confirm' }).click();
}
const sides = async (db: { query: (sql: string) => Promise<{ rows: any[] }> }) =>
  (await db.query(`SELECT side, weight::float8 AS weight, reps::int AS reps FROM logged_sets ORDER BY id`)).rows;

test.describe('gym unilateral logging', { tag: ['@gym'] }, () => {
  test('toggle: the Edit dialog turns unilateral logging on', async ({ page, db }) => {
    await page.goto(`/gym/plans/${LEGS_PLAN}`);
    await page.getByRole('button', { name: `1 ${SQUAT.name}` }).click();
    const dialog = page.getByRole('dialog', { name: `Edit ${SQUAT.name}` });
    await dialog.getByRole('switch', { name: 'Unilateral' }).click();
    await dialog.getByRole('button', { name: 'Save' }).click();
    await expect(dialog).toBeHidden();
    const { rows } = await db.query(`SELECT unilateral FROM exercises WHERE id = ${SQUAT.id}`);
    expect(rows[0].unilateral).toBe(true);
    await openExercise(page, SQUAT);
    await expect(page.getByRole('button', { name: 'Log Set 1', exact: true })).toBeVisible();
    await expect(page.getByRole('region', { name: 'Side' })).toHaveCount(0);
  });

  test('flow: confirm the first side, then sides alternate on one clock', async ({ page, db }) => {
    await makeUnilateral(db);
    await openExercise(page, SQUAT);
    await setReps(page, 8);

    let saved = waitForApi(page, SETS);
    await logFirst(page, 'Right');
    await saved;
    await expect(logSide(page, 1, 'Left')).toBeVisible();
    await expect(page.getByRole('region', { name: 'Side' })).toHaveCount(0); // no picker: sides alternate
    const clock = page.getByRole('timer');
    await expect(clock).toHaveClass(/text-danger/);
    await page.clock.fastForward(15_000);
    await expect(clock).toHaveClass(/text-rest-near/);
    await page.clock.fastForward(10_000);
    await expect(clock).toHaveClass(/text-success/);

    saved = waitForApi(page, SETS);
    await logSide(page, 1, 'Left').click();
    await saved;
    // Rest counts on from the first side: no reset, normal stages (red under 1:00).
    await expect(clock).toHaveText(/^0:2\d$/);
    await expect(clock).toHaveClass(/text-danger/);
    await expect(logSide(page, 2, 'Right')).toBeVisible();
    await expect(setsList(page)).toContainText('Sets 1/');

    expect(await sides(db)).toEqual([
      { side: 'right', weight: 0, reps: 8 },
      { side: 'left', weight: 0, reps: 8 },
    ]);
    await page.reload();
    await expect(half(page, 1, 'left')).toContainText('8 reps');
    await expect(half(page, 1, 'right')).toContainText('8 reps');
    await expect(logSide(page, 2, 'Right')).toBeVisible();
  });

  test('first side: closing the dialog logs nothing', async ({ page, db }) => {
    await makeUnilateral(db);
    await openExercise(page, SQUAT);
    await page.getByRole('button', { name: 'Log Set 1', exact: true }).click();
    await page.keyboard.press('Escape');
    await expect(page.getByRole('dialog')).toBeHidden();
    expect(await sides(db)).toEqual([]);
  });

  test('edit: a side can be moved to the other side, or deleted', async ({ page, api, db }) => {
    await makeUnilateral(db);
    await api.logSet({ exerciseId: SQUAT.id, reps: 8, weight: 100, date: TODAY, side: 'left' });
    await api.logSet({ exerciseId: SQUAT.id, reps: 7, weight: 100, date: TODAY, side: 'left' });
    await openExercise(page, SQUAT);
    // Logged left twice by mistake: the second is set 2's left; move it to set 1's right.
    await half(page, 2, 'left').click();
    await expectReps(page, 7); // the slider shows the tapped side, not a drifted value
    await page.getByRole('region', { name: 'Side' }).getByText('Right', { exact: true }).click();
    const patched = waitForApi(page, 'PATCH /workout/sets/:id');
    await page.getByRole('button', { name: 'Save Set 2 · Right', exact: true }).click();
    await patched;
    await expect(half(page, 1, 'right')).toContainText('7 reps');
    await expect(pairCard(page, 2)).toHaveCount(0);

    await half(page, 1, 'right').click();
    await expectReps(page, 7);
    const deleted = waitForApi(page, 'DELETE /workout/sets/:id');
    await page.getByRole('button', { name: 'Delete right side' }).click();
    await deleted;
    await expect(half(page, 1, 'right')).toHaveCount(0);
    expect(await sides(db)).toEqual([{ side: 'left', weight: 100, reps: 8 }]);
  });

  test('delete: removing one side of set 1 leaves set 2 paired as it was', async ({ page, db }) => {
    await makeUnilateral(db);
    await openExercise(page, SQUAT);
    await setReps(page, 8);
    let saved = waitForApi(page, SETS);
    await logFirst(page, 'Left');
    await saved;
    for (const [n, side] of [[1, 'Right'], [2, 'Left'], [2, 'Right']] as const) {
      saved = waitForApi(page, SETS);
      await logSide(page, n, side).click();
      await saved;
    }
    // Both sides of a set number share the first side's id.
    const { rows } = await db.query(`SELECT client_id::text AS cid, pair_id::text AS pair FROM logged_sets ORDER BY logged_sets.id`);
    expect(rows.map((r: any) => r.pair)).toEqual([rows[0].cid, rows[0].cid, rows[2].cid, rows[2].cid]);

    await half(page, 1, 'left').click();
    const deleted = waitForApi(page, 'DELETE /workout/sets/:id');
    await page.getByRole('button', { name: 'Delete left side' }).click();
    await deleted;
    await expect(half(page, 1, 'left')).toHaveCount(0);
    await expect(half(page, 1, 'right')).toContainText('8 reps');
    await expect(half(page, 2, 'left')).toContainText('8 reps');
    await expect(half(page, 2, 'right')).toContainText('8 reps');
    await expect(logSide(page, 3, 'Left')).toBeVisible();
    await page.reload();
    await expect(half(page, 2, 'left')).toContainText('8 reps');
    await expect(half(page, 2, 'right')).toContainText('8 reps');
  });

  test('delete: a bilateral set among paired cards can be deleted from edit mode', async ({ page, api, db }) => {
    await makeUnilateral(db);
    await api.logSet({ exerciseId: SQUAT.id, reps: 8, weight: 100, date: TODAY });
    await api.logSet({ exerciseId: SQUAT.id, reps: 8, weight: 100, date: TODAY, side: 'left' });
    await openExercise(page, SQUAT);
    await half(page, 1, 'left').click();
    const deleted = waitForApi(page, 'DELETE /workout/sets/:id');
    await page.getByRole('button', { name: 'Delete set 1', exact: true }).click();
    await deleted;
    expect(await sides(db)).toEqual([{ side: 'left', weight: 100, reps: 8 }]);
  });

  test('history: the first side defaults to last session\'s, compared side by side', async ({ page, api, db }) => {
    await makeUnilateral(db);
    await api.logSet({ exerciseId: SQUAT.id, reps: 10, weight: 100, date: LAST_WEEK, side: 'right' });
    await api.logSet({ exerciseId: SQUAT.id, reps: 9, weight: 100, date: LAST_WEEK, side: 'left' });
    await openExercise(page, SQUAT);
    await expect(page.getByText('Last 100 lb × L9 R10')).toBeVisible();
    await page.getByRole('button', { name: 'Log Set 1', exact: true }).click();
    const dialog = page.getByRole('dialog', { name: 'Which side was that?' });
    await expect(dialog.getByRole('radio', { name: 'Right' }).or(dialog.getByRole('button', { name: 'Right' }))).toHaveAttribute('data-state', 'on');
  });

  test('bilateral history: an old set counts for both sides', async ({ page, api, db }) => {
    await makeUnilateral(db);
    await api.logSet({ exerciseId: SQUAT.id, reps: 10, weight: 100, date: LAST_WEEK });
    await api.logSet({ exerciseId: SQUAT.id, reps: 9, weight: 100, date: TODAY, side: 'left' });
    await api.logSet({ exerciseId: SQUAT.id, reps: 10, weight: 100, date: TODAY, side: 'right' });
    await openExercise(page, SQUAT);
    await expect(half(page, 1, 'left')).toContainText('Last 10');
    await expect(half(page, 1, 'left')).toContainText('−1');
    await expect(half(page, 1, 'right')).toContainText('=');
    // One old set planned one pair, and that pair is done.
    await expect(page.getByRole('button', { name: `Finish ${SQUAT.name}` })).toBeVisible();
  });

  test('uneven weights: each side keeps its own weight for the next set', async ({ page, api, db }) => {
    await makeUnilateral(db);
    await api.logSet({ exerciseId: SQUAT.id, reps: 8, weight: 100, date: TODAY, side: 'left' });
    await api.logSet({ exerciseId: SQUAT.id, reps: 8, weight: 90, date: TODAY, side: 'right' });
    await openExercise(page, SQUAT);
    await expect(half(page, 1, 'left')).toContainText('100 lb × 8');
    await expect(half(page, 1, 'right')).toContainText('90 lb × 8');
    await expect(logSide(page, 2, 'Left')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Weight 100 pounds. Change weight and setup' })).toBeVisible();

    const saved = waitForApi(page, SETS);
    await logSide(page, 2, 'Left').click();
    await saved;
    await expect(logSide(page, 2, 'Right')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Weight 90 pounds. Change weight and setup' })).toBeVisible();
    expect((await sides(db)).at(-1)).toEqual({ side: 'left', weight: 100, reps: 8 });
  });

});
