import type { Locator, Page } from '@playwright/test';
import { expect } from '../../fixtures';

// Shared helpers for the per-exercise logger (/gym/logging/:id). Not a spec: Playwright only runs
// *.spec.ts. Seed: "Legs" is today's plan (the epoch is a Thursday), plan 3, exercises 9-12.

export const TODAY = '2026-01-15';
export const LAST_WEEK = '2026-01-08';
export const LEGS_PLAN = 3;
export const LEGS = {
  squat: { id: 9, name: 'Back Squat' },
  rdl: { id: 10, name: 'Romanian Deadlift' },
  legPress: { id: 11, name: 'Leg Press' },
  legCurl: { id: 12, name: 'Leg Curl' },
} as const;
export const SETS = 'POST /workout/sets';

export const setsList = (page: Page) => page.getByRole('region', { name: 'Sets' });
export const setRow = (page: Page, n: number) => setsList(page).getByText(new RegExp(`^Set ${n} ·`));
/** The primary action: "Log Set N" while logging, "Save Set N" while editing a set. */
export const logButton = (page: Page, n: number) => page.getByRole('button', { name: `Log Set ${n}`, exact: true });
export const saveButton = (page: Page, n: number) => page.getByRole('button', { name: `Save Set ${n}`, exact: true });

export async function openExercise(page: Page, ex: { id: number; name: string }) {
  await page.goto(`/gym/logging/${ex.id}`);
  await expect(page.getByRole('heading', { name: ex.name })).toBeVisible();
}

/** Resolves once nothing on the page has scrolled for ~300ms. */
export function scrollSettled(page: Page) {
  return page.evaluate(
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
}

/**
 * Pick a value on a scroll-driven NumberSlider (reps, cardio minutes) inside `scope`. A tick click
 * smooth-scrolls the strip and the selected tick gets aria-current. Wait for the scroll to settle and
 * check again: a late scroll event could still move the value (BACKLOG: reps slider race).
 */
export async function pickSliderValue(page: Page, scope: Locator, value: number) {
  const tick = scope.getByRole('button', { name: String(value), exact: true });
  await tick.click();
  await expect(tick).toHaveAttribute('aria-current', 'true');
  await scrollSettled(page);
  await expect(tick).toHaveAttribute('aria-current', 'true');
}

const repsSlider = (page: Page) => page.getByRole('region', { name: 'Reps' });

/** Set the logger's reps slider. */
export const setReps = (page: Page, reps: number) => pickSliderValue(page, repsSlider(page), reps);

/** Assert the reps slider's current value without touching it. */
export async function expectReps(page: Page, reps: number) {
  await expect(repsSlider(page).getByRole('button', { name: String(reps), exact: true })).toHaveAttribute('aria-current', 'true');
}
