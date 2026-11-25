import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// /gym/weight has no list: the history is a canvas chart. What the page shows as text is the slider's
// value (prefilled with the selected day's log, else the latest one), the metrics card for the selected
// day (only from an import), and the overwrite prompt, which quotes the saved value. Specs assert those.
// The overwrite prompt's optimistic path is in tests/optimistic/instant-state.spec.ts.
const SAVE = 'POST /tracking/weight';
const IMPORT = 'POST /tracking/weight/import';

const sliderValue = (page: Page) => page.getByRole('paragraph').filter({ hasText: /^\s*[\d.]+\s*$/ }).first();
const saveButton = (page: Page) => page.getByRole('button', { name: 'Save', exact: true });

/** Click a slider tick and wait for its smooth scroll to settle (see logging.spec.ts, setReps). */
async function pickWeight(page: Page, value: string) {
  await page.getByRole('button', { name: value, exact: true }).click();
  await expect(sliderValue(page)).toHaveText(value);
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
  await expect(sliderValue(page)).toHaveText(value);
}

/** Open the overwrite prompt and return its text, then cancel it (nothing is written). */
async function overwritePrompt(page: Page) {
  await saveButton(page).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  const text = await dialog.innerText();
  await dialog.getByRole('button', { name: 'Cancel' }).click();
  await expect(dialog).toHaveCount(0);
  return text;
}

