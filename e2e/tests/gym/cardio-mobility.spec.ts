import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import { LEGS_PLAN, pickSliderValue } from './logger';

// The non-lift steps of a session: cardio (/gym/logging/cardio) and the warmup/stretching checklists
// (/gym/logging/mobility/pre|post). The seed plans neither, so each test plans them on today's
// "Legs" plan through the plan endpoints.

const sessionList = (page: Page) => page.getByRole('region', { name: 'Session' });
const stepRow = (page: Page, name: string) => sessionList(page).getByRole('button', { name: new RegExp(`^${name}\\b`) });

test.describe('gym cardio', { tag: ['@gym'] }, () => {
  test('planned cardio: logging minutes marks the step done, and it persists', async ({ page, api }) => {
    await api.setPlannedCardio(LEGS_PLAN, 'Bike', 20);
    await page.goto('/gym/logging');
    await expect(stepRow(page, 'Bike')).toHaveAccessibleName(/^Bike 20 min planned$/);
    await stepRow(page, 'Bike').click();

    await expect(page).toHaveURL(/\/gym\/logging\/cardio$/);
    await expect(page.getByRole('heading', { name: 'Bike', level: 2 })).toBeVisible();
    await pickSliderValue(page, page.getByRole('main'), 30);
    await page.getByRole('textbox', { name: 'Notes' }).fill('Intervals');
    const saved = waitForApi(page, 'POST /workout/logs/cardio');
    await page.getByRole('button', { name: 'Finish cardio' }).click();

    await expect(page).toHaveURL(/\/gym\/logging$/);
    await expect(stepRow(page, 'Bike')).toHaveAccessibleName(/^Bike , done 30 min$/);
    await saved;
    await page.reload();
    await expect(stepRow(page, 'Bike')).toHaveAccessibleName(/^Bike , done 30 min$/);

    await stepRow(page, 'Bike').click();
    await expect(page.getByText('Logged', { exact: true })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'Notes' })).toHaveValue('Intervals');
  });

  // Without planned cardio the session list has no cardio step and no way to add one; the route still
  // opens by URL as "Cardio", but saving fails: the client sends no type and the server answers 400
  // "cardio type is required when the plan has no planned cardio". Pinned: notes/gym-session.md.
  test('unplanned cardio opened directly: the logged minutes show as a done step', async ({ page }) => {
    test.fail(true, 'bug: unplanned cardio cannot be saved (400, no type); see e2e/notes/gym-session.md');
    await page.goto('/gym/logging');
    await expect(stepRow(page, 'Leg Curl')).toBeVisible();
    await expect(stepRow(page, 'Cardio')).toHaveCount(0);

    await page.goto('/gym/logging/cardio');
    await expect(page.getByRole('heading', { name: 'Cardio', level: 2 })).toBeVisible();
    await pickSliderValue(page, page.getByRole('main'), 15);
    const response = page.waitForResponse((r) => r.request().method() === 'POST' && new URL(r.url()).pathname === '/api/workout/logs/cardio');
    await page.getByRole('button', { name: 'Finish cardio' }).click();
    expect((await response).status()).toBe(200);
    await page.reload();
    await expect(stepRow(page, 'Cardio')).toHaveAccessibleName(/^Cardio , done 15 min$/);
  });
});

test.describe('gym mobility', { tag: ['@gym'] }, () => {
  test('warmup: ticked items stay ticked after a reload; all ticked marks the step done', async ({ page, api }) => {
    await api.setPlannedMobility(LEGS_PLAN, { pre: ['Leg swings', 'Hip circles'], post: ['Pigeon stretch'] });
    await page.goto('/gym/logging');
    await expect(stepRow(page, 'Dynamic warmup')).toHaveAccessibleName(/^Dynamic warmup , up next 2 items$/);
    await page.getByRole('button', { name: 'Start warmup' }).click();

    await expect(page).toHaveURL(/\/gym\/logging\/mobility\/pre$/);
    await expect(page.getByRole('heading', { name: 'Dynamic warmup' })).toBeVisible();
    const swings = page.getByRole('checkbox', { name: 'Leg swings' });
    const circles = page.getByRole('checkbox', { name: 'Hip circles' });
    let saved = waitForApi(page, 'POST /workout/logs/mobility/pre');
    await swings.click();
    await expect(swings).toBeChecked();
    await expect(page.getByText('1/2', { exact: true })).toBeVisible();
    await saved;

    await page.reload();
    await expect(swings).toBeChecked();
    await expect(circles).not.toBeChecked();

    saved = waitForApi(page, 'POST /workout/logs/mobility/pre');
    await circles.click();
    await expect(circles).toBeChecked();
    await saved;
    await page.getByRole('button', { name: 'Done' }).click();
    await expect(stepRow(page, 'Dynamic warmup')).toHaveAccessibleName(/^Dynamic warmup , done 2 items$/);
    await page.reload();
    await expect(stepRow(page, 'Dynamic warmup')).toHaveAccessibleName(/^Dynamic warmup , done 2 items$/);
  });

  test('stretching: a ticked item persists and unticking it is saved too', async ({ page, api }) => {
    await api.setPlannedMobility(LEGS_PLAN, { post: ['Pigeon stretch', 'Couch stretch'] });
    await page.goto('/gym/logging');
    await stepRow(page, 'Static stretching').click();
    await expect(page).toHaveURL(/\/gym\/logging\/mobility\/post$/);

    const pigeon = page.getByRole('checkbox', { name: 'Pigeon stretch' });
    let saved = waitForApi(page, 'POST /workout/logs/mobility/post');
    await pigeon.click();
    await saved;
    await page.reload();
    await expect(pigeon).toBeChecked();

    saved = waitForApi(page, 'POST /workout/logs/mobility/post');
    await pigeon.click();
    await expect(pigeon).not.toBeChecked();
    await saved;
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Static stretching' })).toBeVisible();
    await expect(pigeon).not.toBeChecked();
  });
});
