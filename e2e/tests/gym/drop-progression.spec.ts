import type { Page } from '@playwright/test';
import { expect, test } from '../../fixtures';
import { LAST_WEEK, LEGS, TODAY, openExercise, setsList } from './logger';

// A weight drop partway through a session: the next session starts at the final set's weight, and +5
// is earned only when every set was at one weight and hit the rollover (`sessionOutcome`).
// Back Squat with rollover 12; each test arranges its own sets.

const SQUAT = LEGS.squat;
const weightSwitch = (page: Page) => page.getByRole('group', { name: "Today's weight" });
const setupCard = (page: Page) => page.getByRole('button', { name: /Change weight and setup$/ });

test.describe('gym drop progression', { tag: ['@gym'] }, () => {
  test.beforeEach(async ({ db }) => {
    await db.query(`UPDATE exercises SET rep_rollover = 12 WHERE id = $1`, [SQUAT.id]);
  });

  test('a drop last time opens at the dropped weight, with no increase, and says why', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT.id, reps: 12, weight: 100, date: LAST_WEEK });
    await api.logSet({ exerciseId: SQUAT.id, reps: 12, weight: 90, date: LAST_WEEK });

    await openExercise(page, SQUAT);
    await expect(setupCard(page)).toHaveAccessibleName(/^Weight 90 pounds/);
    await expect(weightSwitch(page)).toHaveCount(0);

    await setupCard(page).click();
    await expect(page.getByRole('dialog')).toContainText(
      'Dropped to 90 lb last time, so starting there. Keep every set at 90 lb and hit 12 to go up.',
    );
  });

  test('a drop today: the verdict says where next time starts', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT.id, reps: 12, weight: 100, date: TODAY });
    await api.logSet({ exerciseId: SQUAT.id, reps: 12, weight: 90, date: TODAY });

    await openExercise(page, SQUAT);
    await expect(page.getByText('Dropped to 90 lb · start at 90 next time')).toBeVisible();
    await expect(setsList(page)).toBeVisible();
  });
});
