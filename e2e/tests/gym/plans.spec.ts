import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// /gym/plans (the week schedule + routine library) and a routine's page (/gym/plans/:programId/days/:id).
// The risk here is losing a routine or its exercises, so every delete checks that only the clicked
// thing went away. Seed: program 1 "Push Pull Legs" (active): Push id 1 (Mon, exercises 1-4), Pull id 2
// (Wed, 5-8), Legs id 3 (Thu, 9-12: Back Squat, Romanian Deadlift, Leg Press, Leg Curl). The epoch is
// Thursday 2026-01-15; weeks start on Sunday.
//
// Desktop (lg) shows the week as seven day cards ("Friday: rest day. Choose routine"); mobile shows a
// list with one row per day ("Fri Rest"). Both open the same day picker.

const PPL = 1;
const PUSH = 1;
const LEGS = 3;

const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

/** The schedule's control for one weekday (0 = Sunday), whichever layout this project renders. */
function dayCell(page: Page, isMobile: boolean, dow: number) {
  const name = DAYS[dow];
  return isMobile
    ? page.getByRole('button', { name: new RegExp(`^${name.slice(0, 3)}\\b`) })
    : page.getByRole('button', { name: new RegExp(`^${name}: `) });
}

/** What the schedule shows on a weekday: "Rest" or the routine name. */
async function expectDay(page: Page, isMobile: boolean, dow: number, routine: string | null) {
  const cell = dayCell(page, isMobile, dow);
  if (isMobile) await expect(cell).toHaveAccessibleName(new RegExp(`^${DAYS[dow].slice(0, 3)} ${routine ?? 'Rest'}\\b`));
  else await expect(cell).toHaveAccessibleName(`${DAYS[dow]}: ${routine ?? 'rest day'}. Choose routine`);
}

/** The /gym home week strip's link for a weekday of the epoch's week. */
const stripDay = (page: Page, dow: number) =>
  page.getByRole('list', { name: 'Days this week' }).getByRole('link', { name: new RegExp(`^${DAYS[dow]},`) });

async function openSchedule(page: Page) {
  await page.goto('/gym/plans');
  await expect(page.getByRole('combobox', { name: 'Program' })).toBeVisible();
}

async function openRoutine(page: Page, id: number, name: string, programId = PPL) {
  await page.goto(`/gym/plans/${programId}/days/${id}`);
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible();
}

const lifts = (page: Page) => page.getByRole('region', { name: /^Lifts/ });
/** Exercise names on the routine page, in order (row text is "<n> <name>"). */
async function expectLifts(page: Page, names: string[]) {
  await expect(lifts(page).getByRole('listitem')).toHaveText(names.map((n, i) => new RegExp(`^\\s*${i + 1}\\s*${n}\\s*$`)));
}

async function rowAction(page: Page, exercise: string, action: string) {
  await page.getByRole('button', { name: `${exercise} actions` }).click();
  await page.getByRole('menuitem', { name: action }).click();
}

async function count(db: import('pg').Client, sql: string, params: unknown[] = []): Promise<number> {
  const { rows } = await db.query(`SELECT count(*)::int AS n FROM ${sql}`, params);
  return rows[0].n;
}

