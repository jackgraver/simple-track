import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import { LAST_WEEK, LEGS, SETS, TODAY, logButton, openExercise, setReps, setRow, setsList } from './logger';

// The workout session: the session list (/gym/logging) and the per-exercise logger's session flow
// (Up next, Finish, weight/plates setup, notes, save failures, offline). Today is "Legs" (exercises 9-12).

const sessionList = (page: Page) => page.getByRole('region', { name: 'Session' });
/** A step row on the session list; its name reads "<name>[ , done| , up next] <detail>". */
const stepRow = (page: Page, name: string) => sessionList(page).getByRole('button', { name: new RegExp(`^${name}\\b`) });
const setupCard = (page: Page) => page.getByRole('button', { name: /Change weight and setup$/ });

/** A reload that starts from the server's state only: the logger keeps its draft in sessionStorage. */
async function freshReload(page: Page) {
  await page.evaluate(() => sessionStorage.clear());
  await page.reload();
}

test.describe('gym session', { tag: ['@gym'] }, () => {
  test('session list: today\'s Legs exercises with their done state; a tap opens the logger', async ({ page, api }) => {
    await api.logSet({ exerciseId: LEGS.squat.id, reps: 8, weight: 100, date: TODAY });
    await page.goto('/gym/logging');
    await expect(page.getByRole('heading', { name: 'Legs', level: 1 })).toBeVisible();
    await expect(page.getByText('1 / 4 done')).toBeVisible();

    await expect(stepRow(page, 'Back Squat')).toHaveAccessibleName(/^Back Squat , done 100 lb × 8$/);
    for (const ex of [LEGS.rdl, LEGS.legPress, LEGS.legCurl]) {
      await expect(stepRow(page, ex.name)).toBeVisible();
      await expect(stepRow(page, ex.name)).not.toHaveAccessibleName(/, done/);
    }
    await expect(stepRow(page, LEGS.rdl.name)).toHaveAccessibleName(/, up next/);
    await expect(page.getByRole('heading', { name: 'Romanian Deadlift', level: 2 })).toBeVisible(); // the Up next card

    await stepRow(page, LEGS.legPress.name).click();
    await expect(page).toHaveURL(/\/gym\/logging\/11$/);
    await expect(page.getByRole('heading', { name: 'Leg Press' })).toBeVisible();
    await expect(logButton(page, 1)).toBeVisible();
  });

  test('up next: finishing an exercise suggests the next one, and opening it works', async ({ page, api }) => {
    // One set last week: today's plan is one set at 100 x 8, so the logger opens ready to log it.
    await api.logSet({ exerciseId: LEGS.squat.id, reps: 8, weight: 100, date: LAST_WEEK });
    await openExercise(page, LEGS.squat);
    const saved = waitForApi(page, SETS);
    await logButton(page, 1).click();
    await expect(setRow(page, 1)).toContainText('100 × 8');
    await saved;

    await page.getByRole('button', { name: 'Finish Back Squat' }).click();
    await expect(page.getByRole('heading', { name: 'Up next' })).toBeVisible();
    await expect(page.getByLabel('Finished exercise')).toContainText('100 lb × 8');
    const upNext = page.getByRole('region', { name: 'Up next' });
    await expect(upNext).toContainText('Romanian Deadlift');
    await expect(page.getByRole('region', { name: 'Also left' })).toContainText('Leg Press');

    await upNext.getByRole('button').click();
    await expect(page).toHaveURL(/\/gym\/logging\/10$/);
    await expect(page.getByRole('heading', { name: 'Romanian Deadlift' })).toBeVisible();

    await page.goto('/gym/logging');
    await expect(stepRow(page, 'Back Squat')).toHaveAccessibleName(/, done 100 lb × 8$/);
  });

  test('finish workout: the last exercise completes the session and the summary adds up', async ({ page, api }) => {
    await api.logSet({ exerciseId: LEGS.squat.id, reps: 8, weight: 100, date: TODAY });
    await api.logSet({ exerciseId: LEGS.squat.id, reps: 8, weight: 100, date: TODAY });
    await api.logSet({ exerciseId: LEGS.rdl.id, reps: 10, weight: 80, date: TODAY });
    await api.logSet({ exerciseId: LEGS.legPress.id, reps: 12, weight: 200, date: TODAY });
    await api.logSet({ exerciseId: LEGS.legCurl.id, reps: 10, weight: 50, date: LAST_WEEK });

    await openExercise(page, LEGS.legCurl);
    const saved = waitForApi(page, SETS);
    await logButton(page, 1).click(); // opens at last week's 50 x 10
    await expect(setRow(page, 1)).toContainText('50 × 10');
    await saved;
    await page.getByRole('button', { name: 'Finish Leg Curl' }).click();
    await expect(page.getByText("Everything's logged")).toBeVisible();
    await page.getByRole('button', { name: 'Back to session' }).click();

    // 100x8 + 100x8 + 80x10 + 200x12 + 50x10 = 5,300 lb over 5 sets and 48 reps.
    const expectComplete = async () => {
      await expect(page.getByRole('heading', { name: 'Workout complete' })).toBeVisible();
      await expect(page.getByText('All 4 steps logged')).toBeVisible();
      const summary = page.getByRole('region', { name: 'Session summary' });
      await expect(summary.getByRole('definition').first()).toHaveText(/^5,300\s*lb$/);
      await expect(summary).toContainText(/Sets\s*5\s*48 reps/);
      await expect(summary.getByRole('listitem')).toHaveCount(4);
      await expect(summary.getByRole('listitem').filter({ hasText: 'Leg Curl' })).toContainText('Top set 50 lb × 10');
    };
    await expectComplete();
    await page.reload();
    await expectComplete();

    await page.getByRole('button', { name: 'Finish workout' }).click();
    await expect(page).toHaveURL(/\/gym(\?.*)?$/);
  });

  test('weight + plates: the setup is logged with the set and persists after a reload', async ({ page, db }) => {
    await openExercise(page, LEGS.squat);
    await setupCard(page).click();
    const sheet = page.getByRole('dialog', { name: 'Weight & setup · Back Squat' });
    await sheet.getByRole('button', { name: /Not set\s*Edit/ }).click();

    const plates = page.getByRole('dialog', { name: 'Plates' });
    await plates.getByRole('button', { name: 'Increase 45 lb plates' }).click();
    await plates.getByRole('button', { name: 'Increase 45 lb plates' }).click();
    await expect(plates.getByText('225 lbs', { exact: true })).toBeVisible(); // 45 lb bar + 2 x 45 per side
    await plates.getByRole('button', { name: 'Save & set weight to 225 lbs' }).click();
    await expect(plates).toBeHidden();
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 225 pounds/);
    const plateText = (await setupCard(page).getByText(/45/).first().textContent())!.trim();
    expect(plateText).not.toBe('');

    await setReps(page, 5);
    const saved = waitForApi(page, SETS);
    await logButton(page, 1).click();
    await expect(setRow(page, 1)).toContainText('225 × 5');
    await saved;
    const { rows } = await db.query(`SELECT weight::float AS weight, weight_setup FROM logged_sets WHERE deleted_at IS NULL`);
    expect(rows).toEqual([{ weight: 225, weight_setup: plateText }]);

    await freshReload(page);
    await expect(setRow(page, 1)).toContainText('225 × 5');
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 225 pounds/);
    await expect(setupCard(page)).toContainText(plateText);
  });

  test('notes: saved notes persist after a reload', async ({ page, api, db }) => {
    await api.logSet({ exerciseId: LEGS.squat.id, reps: 8, weight: 100, date: TODAY });
    await openExercise(page, LEGS.squat);
    await page.getByRole('button', { name: 'Add notes' }).click();
    const dialog = page.getByRole('dialog', { name: 'Notes · Back Squat' });
    await dialog.getByRole('textbox', { name: 'Notes' }).fill('Knees out, belt on top set');
    const saved = waitForApi(page, 'POST /workout/exercises/log');
    await dialog.getByRole('button', { name: 'Save notes' }).click();
    await saved;
    await expect(page.getByRole('button', { name: 'Edit notes' })).toBeVisible();

    await freshReload(page);
    await expect(setRow(page, 1)).toContainText('100 × 8');
    await page.getByRole('button', { name: 'Edit notes' }).click();
    await expect(page.getByRole('dialog', { name: 'Notes · Back Squat' }).getByRole('textbox', { name: 'Notes' })).toHaveValue(
      'Knees out, belt on top set',
    );
    // Notes are saved by re-posting the whole logged exercise: it must keep its one set.
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets WHERE deleted_at IS NULL`);
    expect(rows[0].n).toBe(1);
  });

  test('notes typed before the first set are saved with it', async ({ page }) => {
    await openExercise(page, LEGS.squat);
    await page.getByRole('button', { name: 'Add notes' }).click();
    const dialog = page.getByRole('dialog', { name: 'Notes · Back Squat' });
    await dialog.getByRole('textbox', { name: 'Notes' }).fill('Felt strong');
    await dialog.getByRole('button', { name: 'Save notes' }).click();
    await setReps(page, 5);
    const notesSaved = waitForApi(page, 'POST /workout/exercises/log');
    await logButton(page, 1).click();
    await notesSaved;

    await freshReload(page);
    await expect(setRow(page, 1)).toContainText('× 5');
    await page.getByRole('button', { name: 'Edit notes' }).click();
    await expect(page.getByRole('dialog').getByRole('textbox', { name: 'Notes' })).toHaveValue('Felt strong');
  });

  test.describe('save failures', () => {
    const failedSet = (page: Page) =>
      page.waitForResponse((r) => r.request().method() === 'POST' && r.url().endsWith('/api/workout/sets') && r.status() === 500);

    test('persistent DB fault: "Didn\'t save" + Retry saves exactly one set', async ({ page, e2e, db }) => {
      await e2e.fault({ table: 'logged_sets', op: 'create', times: 3 }); // the first try and both auto-retries
      await openExercise(page, LEGS.squat);
      await setReps(page, 8);
      const failures = (async () => {
        for (let i = 0; i < 3; i++) {
          await failedSet(page);
          await page.clock.runFor(5_000); // skip the retry backoff
        }
      })();
      await logButton(page, 1).click();
      await expect(setRow(page, 1)).toContainText('× 8'); // optimistic
      await failures;
      const row = setsList(page).getByRole('listitem');
      await expect(row).toContainText("Didn't save");
      expect((await db.query(`SELECT count(*)::int AS n FROM logged_sets`)).rows[0].n).toBe(0);

      const saved = waitForApi(page, SETS);
      await page.getByRole('button', { name: 'Retry saving set 1' }).click();
      await saved;
      await expect(row).not.toContainText("Didn't save");
      await expect(page.getByRole('button', { name: 'Retry saving set 1' })).toHaveCount(0);

      await freshReload(page);
      await expect(setRow(page, 1)).toContainText('× 8');
      await expect(setRow(page, 2)).toHaveCount(0);
      const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets`);
      expect(rows[0].n).toBe(1);
      const ex = await db.query(`SELECT count(*)::int AS n FROM logged_exercises WHERE deleted_at IS NULL`);
      expect(ex.rows[0].n).toBe(1); // no ghost exercise from the failed attempts
    });

    test('offline: a set logged offline syncs once when the connection returns', async ({ page, context, db }) => {
      await openExercise(page, LEGS.squat);
      await setReps(page, 8);
      await context.setOffline(true);
      await logButton(page, 1).click();
      await expect(setRow(page, 1)).toContainText('× 8');
      await expect(setsList(page)).toContainText('Waiting for connection');
      expect((await db.query(`SELECT count(*)::int AS n FROM logged_sets`)).rows[0].n).toBe(0);

      const saved = waitForApi(page, SETS);
      await context.setOffline(false);
      await saved;
      await expect(setsList(page)).not.toContainText('Waiting for connection');

      await freshReload(page);
      await expect(setRow(page, 1)).toContainText('× 8');
      await expect(setRow(page, 2)).toHaveCount(0);
      const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets`);
      expect(rows[0].n).toBe(1);
    });

    test('delete while the set is still saving: nothing is left after the save lands', async ({ page, e2e, db }) => {
      await openExercise(page, LEGS.squat);
      await setReps(page, 8);
      const gate = await e2e.hold(SETS);
      await logButton(page, 1).click();
      await gate.arrived();
      await expect(setsList(page).getByRole('listitem')).toHaveAttribute('aria-busy', 'true');
      const deleted = waitForApi(page, 'DELETE /workout/sets/:id');
      await page.getByRole('button', { name: 'Delete set 1' }).click();
      await expect(setRow(page, 1)).toHaveCount(0);
      await gate.release(); // the create commits, then the queued delete runs
      await deleted;

      await freshReload(page);
      await expect(logButton(page, 1)).toBeVisible();
      await expect(setRow(page, 1)).toHaveCount(0);
      const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets WHERE deleted_at IS NULL`);
      expect(rows[0].n).toBe(0);
    });
  });

  test('edit the weight of a saved set: the new weight persists', async ({ page, api }) => {
    await api.logSet({ exerciseId: LEGS.squat.id, reps: 8, weight: 100, date: TODAY });
    await openExercise(page, LEGS.squat);
    await setRow(page, 1).click();
    await expect(page.getByRole('button', { name: 'Save Set 1', exact: true })).toBeVisible();
    await setupCard(page).click();
    const sheet = page.getByRole('dialog', { name: 'Weight & setup · Back Squat' });
    await sheet.getByRole('spinbutton', { name: 'Weight (lb)' }).fill('110');
    await sheet.getByRole('spinbutton', { name: 'Weight (lb)' }).press('Enter');
    await sheet.getByRole('button', { name: 'Done' }).click();
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 110 pounds/);

    const saved = waitForApi(page, 'PATCH /workout/sets/:id');
    await page.getByRole('button', { name: 'Save Set 1', exact: true }).click();
    await expect(setRow(page, 1)).toContainText('110 × 8');
    await saved;
    await freshReload(page);
    await expect(setRow(page, 1)).toContainText('110 × 8');
    await expect(setRow(page, 2)).toHaveCount(0);
  });
});