test.describe('gym weight', { tag: ['@gym'] }, () => {
  test("save today's weight: the slider shows it after a reload, one row in the DB", async ({ page, api, db }) => {
    await api.saveWeight('2026-01-10', 180); // the slider starts at the latest log
    await page.goto('/gym/weight');
    await expect(page.getByText('Weight in lbs (Thursday January 15th, 2026)')).toBeVisible();
    await expect(sliderValue(page)).toHaveText('180');

    await pickWeight(page, '180.5');
    const saved = waitForApi(page, SAVE);
    await saveButton(page).click(); // no log for today yet: no overwrite prompt
    await expect(page.getByText('Saved', { exact: true })).toBeVisible();
    await saved;

    await page.reload();
    await expect(sliderValue(page)).toHaveText('180.5'); // today's log, not the Jan 10 one
    expect(await overwritePrompt(page)).toContain('You already logged 180.5 lbs for Thursday January 15th, 2026');
    const { rows } = await db.query(
      `SELECT to_char(date AT TIME ZONE 'America/Toronto', 'YYYY-MM-DD') AS day, weight_lbs::float AS w
       FROM body_weight_logs WHERE deleted_at IS NULL ORDER BY date`,
    );
    expect(rows).toEqual([
      { day: '2026-01-10', w: 180 },
      { day: '2026-01-15', w: 180.5 },
    ]);
  });

  test('import: pasted Renpho rows are created and updated, and show on their day', async ({ page, api, db }) => {
    await api.saveWeight('2026-01-13', 170); // the import overwrites this day
    const csv = [
      'Date,Time,Weight(lb),Body Fat Percentage(%),BMI',
      '2026.01.12,07:01:00,182.4,18.1,24.7',
      '2026.01.13,07:02:00,182.0,18.0,24.6',
      '2026.01.14,07:03:00,181.6,17.9,24.5',
    ].join('\n');

    await page.goto('/gym/weight');
    await page.getByRole('button', { name: 'Import' }).click();
    const dialog = page.getByRole('dialog', { name: 'Import from Renpho' });
    await dialog.getByRole('textbox').fill(csv);
    const imported = waitForApi(page, IMPORT);
    await dialog.getByRole('button', { name: 'Import', exact: true }).click();
    await imported;
    await expect(page.getByText('Imported 2 new, 1 updated', { exact: true })).toBeVisible();
    await expect(dialog).toHaveCount(0);

    const { rows } = await db.query(
      `SELECT to_char(date AT TIME ZONE 'America/Toronto', 'YYYY-MM-DD') AS day, weight_lbs::float AS w, body_fat_pct::float AS bf
       FROM body_weight_logs WHERE deleted_at IS NULL ORDER BY date`,
    );
    expect(rows).toEqual([
      { day: '2026-01-12', w: 182.4, bf: 18.1 },
      { day: '2026-01-13', w: 182, bf: 18 },
      { day: '2026-01-14', w: 181.6, bf: 17.9 },
    ]);

    // Yesterday (offset=1) shows its imported weight and metrics; today falls back to the latest.
    await page.goto('/gym/weight?offset=1');
    await expect(page.getByText('Weight in lbs (Wednesday January 14th, 2026)')).toBeVisible();
    await expect(sliderValue(page)).toHaveText('181.6');
    await expect(page.getByText('Body fat', { exact: true })).toBeVisible();
    await expect(page.getByText('17.9%', { exact: true })).toBeVisible();
    await page.goto('/gym/weight?offset=2');
    await expect(sliderValue(page)).toHaveText('182'); // the overwritten day carries the import's value
    await expect(page.getByText('18%', { exact: true })).toBeVisible();
  });

  test.describe('import errors', () => {
    test.use({ allowedErrors: [/status of 400/] });

    test('a CSV without a weight column is rejected and the dialog keeps the text', async ({ page, db }) => {
      await page.goto('/gym/weight');
      await page.getByRole('button', { name: 'Import' }).click();
      const dialog = page.getByRole('dialog', { name: 'Import from Renpho' });
      await dialog.getByRole('textbox').fill('Date,Steps\n2026.01.14,8000');
      await dialog.getByRole('button', { name: 'Import', exact: true }).click();
      await expect(page.getByText('missing Weight(lb) column', { exact: true }).first()).toBeVisible();
      await expect(dialog.getByRole('textbox')).toHaveValue('Date,Steps\n2026.01.14,8000');
      const { rows } = await db.query(`SELECT count(*)::int AS n FROM body_weight_logs`);
      expect(rows[0].n).toBe(0);
    });

    // Bug (e2e/notes/gym-tracking.md): the dialog toasts the error and so does the global
    // MutationCache.onError (the import mutation has no `meta.handled`), so it shows twice.
    test('the import error is shown once', async ({ page }) => {
      test.fail(true, 'duplicate error toast on import, see e2e/notes/gym-tracking.md');
      await page.goto('/gym/weight');
      await page.getByRole('button', { name: 'Import' }).click();
      const dialog = page.getByRole('dialog', { name: 'Import from Renpho' });
      await dialog.getByRole('textbox').fill('Date,Steps\n2026.01.14,8000');
      await dialog.getByRole('button', { name: 'Import', exact: true }).click();
      await expect(page.getByText('missing Weight(lb) column', { exact: true }).first()).toBeVisible();
      await expect(page.getByText('missing Weight(lb) column', { exact: true })).toHaveCount(1, { timeout: 1_000 });
    });
  });

  test.describe('late evening in Toronto (already the next day in UTC)', () => {
    // 22:30 EST is 03:30 UTC on Jan 16: a UTC slip would save "yesterday" as Jan 15.
    test.use({ now: new Date('2026-01-15T22:30:00-05:00') });

    test('past day (offset=1): the weight lands on Jan 14, Toronto time', async ({ page, api, db }) => {
      await api.saveWeight('2026-01-10', 180);
      await page.goto('/gym/weight?offset=1');
      await expect(page.getByText('Weight in lbs (Wednesday January 14th, 2026)')).toBeVisible();
      await expect(sliderValue(page)).toHaveText('180');
      const saved = waitForApi(page, SAVE);
      await saveButton(page).click();
      await saved;

      const { rows } = await db.query(
        `SELECT to_char(date AT TIME ZONE 'America/Toronto', 'YYYY-MM-DD HH24:MI') AS at
         FROM body_weight_logs WHERE deleted_at IS NULL ORDER BY date`,
      );
      expect(rows.map((r) => r.at)).toEqual(['2026-01-10 00:00', '2026-01-14 00:00']);

      await page.reload();
      expect(await overwritePrompt(page)).toContain('You already logged 180 lbs for Wednesday January 14th, 2026');
      await page.goto('/gym/weight'); // today (Jan 15) has no log: Save goes straight through, no prompt
      await expect(page.getByText('Weight in lbs (Thursday January 15th, 2026)')).toBeVisible();
      const today = waitForApi(page, SAVE);
      await saveButton(page).click();
      await expect(page.getByRole('dialog')).toHaveCount(0);
      await today;
    });
  });
});