test.describe('gym plans: programs', { tag: ['@gym'] }, () => {
  test('create a program, schedule a routine, make it active: /gym follows it', async ({ page, isMobile, db }) => {
    await openSchedule(page);
    await page.getByRole('button', { name: 'Program actions' }).click();
    await page.getByRole('menuitem', { name: 'New program' }).click();
    const create = page.getByRole('dialog', { name: 'New program' });
    await create.getByRole('textbox', { name: 'Program name' }).fill('Upper Lower');
    const created = waitForApi(page, 'POST /workout/programs');
    await create.getByRole('button', { name: 'Create' }).click();
    await created;

    // The new program is shown, not active (Push Pull Legs still is), with an empty week.
    await expect(page.getByRole('combobox', { name: 'Program' })).toHaveValue(/\d+/);
    await expect(page.getByRole('combobox', { name: 'Program' }).locator('option:checked')).toHaveText('Upper Lower');
    await expect(page.getByRole('button', { name: 'Make active' })).toBeVisible();
    await expect(page.getByText('No routines in this program')).toBeVisible();

    // A routine for Friday, straight from the day picker.
    await dayCell(page, isMobile, 5).click();
    await page.getByRole('dialog', { name: 'Friday' }).getByRole('button', { name: 'New routine for Friday' }).click();
    const routine = page.getByRole('dialog', { name: 'New routine for Friday' });
    await routine.getByRole('textbox', { name: 'Routine name' }).fill('Upper');
    const planned = waitForApi(page, 'POST /workout/programs/:id/plans');
    await routine.getByRole('button', { name: 'Create' }).click();
    await planned;
    await expectDay(page, isMobile, 5, 'Upper');

    const activated = waitForApi(page, 'POST /workout/programs/:id/activate');
    await page.getByRole('button', { name: 'Make active' }).click();
    await activated;
    await expect(page.getByText('Active', { exact: true })).toBeVisible();

    await page.reload();
    // The active program with routines is the one shown first.
    await expect(page.getByRole('combobox', { name: 'Program' }).locator('option:checked')).toHaveText('Upper Lower');
    await expect(page.getByText('Active', { exact: true })).toBeVisible();
    await expectDay(page, isMobile, 5, 'Upper');
    await expectDay(page, isMobile, 1, null);
    // The old program is no longer active, and still has its routines.
    await page.getByRole('combobox', { name: 'Program' }).selectOption({ label: 'Push Pull Legs' });
    await expect(page.getByRole('button', { name: 'Make active' })).toBeVisible();
    await expectDay(page, isMobile, 1, 'Push');

    const { rows } = await db.query(`SELECT name FROM workout_programs WHERE is_active ORDER BY id`);
    expect(rows.map((r) => r.name)).toEqual(['Upper Lower']);

    // /gym's week strip follows the new active program.
    await page.goto('/gym');
    await expect(stripDay(page, 5)).toHaveAccessibleName(/^Friday, January 16, Upper\b/);
    await expect(stripDay(page, 1)).toHaveAccessibleName(/^Monday, January 12, rest day/);
    await expect(stripDay(page, 3)).toHaveAccessibleName(/^Wednesday, January 14, rest day/);
  });

  test('empty state: with no programs, "Create a program" makes an active one', async ({ page, db }) => {
    // No API deletes programs, so the state is arranged in the DB (README principle 7).
    for (const t of ['workout_plan_exercises', 'workout_plan_days', 'workout_plans', 'workout_programs']) await db.query(`DELETE FROM ${t}`);
    await page.goto('/gym/plans');
    await expect(page.getByText('No programs yet', { exact: true })).toBeVisible();
    await page.getByRole('button', { name: 'Create a program' }).click();
    const dialog = page.getByRole('dialog', { name: 'New program' });
    await dialog.getByRole('textbox', { name: 'Program name' }).fill('Full Body');
    const created = waitForApi(page, 'POST /workout/programs');
    await dialog.getByRole('button', { name: 'Create' }).click();
    await created;
    await expect(page.getByRole('combobox', { name: 'Program' }).locator('option:checked')).toHaveText('Full Body');
    await page.reload();
    await expect(page.getByRole('combobox', { name: 'Program' }).locator('option:checked')).toHaveText('Full Body');
    await expect(page.getByText('Active', { exact: true })).toBeVisible(); // the first program is made active
    await expect(page.getByText('No routines in this program')).toBeVisible();
  });

  test('load error: "Couldn\'t load your programs", and Retry recovers', async ({ page, e2e }) => {
    await e2e.fault({ table: 'workout_programs', op: 'query', times: 3 }); // the first try and both auto-retries
    const failed = () =>
      page.waitForResponse((r) => r.url().includes('/api/workout/programs') && r.status() === 500);
    const failures = (async () => {
      for (let i = 0; i < 3; i++) {
        await failed();
        await page.clock.runFor(5_000); // skip the retry backoff
      }
    })();
    await page.goto('/gym/plans');
    await failures;
    await expect(page.getByRole('alert')).toContainText("Couldn't load your programs");
    await page.getByRole('button', { name: 'Retry' }).click();
    await expect(page.getByRole('combobox', { name: 'Program' }).locator('option:checked')).toHaveText('Push Pull Legs');
    await expect(page.getByRole('alert')).toHaveCount(0);
  });
});

