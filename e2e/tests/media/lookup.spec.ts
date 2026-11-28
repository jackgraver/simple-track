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

// The catalog is shared: adding something already in it saves that entry instead of re-entering it.
test('a catalog entry can be saved to your list from the add page', { tag: ['@media'] }, async ({ page, api }) => {
  const created = await api.post<{ item: { ID: number; media_item_id: number } }>('/media', {
    type: 'movie', title: 'Heat', source: 'tmdb', external_id: 'movie/949', status: 'completed', rating: 4.5,
  });
  await api.delete(`/media/${created.item.ID}`); // off your list, still in the catalog

  await page.goto('/media/new');
  await page.getByLabel('Title').fill('heat');
  await page.getByRole('button', { name: /Heat/ }).click();
  await expect(page).toHaveURL(/\/media\/\d+$/);

  const { items } = await api.get<{ items: { title: string; media_item_id: number; status: string; rating: number | null }[] }>('/media');
  // Same catalog row, but a fresh save: the old rating is gone, status starts at planned.
  expect(items).toMatchObject([{ title: 'Heat', media_item_id: created.item.media_item_id, status: 'planned', rating: null }]);
});

test('the shared catalog only takes covers from lookup image hosts', { tag: ['@media'] }, async ({ api }) => {
  const res = await api.request.post(`${api.baseUrl}/media`, { data: { type: 'movie', title: 'x', cover_url: 'https://evil.example/p.gif' } });
  expect(res.status()).toBe(400);
});
