import { expect, test, waitForApi } from '../../fixtures';

// Lookup runs the real backend against the fake upstream (fixtures `upstream`): AniList's GraphQL
// request, parsing and error mapping are real, only the third party is replaced.

const FRIEREN = { id: 154587, title: 'Frieren: Beyond Journey', year: 2023, studio: 'Madhouse', episodes: 28, genres: ['Adventure', 'Fantasy'] };

test('look up an anime, fill the form, save, and it persists', { tag: ['@media'] }, async ({ page, api, upstream }) => {
  upstream.setAniList([FRIEREN]);
  await page.goto('/media/new');
  await page.getByRole('combobox', { name: 'Type' }).selectOption('anime');
  await page.getByLabel('Title').fill('frieren');
  await page.getByRole('button', { name: 'Look up' }).click();
  await page.getByRole('button', { name: /Frieren: Beyond Journey/ }).click();

  await expect(page.getByLabel('Creator')).toHaveValue('Madhouse');
  await expect(page.getByLabel('Year')).toHaveValue('2023');
  await expect(page.getByLabel('Total')).toHaveValue('28');

  const saved = waitForApi(page, 'POST /media');
  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await saved;

  const { items } = await api.get<{ items: { title: string; source: string; external_id: string }[] }>('/media');
  expect(items).toMatchObject([{ title: FRIEREN.title, source: 'anilist', external_id: '154587' }]);
  expect(upstream.requests.length).toBeGreaterThanOrEqual(2); // search + details
});

test('upstream down: lookup shows an error and the form stays usable', { tag: ['@media'] }, async ({ page, upstream, e2e }) => {
  upstream.failAniList(500, 1);
  e2e.expectServerErrors = true; // the backend maps it to 502
  await page.goto('/media/new');
  await page.getByRole('combobox', { name: 'Type' }).selectOption('anime');
  await page.getByLabel('Title').fill('frieren');
  await page.getByRole('button', { name: 'Look up' }).click();
  await expect(page.getByText('Lookup failed').first()).toBeVisible();
  await expect(page.getByLabel('Title')).toHaveValue('frieren');
});