test.describe('gym plans: routines', { tag: ['@gym'] }, () => {
  test('create a routine: it joins the library and persists', async ({ page, db }) => {
    await openSchedule(page);
    await page.getByRole('button', { name: 'New routine' }).click();
    const dialog = page.getByRole('dialog', { name: 'New routine' });
    await dialog.getByRole('textbox', { name: 'Routine name' }).fill('Arms');
    const created = waitForApi(page, 'POST /workout/programs/:id/plans');
    await dialog.getByRole('button', { name: 'Create' }).click();
    await created;
    await expect(page.getByRole('button', { name: 'Open Arms' })).toBeVisible();
    await page.reload();
    await page.getByRole('button', { name: 'Open Arms' }).click();
    await expect(page.getByRole('heading', { level: 1, name: 'Arms' })).toBeVisible();
    await expect(page.getByText('No days yet')).toBeVisible();
    expect(await count(db, `workout_plans WHERE deleted_at IS NULL`)).toBe(4);
  });

  test('rename a routine: the new name persists, exercises and days stay', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    await page.getByRole('button', { name: 'Routine actions' }).click();
    await page.getByRole('menuitem', { name: 'Rename' }).click();
    const dialog = page.getByRole('dialog', { name: 'Rename routine' });
    await expect(dialog.getByRole('textbox', { name: 'Routine name' })).toHaveValue('Legs');
    await dialog.getByRole('textbox', { name: 'Routine name' }).fill('Lower');
    const renamed = waitForApi(page, 'PATCH /workout/plans/:id');
    await dialog.getByRole('button', { name: 'Rename' }).click();
    await expect(page.getByRole('heading', { level: 1, name: 'Lower' })).toBeVisible();
    await renamed;
    await page.reload();
    await expect(page.getByRole('heading', { level: 1, name: 'Lower' })).toBeVisible();
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl']);
    await expect(page.getByRole('button', { name: 'Thursday', pressed: true })).toBeVisible();
  });

  test('delete a routine: only that routine leaves the schedule', async ({ page, isMobile, db }) => {
    await openRoutine(page, LEGS, 'Legs');
    await page.getByRole('button', { name: 'Routine actions' }).click();
    const deleted = waitForApi(page, 'DELETE /workout/plans/:id');
    await page.getByRole('menuitem', { name: 'Delete routine' }).click();
    await deleted;
    await expect(page).toHaveURL(/\/gym\/plans\?program=1$/);
    await expect(page.getByText('Deleted Legs', { exact: true })).toBeVisible();
    await page.reload();
    await expect(page.getByRole('button', { name: 'Open Legs' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Open Push' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Open Pull' })).toBeVisible();
    await expectDay(page, isMobile, 4, null);
    await expectDay(page, isMobile, 1, 'Push');
    await expectDay(page, isMobile, 3, 'Pull');
    // Soft delete: Legs is marked deleted, the others and every routine's exercises are untouched.
    const { rows } = await db.query(`SELECT name FROM workout_plans WHERE deleted_at IS NOT NULL`);
    expect(rows.map((r) => r.name)).toEqual(['Legs']);
    expect(await count(db, `workout_plans WHERE deleted_at IS NULL`)).toBe(2);
    expect(await count(db, `workout_plan_exercises`)).toBe(12);
    expect(await count(db, `exercises`)).toBe(12);
  });

  test('delete a routine, then Undo: it comes back with its day and exercises', async ({ page, isMobile }) => {
    await openRoutine(page, LEGS, 'Legs');
    await page.getByRole('button', { name: 'Routine actions' }).click();
    const deleted = waitForApi(page, 'DELETE /workout/plans/:id');
    await page.getByRole('menuitem', { name: 'Delete routine' }).click();
    await deleted;
    const restored = waitForApi(page, 'POST /workout/plans/:id/restore');
    await page.getByRole('button', { name: 'Undo', exact: true }).click();
    await restored;
    await expect(page.getByRole('heading', { level: 1, name: 'Legs' })).toBeVisible();
    await page.reload();
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl']);
    await expect(page.getByRole('button', { name: 'Thursday', pressed: true })).toBeVisible();
    await openSchedule(page);
    await expectDay(page, isMobile, 4, 'Legs');
  });
});

