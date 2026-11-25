import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// /diet/targets ("Macro targets"): edit the calorie target and the macro split, save, and check what
// the plan snapshot does. Saving (PUT /diet/plans/plan/:id) does NOT edit the plan row: the backend
// creates a new plan row effective today and re-points every diet day from today forward at it, so
// past days keep the plan they had (UpdatePlanMacros in backend/internal/core/diet/repository).
// Notes and findings: e2e/notes/diet-targets.md.

const SAVE = 'PUT /diet/plans/plan/:id';

// The seed's "Maintenance" plan: 2500 kcal, P180, C280, F70, fiber 35.
const calories = (page: Page) => page.getByLabel('Daily calories (kcal)');
const protein = (page: Page) => page.getByLabel('Protein (g)');
const carbs = (page: Page) => page.getByLabel('Carbs (g)');
const fat = (page: Page) => page.getByLabel('Fat (g)');
const fiber = (page: Page) => page.getByLabel('Fiber (g)');
// Desktop: header button (disabled until dirty). Mobile: bottom button, "Targets saved" until dirty.
const saveButton = (page: Page) => page.getByRole('button', { name: /^(Save targets|Targets saved)$/ });
const unsaved = (page: Page) => page.getByText('Unsaved changes');

async function openTargets(page: Page) {
  await page.goto('/diet/targets');
  await expect(page.getByRole('heading', { name: 'Macro targets' })).toBeVisible();
  await expect(calories(page)).toHaveValue('2500');
}

/**
 * Calories first: every calorie / protein / fat edit refills carbs from what is left
 * (carbs = (kcal - 4P - 9F) / 4), so 2200 kcal, P200, F60 leaves exactly C215.
 */
async function enterNewTargets(page: Page) {
  await calories(page).fill('2200');
  await protein(page).fill('200');
  await fat(page).fill('60');
  await fiber(page).fill('30');
  await expect(carbs(page)).toHaveValue('215');
}

test.describe('diet targets: edit and save', { tag: ['@diet'] }, () => {
  test('change calories and split, save: persists after reload, history gets a new Current snapshot', async ({ page, api }) => {
    await openTargets(page);
    await expect(page.getByText('Maintenance · since')).toBeVisible();
    await enterNewTargets(page);
    await expect(unsaved(page)).toBeVisible();

    const saved = waitForApi(page, SAVE);
    await saveButton(page).click();
    await saved;
    await expect(page.getByText('Targets saved', { exact: true }).first()).toBeVisible();
    await expect(unsaved(page)).toBeHidden();

    await page.reload();
    await expect(calories(page)).toHaveValue('2200');
    await expect(protein(page)).toHaveValue('200');
    await expect(carbs(page)).toHaveValue('215');
    await expect(fat(page)).toHaveValue('60');
    await expect(fiber(page)).toHaveValue('30');
    await expect(unsaved(page)).toBeHidden();

    // History: the new snapshot is Current and effective today (Jan 15); the old plan is still listed.
    const rows = page.getByRole('listitem');
    const current = rows.filter({ hasText: '2,200 kcal · P200 C215 F60' });
    await expect(current).toHaveCount(1);
    await expect(current).toContainText('Current');
    await expect(current).toContainText('since Jan 15');
    const old = rows.filter({ hasText: '2,500 kcal · P180 C280 F70' });
    await expect(old).toHaveCount(1);
    await expect(old).not.toContainText('Current');

    const plans = await api.dietPlans();
    expect(plans).toHaveLength(2);
    expect(plans.filter((p) => p.calories === 2200 && p.fiber === 30)).toHaveLength(1);
  });

  test('goal segment sets calories from the formula TDEE; the pick survives a save and reload', async ({ page, api }) => {
    await api.saveProfile({ height_in: 70, age: 30, sex: 'male' });
    await api.saveWeight('2026-01-14', 180);
    await openTargets(page);

    const hint = page.getByText(/kcal\/day from formula TDEE \(([\d,]+)\)/);
    await expect(hint).toBeVisible();
    const tdee = Number((await hint.innerText()).match(/\(([\d,]+)\)/)![1].replace(/,/g, ''));
    expect(tdee).toBeGreaterThan(2000); // 180 lb, 70 in, 30 y, male, moderately active: ~2,800

    const goal = page.getByRole('group', { name: 'Goal' });
    await goal.getByRole('button', { name: '−500' }).click();
    await expect(calories(page)).toHaveValue(String(tdee - 500));
    await expect(unsaved(page)).toBeVisible();

    const saved = waitForApi(page, SAVE);
    await saveButton(page).click();
    await saved;

    await page.reload();
    await expect(calories(page)).toHaveValue(String(tdee - 500));
    await expect(goal.getByRole('button', { name: '−500' })).toHaveAttribute('aria-pressed', 'true');
    expect((await api.dietPlans()).filter((p) => p.calories === tdee - 500)).toHaveLength(1);
  });

  test("today's /diet summary uses the new targets; past days keep the plan they had", async ({ page, api }) => {
    // Day 3 (Jan 12) exists before the save; day 5 (Jan 10) is only created when it is first opened.
    const seedPlanId = (await api.dietDay(3)).day.plan.ID;
    await openTargets(page);
    await enterNewTargets(page);
    const saved = waitForApi(page, SAVE);
    await saveButton(page).click();
    await saved;

    await page.goto('/diet');
    await expect(page.getByText('0 / 2,200')).toBeVisible();
    await expect(page.getByText('0 / 200 g')).toBeVisible(); // protein
    await expect(page.getByText('0 / 60 g')).toBeVisible(); // fat

    for (const offset of [3, 5]) {
      await page.goto(`/diet?offset=${offset}`);
      await expect(page.getByText('0 / 2,500')).toBeVisible();
      await expect(page.getByText('0 / 2,200')).toBeHidden();
      await expect(page.getByText('0 / 180 g')).toBeVisible();
    }
    const past = (await api.dietDay(3)).day.plan;
    expect(past.ID).toBe(seedPlanId);
    expect(past.calories).toBe(2500);
    expect((await api.dietDay(5)).day.plan.calories).toBe(2500);
    expect((await api.dietDay(0)).day.plan.calories).toBe(2200);
  });

  test('unsaved changes: badge and enabled Save while dirty, Reset restores the saved values', async ({ page }) => {
    await openTargets(page);
    await expect(unsaved(page)).toBeHidden();
    await expect(saveButton(page)).toBeDisabled();
    await expect(page.getByRole('button', { name: 'Reset' })).toBeHidden();

    await enterNewTargets(page);
    await expect(unsaved(page)).toBeVisible();
    await expect(saveButton(page)).toBeEnabled();

    await page.getByRole('button', { name: 'Reset' }).click();
    await expect(calories(page)).toHaveValue('2500');
    await expect(protein(page)).toHaveValue('180');
    await expect(carbs(page)).toHaveValue('280');
    await expect(fat(page)).toHaveValue('70');
    await expect(fiber(page)).toHaveValue('35');
    await expect(unsaved(page)).toBeHidden();
    await expect(saveButton(page)).toBeDisabled();
  });

  test('a failed save keeps the edits dirty, saves nothing, and a retry works', async ({ page, api, e2e }) => {
    await openTargets(page);
    await enterNewTargets(page);

    await e2e.fault({ table: 'plans', op: 'create', times: 1 });
    await saveButton(page).click();
    await expect(page.getByText('Could not save targets', { exact: true })).toBeVisible();
    await expect(unsaved(page)).toBeVisible();
    await expect(saveButton(page)).toBeEnabled();
    await expect(calories(page)).toHaveValue('2200');
    await expect(fiber(page)).toHaveValue('30');
    expect(await api.dietPlans()).toHaveLength(1); // the transaction rolled back: still only the seed plan
    expect((await api.dietDay(0)).day.plan.calories).toBe(2500);

    const saved = waitForApi(page, SAVE);
    await saveButton(page).click();
    await saved;
    await expect(unsaved(page)).toBeHidden();
    await page.reload();
    await expect(calories(page)).toHaveValue('2200');
  });
});

