import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import { LAST_WEEK, LEGS, SETS, logButton, openExercise, setRow } from './logger';

// The go-up/stay switch on the logger's setup card. When every set of last session hit the exercise's
// rep rollover, today's weight is last weight + 5 ("up", the default); the switch keeps last weight
// ("stay") until the first set is logged. Seed: last week 100 lb x 12, 12 on Back Squat, rollover 12.

const SQUAT = LEGS.squat;
const weightSwitch = (page: Page) => page.getByRole('group', { name: "Today's weight" });
const option = (page: Page, weight: number) => weightSwitch(page).getByRole('button', { name: String(weight), exact: true });
const setupCard = (page: Page) => page.getByRole('button', { name: /Change weight and setup$/ });

test.describe('gym weight choice', { tag: ['@gym'] }, () => {
  test.beforeEach(async ({ api, db }) => {
    await db.query(`UPDATE exercises SET rep_rollover = 12 WHERE id = $1`, [SQUAT.id]);
    await api.logSet({ exerciseId: SQUAT.id, reps: 12, weight: 100, date: LAST_WEEK });
    await api.logSet({ exerciseId: SQUAT.id, reps: 12, weight: 100, date: LAST_WEEK });
  });

  test('earned increase: goes up by default, one tap stays, and it can flip back', async ({ page }) => {
    await openExercise(page, SQUAT);
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 105 pounds/);
    await expect(option(page, 105)).toHaveAttribute('aria-pressed', 'true');

    await option(page, 100).click();
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 100 pounds/);
    await expect(option(page, 100)).toHaveAttribute('aria-pressed', 'true');
    await expect(page.getByRole('dialog')).toHaveCount(0); // the switch is not the card's tap target

    await option(page, 105).click();
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 105 pounds/);
  });

  test('the choice is logged with the set, and the switch goes away after set 1', async ({ page, db }) => {
    await openExercise(page, SQUAT);
    await option(page, 100).click();
    const saved = waitForApi(page, SETS);
    await logButton(page, 1).click();
    await expect(setRow(page, 1)).toContainText('100 ×');
    await saved;
    await expect(weightSwitch(page)).toHaveCount(0);
    const { rows } = await db.query(`SELECT weight FROM logged_sets ORDER BY id DESC LIMIT 1`);
    expect(Number(rows[0].weight)).toBe(100);
  });

  test('leaving and coming back before set 1 keeps the choice', async ({ page }) => {
    await openExercise(page, SQUAT);
    await option(page, 100).click();
    await page.goto('/gym/logging');
    await openExercise(page, SQUAT);
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 100 pounds/);
    await expect(option(page, 100)).toHaveAttribute('aria-pressed', 'true');
  });

  test('no earned increase: no switch', async ({ page, db }) => {
    await db.query(`UPDATE exercises SET rep_rollover = 15 WHERE id = $1`, [SQUAT.id]);
    await openExercise(page, SQUAT);
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 100 pounds/);
    await expect(weightSwitch(page)).toHaveCount(0);
  });
});