test.describe('gym plans: routine exercises', { tag: ['@gym'] }, () => {
  test('add an existing exercise: it goes to the end and persists', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    await page.getByRole('button', { name: 'Add exercise' }).click();
    const dialog = page.getByRole('dialog', { name: 'Add to Legs' });
    await dialog.getByRole('textbox', { name: 'Search' }).fill('Face');
    const added = waitForApi(page, 'POST /workout/plans/:id/exercises/add');
    await dialog.getByRole('button', { name: 'Face Pull' }).click();
    await expect(dialog).toHaveCount(0);
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl', 'Face Pull']);
    await added;
    await page.reload();
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl', 'Face Pull']);
  });

  test('"Create and add": a new exercise is created and added to the routine', async ({ page, db }) => {
    await openRoutine(page, LEGS, 'Legs');
    await page.getByRole('button', { name: 'Add exercise' }).click();
    const dialog = page.getByRole('dialog', { name: 'Add to Legs' });
    await dialog.getByRole('textbox', { name: 'Search' }).fill('Hip Thrust');
    const added = waitForApi(page, 'POST /workout/plans/:id/exercises/add');
    await dialog.getByRole('button', { name: 'Create and add "Hip Thrust"' }).click();
    await added;
    await expect(dialog).toHaveCount(0);
    await page.reload();
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl', 'Hip Thrust']);
    expect(await count(db, `exercises WHERE name = 'Hip Thrust'`)).toBe(1);
  });

  test('remove an exercise: only that one goes, the rest keep their order', async ({ page, db }) => {
    await openRoutine(page, LEGS, 'Legs');
    const removed = waitForApi(page, 'DELETE /workout/plans/:id/exercises/remove');
    await rowAction(page, 'Romanian Deadlift', 'Remove from routine');
    await expectLifts(page, ['Back Squat', 'Leg Press', 'Leg Curl']);
    await removed;
    await page.reload();
    await expectLifts(page, ['Back Squat', 'Leg Press', 'Leg Curl']);
    // Only the routine link went: the exercise stays in the library, other routines are untouched.
    expect(await count(db, `workout_plan_exercises WHERE workout_plan_id = $1`, [LEGS])).toBe(3);
    expect(await count(db, `workout_plan_exercises`)).toBe(11);
    expect(await count(db, `exercises WHERE name = 'Romanian Deadlift'`)).toBe(1);
  });

  test('remove an exercise, then Undo: it goes back in its place', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    const removed = waitForApi(page, 'DELETE /workout/plans/:id/exercises/remove');
    await rowAction(page, 'Romanian Deadlift', 'Remove from routine');
    await removed;
    // Undo re-adds (to the end) and then reorders back to the old position.
    const reordered = waitForApi(page, 'PUT /workout/plans/:id/exercises/reorder');
    await page.getByRole('button', { name: 'Undo', exact: true }).click();
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl']);
    await reordered;
    await page.reload();
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl']);
  });

  test('edit an exercise: the rep rollover persists', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    await rowAction(page, 'Leg Press', 'Edit exercise');
    const dialog = page.getByRole('dialog', { name: 'Edit Leg Press' });
    const rollover = dialog.getByLabel('Rep rollover', { exact: true });
    await expect(rollover).toHaveValue('12');
    await rollover.fill('15');
    const saved = waitForApi(page, 'PUT /workout/exercises/:id');
    await dialog.getByRole('button', { name: 'Save' }).click();
    await saved;
    await expect(dialog).toHaveCount(0);
    await page.reload();
    await rowAction(page, 'Leg Press', 'Edit exercise');
    await expect(page.getByRole('dialog', { name: 'Edit Leg Press' }).getByLabel('Rep rollover', { exact: true })).toHaveValue('15');
  });
});

test.describe('gym plans: machine settings', { tag: ['@gym'] }, () => {
  test('edit an exercise: your own machine settings persist, and can be cleared', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    await rowAction(page, 'Leg Press', 'Edit exercise');
    let dialog = page.getByRole('dialog', { name: 'Edit Leg Press' });
    await dialog.getByRole('button', { name: 'Add setting' }).click();
    await dialog.getByLabel('Setting label').fill('Seat');
    await dialog.getByLabel('Setting value').fill('4');
    const saved = waitForApi(page, 'PUT /workout/exercises/:id/settings');
    await dialog.getByRole('button', { name: 'Save' }).click();
    await saved;
    await expect(dialog).toHaveCount(0);
    await page.reload();
    await rowAction(page, 'Leg Press', 'Edit exercise');
    dialog = page.getByRole('dialog', { name: 'Edit Leg Press' });
    await expect(dialog.getByLabel('Setting label')).toHaveValue('Seat');
    await expect(dialog.getByLabel('Setting value')).toHaveValue('4');
    await dialog.getByRole('button', { name: 'Remove setting' }).click();
    await dialog.getByRole('button', { name: 'Save' }).click();
    await expect(dialog).toHaveCount(0);
    await page.reload();
    await rowAction(page, 'Leg Press', 'Edit exercise');
    await expect(page.getByRole('dialog', { name: 'Edit Leg Press' }).getByLabel('Setting label')).toHaveCount(0);
  });
});