test.describe('diet targets: load error', { tag: ['@diet'] }, () => {
  test('a failed load shows the error; Retry brings the page back', async ({ page, e2e }) => {
    // Queries retry twice on 5xx (3 attempts), so a failing load needs 3 faults per request. The page also
    // fetches the last 7 days (the same endpoint, offsets 1-7) for its intake line, so all 8 requests x
    // 3 attempts = 24 statements fail, whichever request each one lands on. That drains the fault, so the
    // Retry below reaches a healthy database. Skip the backoff after each round of failures.
    const FAILING = 24;
    await e2e.fault({ table: 'days', op: 'query', times: FAILING });
    let failures = 0; // the target request only: offset=0
    page.on('response', (r) => {
      const u = new URL(r.url());
      if (u.pathname === '/api/diet/logs/today' && u.searchParams.get('offset') === '0' && r.status() === 500) failures++;
    });
    await page.goto('/diet/targets');
    for (let i = 1; i <= 3; i++) {
      await expect.poll(() => failures).toBe(i);
      if (i < 3) await page.clock.runFor(5_000);
    }
    await expect.poll(async () => (await e2e.state()).faults[0]?.hits).toBe(FAILING);
    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load your targets" });
    await expect(alert).toBeVisible();
    await expect(calories(page)).toHaveCount(0);

    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(alert).toHaveCount(0);
    await expect(calories(page)).toHaveValue('2500');
  });
});

test.describe('diet targets: activity level', { tag: ['@diet'] }, () => {
  test('changing the activity level saves the profile and persists', async ({ page, api }) => {
    await api.saveProfile({ height_in: 70, age: 30, sex: 'male', activity_level: 'moderately_active' });
    await api.saveWeight('2026-01-14', 180);
    await openTargets(page);

    const activity = page.getByLabel('Activity level');
    await expect(activity).toHaveValue('moderately_active');
    const saved = waitForApi(page, 'PUT /tracking/profile');
    await activity.selectOption('very_active');
    await saved;
    await expect(unsaved(page)).toBeHidden(); // a profile change is not a target change

    await page.reload();
    await expect(page.getByLabel('Activity level')).toHaveValue('very_active');
    expect((await api.profile())?.activity_level).toBe('very_active');
  });
});