test.describe('gym plans: scheduling', { tag: ['@gym'] }, () => {
  test('day picker: put a routine on a free day; the schedule and /gym show it', async ({ page, isMobile }) => {
    await openSchedule(page);
    await dayCell(page, isMobile, 5).click();
    const picker = page.getByRole('dialog', { name: 'Friday' });
    await expect(picker.getByRole('button', { name: /^Pull\b/ })).toContainText('also Wed');
    const assigned = waitForApi(page, 'POST /workout/plans/:id/assign-day');
    await picker.getByRole('button', { name: /^Pull\b/ }).click();
    await expectDay(page, isMobile, 5, 'Pull');
    await assigned;
    await page.reload();
    await expectDay(page, isMobile, 5, 'Pull');
    await expectDay(page, isMobile, 3, 'Pull'); // a routine can be on several days
    await page.goto('/gym');
    await expect(stripDay(page, 5)).toHaveAccessibleName(/^Friday, January 16, Pull\b/);
    await expect(stripDay(page, 3)).toHaveAccessibleName(/^Wednesday, January 14, Pull\b/);
  });

  test('day picker: "Rest day" clears a day and keeps the routine', async ({ page, isMobile, db }) => {
    await openSchedule(page);
    await dayCell(page, isMobile, 1).click();
    const cleared = waitForApi(page, 'DELETE /workout/plans/:id/assign-day');
    await page.getByRole('dialog', { name: 'Monday' }).getByRole('button', { name: /^Rest day/ }).click();
    await expectDay(page, isMobile, 1, null);
    await cleared;
    await page.reload();
    await expectDay(page, isMobile, 1, null);
    await expect(page.getByRole('button', { name: 'Open Push' })).toBeVisible();
    expect(await count(db, `workout_plan_exercises WHERE workout_plan_id = $1`, [PUSH])).toBe(4);
  });

  test('routine page: a day taken by another routine moves after confirming', async ({ page, isMobile }) => {
    await openRoutine(page, LEGS, 'Legs');
    // Monday is Push's: the toggle says so, and taking it asks first.
    const monday = page.getByRole('button', { name: 'Monday (Push)' });
    await expect(monday).toHaveAttribute('aria-pressed', 'false');
    await monday.click();
    const confirm = page.getByRole('dialog', { name: 'Move Monday to Legs?' });
    await expect(confirm).toContainText('Monday is Push now. Push keeps its other days.');
    const assigned = waitForApi(page, 'POST /workout/plans/:id/assign-day');
    await confirm.getByRole('button', { name: 'Move to Legs' }).click();
    await expect(page.getByRole('button', { name: 'Monday', pressed: true })).toBeVisible();
    await assigned;
    await page.reload();
    await expect(page.getByRole('button', { name: 'Monday', pressed: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Thursday', pressed: true })).toBeVisible();

    await openSchedule(page);
    await expectDay(page, isMobile, 1, 'Legs');
    await expectDay(page, isMobile, 4, 'Legs');
    await expect(page.getByRole('button', { name: 'Open Push' })).toBeVisible(); // unscheduled, not deleted
    await page.goto('/gym');
    await expect(stripDay(page, 1)).toHaveAccessibleName(/^Monday, January 12, Legs\b/);
  });

  test('routine page: cancelling the move leaves both routines as they were', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    await page.getByRole('button', { name: 'Monday (Push)' }).click();
    await page.getByRole('dialog', { name: 'Move Monday to Legs?' }).getByRole('button', { name: 'Cancel' }).click();
    await page.reload();
    await expect(page.getByRole('button', { name: 'Monday (Push)' })).toHaveAttribute('aria-pressed', 'false');
  });

  test('drag a routine from the library onto a day (desktop)', async ({ page, isMobile }) => {
    test.skip(isMobile, 'drag and drop is desktop only (the day columns render at lg)');
    await openSchedule(page);
    const assigned = waitForApi(page, 'POST /workout/plans/:id/assign-day');
    await page.getByRole('button', { name: 'Open Legs' }).dragTo(dayCell(page, false, 6));
    await expectDay(page, false, 6, 'Legs');
    await assigned;
    await page.reload();
    await expectDay(page, false, 6, 'Legs');
    await expectDay(page, false, 4, 'Legs');
  });

  test('drag a day onto another day moves the routine (desktop)', async ({ page, isMobile }) => {
    test.skip(isMobile, 'drag and drop is desktop only (the day columns render at lg)');
    await openSchedule(page);
    const assigned = waitForApi(page, 'POST /workout/plans/:id/assign-day');
    const cleared = waitForApi(page, 'DELETE /workout/plans/:id/assign-day');
    await dayCell(page, false, 3).dragTo(dayCell(page, false, 2));
    await assigned;
    await cleared;
    await page.reload();
    await expectDay(page, false, 2, 'Pull');
    await expectDay(page, false, 3, null);
  });
});

test.describe('gym plans: warmup, cardio, stretch', { tag: ['@gym'] }, () => {
  const MOBILITY = 'PUT /workout/plans/:id/planned-mobility';

  test('warmup and stretch items: add and remove persist', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    const warmup = page.getByRole('region', { name: /^Warmup/ });
    const stretch = page.getByRole('region', { name: /^Stretch/ });
    for (const item of ['Cat cow', 'Leg swings']) {
      const saved = waitForApi(page, MOBILITY);
      await warmup.getByRole('textbox', { name: 'Add to warmup' }).fill(item);
      await warmup.getByRole('button', { name: 'Add' }).click();
      await saved;
    }
    let saved = waitForApi(page, MOBILITY);
    await stretch.getByRole('textbox', { name: 'Add to stretch' }).fill('Pigeon');
    await stretch.getByRole('textbox', { name: 'Add to stretch' }).press('Enter');
    await saved;
    await page.reload();
    await expect(warmup.getByRole('listitem')).toHaveText(['Cat cow', 'Leg swings'].map((t) => new RegExp(t)));
    await expect(stretch.getByRole('listitem')).toHaveText([/Pigeon/]);

    saved = waitForApi(page, MOBILITY);
    await warmup.getByRole('button', { name: 'Remove Cat cow' }).click();
    await expect(warmup.getByRole('listitem')).toHaveText([/Leg swings/]);
    await saved;
    await page.reload();
    await expect(warmup.getByRole('listitem')).toHaveText([/Leg swings/]);
    await expect(stretch.getByRole('listitem')).toHaveText([/Pigeon/]); // the other list is untouched
    await expectLifts(page, ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl']);
  });

  test('cardio: type and minutes save when focus leaves the fields', async ({ page }) => {
    await openRoutine(page, LEGS, 'Legs');
    const cardio = page.getByRole('region', { name: /^Cardio/ });
    await cardio.getByRole('textbox', { name: 'Cardio type' }).fill('Bike');
    // The minutes input can't be typed in (0px wide, see the next test), so use the stepper: from empty
    // the first + gives the minimum (0), then each one adds 5.
    for (let i = 0; i < 4; i++) await cardio.getByRole('button', { name: 'Increase cardio minutes' }).click();
    await expect(cardio.getByRole('spinbutton')).toHaveValue('15');
    const saved = waitForApi(page, 'PUT /workout/plans/:id/planned-cardio');
    await page.getByRole('heading', { level: 1, name: 'Legs' }).click(); // move focus out of the form
    await saved;
    await page.reload();
    await expect(cardio.getByRole('textbox', { name: 'Cardio type' })).toHaveValue('Bike');
    await expect(cardio.getByRole('spinbutton')).toHaveValue('15');
  });

  // Bug (e2e/notes/gym-plans.md): the minutes box is w-36; the two 44px steppers, padding and the
  // "min" unit leave the input itself 0px wide, so minutes can only be changed with +/-.
  test('cardio minutes input is wide enough to type in', async ({ page }) => {
    test.fail(true, 'known bug: the cardio minutes input renders 0px wide (notes/gym-plans.md)');
    await openRoutine(page, LEGS, 'Legs');
    await expect(page.getByRole('region', { name: /^Cardio/ }).getByRole('spinbutton')).toBeVisible({ timeout: 2_000 });
  });
});
