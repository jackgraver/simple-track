import type { APIRequestContext, APIResponse } from '@playwright/test';

/**
 * Typed arrange helpers over the app's real API (bound to this worker's backend, authenticated).
 * Grow this per module (README: "Arrange via API, act and assert via UI"). Paths are backend paths,
 * without the /api prefix the browser uses.
 */
export class Api {
  constructor(
    readonly request: APIRequestContext,
    readonly baseUrl: string,
  ) {}

  private async check(r: APIResponse, what: string): Promise<APIResponse> {
    if (!r.ok()) throw new Error(`${what}: ${r.status()} ${await r.text()}`);
    return r;
  }

  async get<T = unknown>(path: string): Promise<T> {
    return (await this.check(await this.request.get(this.baseUrl + path), `GET ${path}`)).json();
  }

  async post<T = unknown>(path: string, data?: unknown): Promise<T> {
    return (await this.check(await this.request.post(this.baseUrl + path, { data }), `POST ${path}`)).json();
  }

  async patch<T = unknown>(path: string, data?: unknown): Promise<T> {
    return (await this.check(await this.request.patch(this.baseUrl + path, { data }), `PATCH ${path}`)).json();
  }

  async delete(path: string): Promise<void> {
    await this.check(await this.request.delete(this.baseUrl + path), `DELETE ${path}`);
  }

  // --- diet ---

  createFood(food: { name: string; serving_type?: string; serving_amount?: number; calories?: number; protein?: number }) {
    return this.post<{ food: { ID: number; name: string } }>('/diet/foods', {
      serving_type: 'g',
      serving_amount: 100,
      calories: 100,
      protein: 10,
      ...food,
    });
  }

  // --- gym ---

  /** Log a set through the real per-set endpoint (idempotent by client_id). `date` is YYYY-MM-DD. */
  logSet(set: { exerciseId: number; reps: number; weight: number; date: string; clientId?: string }) {
    return this.post<{ set: { ID: number; reps: number; weight: number }; created: boolean }>('/workout/sets', {
      client_id: set.clientId ?? crypto.randomUUID(),
      date: set.date,
      exercise_id: set.exerciseId,
      reps: set.reps,
      weight: set.weight,
      weight_setup: '',
    });
  }

  /** Today's workout log (planned + logged exercises), as the logging page loads it. */
  gymToday(offset = 0) {
    return this.get<any>(`/workout/logs/today?offset=${offset}`);
  }

  // --- finances ---

  async createAccount(a: { name: string; kind?: string; opening_balance?: number; balance_date?: string }) {
    const r = await this.post<{ account: { ID: number; name: string } }>('/money/accounts', {
      kind: 'savings',
      institution: '',
      notes: '',
      balance_date: '2026-01-10',
      ...a,
    });
    return r.account;
  }

  addTransaction(accountId: number, t: { kind: 'deposit' | 'withdrawal'; amount: number; date: string; note?: string }) {
    return this.post(`/money/accounts/${accountId}/transactions`, { note: '', ...t });
  }

  // --- grocery ---

  async createGroceryItem(name: string) {
    const r = await this.post<{ item: { ID: number; name: string } }>('/tracking/grocery/items', { name });
    return r.item;
  }

  deleteGroceryItem(id: number) {
    return this.delete(`/tracking/grocery/items/${id}`);
  }

  // --- tracking ---

  saveWeight(date: string, weight_lbs: number) {
    return this.post('/tracking/weight', { date, weight_lbs });
  }

  // gym session

  /** Planned cardio for a workout plan (an empty type clears it). */
  async setPlannedCardio(planId: number, type: string, minutes: number) {
    const path = `/workout/plans/${planId}/planned-cardio`;
    await this.check(await this.request.put(this.baseUrl + path, { data: { type, minutes } }), `PUT ${path}`);
  }

  /** Planned warmup (pre) and stretching (post) checklists for a workout plan. */
  async setPlannedMobility(planId: number, items: { pre?: string[]; post?: string[] }) {
    const path = `/workout/plans/${planId}/planned-mobility`;
    const data = { pre_mobility_items: items.pre ?? [], post_mobility_items: items.post ?? [] };
    await this.check(await this.request.put(this.baseUrl + path, { data }), `PUT ${path}`);
  }

  // diet home

  /** A saved meal from catalogue foods (`amount` is a multiple of the food's serving). Returns its id. */
  async createSavedMeal(name: string, items: { food_id: number; amount: number; group_id?: string; group_label?: string }[]) {
    const r = await this.post<{ saved_meal_id: number }>('/diet/meals/saved-meal/new', {
      name,
      items: items.map((i) => ({ group_id: '', group_label: '', composite_food_id: null, ...i })),
    });
    return r.saved_meal_id;
  }

  /** Plan a saved meal on a day (offset 0 = today, positive = past). Returns the day as the app loads it. */
  addPlannedFromSaved(savedMealId: number, offset = 0) {
    return this.post<DietDayResponse>('/diet/meals/planned/from-saved', {
      saved_meal_id: savedMealId,
      offset,
      client_id: crypto.randomUUID(),
    });
  }

  /** Log a planned meal (by its `meal_id`) on a day row, as the Log button does. */
  logPlannedMeal(mealId: number, dayId: number) {
    return this.post<DietDayResponse>('/diet/meals/meal/log-planned', { meal_id: mealId, day_id: dayId });
  }

  quickLog(entry: { name: string; calories: number; protein?: number; carbs?: number; fat?: number; fiber?: number; offset?: number }) {
    return this.post<DietDayResponse>('/diet/meals/quick-log', {
      protein: 0,
      carbs: 0,
      fat: 0,
      fiber: 0,
      offset: 0,
      client_id: crypto.randomUUID(),
      ...entry,
    });
  }

  /** The day as the diet page loads it. */
  dietDay(offset = 0) {
    return this.get<DietDayResponse>(`/diet/logs/today?offset=${offset}`);
  }

  // diet plan

  /** A day's planned meal names in display order, as the day view loads them. */
  async plannedMeals(offset: number): Promise<string[]> {
    const r = await this.dietDay(offset);
    return [...(r.day.plannedMeals ?? [])]
      .sort((a, b) => (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID)
      .map((m) => m.meal?.name ?? m.name ?? '');
  }

  /** Drink-size presets are stored in oz. */
  async createDrinkSize(name: string, amount_oz: number) {
    return (await this.post<{ preset: { ID: number; name: string; amount_oz: number } }>('/tracking/water/presets', { name, amount_oz })).preset;
  }

  // diet logmeal

  savedMeals() {
    return this.get<{
      saved_meals: { ID: number; name: string; items: { food_id: number; amount: number; group_id: string; group_label: string }[] }[];
    }>('/diet/meals/saved-meal/all');
  }

  /** A logged meal on a day (`offset` days in the past) built from catalogue foods, without the UI. */
  async logCatalogueMeal(name: string, items: { food_id: number; amount: number }[], offset = 0) {
    const r = await this.post<{ meal_id: number }>('/diet/meals/meal/new', {
      meal: { name, items: items.map((i) => ({ group_id: '', group_label: '', composite_food_id: null, ...i })) },
      log: true,
      save_to_library: false,
      offset,
    });
    return r.meal_id;
  }

  allFoods() {
    return this.get<{ foods: { ID: number; name: string; calories: number }[] }>('/diet/meals/food/all');
  }

  // diet targets

  /** Body profile (height, age, sex, activity). Gives /diet/targets a formula TDEE to anchor the goal chips on. */
  saveProfile(p: { height_in: number; age: number; sex: 'male' | 'female'; activity_level?: string }) {
    return this.request
      .put(this.baseUrl + '/tracking/profile', { data: { activity_level: 'moderately_active', ...p } })
      .then((r) => this.check(r, 'PUT /tracking/profile'));
  }

  /** The profile row (null before it is saved). */
  async profile() {
    return (await this.get<{ profile: { activity_level: string } | null }>('/tracking/profile')).profile;
  }

  /** Every diet plan row, as the targets history lists them. */
  async dietPlans() {
    return (
      await this.get<{ plans: { ID: number; name: string; calories: number; protein: number; carbs: number; fat: number; fiber: number; effective_from: string }[] }>(
        '/diet/plans/plan/all',
      )
    ).plans;
  }

  async drinkSizes() {
    return (await this.get<{ presets: { ID: number; name: string; amount_oz: number }[] }>('/tracking/water/presets')).presets;
  }
}

/** A meal inside a diet day (planned or logged). */
export interface DietMeal {
  ID: number;
  name: string;
  items: { food_id: number; amount: number; group_id: string; group_label: string }[];
}

/** The diet day response (`GET /diet/logs/today`, and every day write), loosely typed. */
export interface DietDayResponse {
  day: {
    ID: number;
    plan: { ID: number; name: string; calories: number; protein: number };
    plannedMeals: { ID: number; meal_id: number; display_order?: number; name?: string; meal: DietMeal }[];
    loggedMeals: { ID: number; meal_id: number; meal: DietMeal }[];
  };
  totalCalories: number;
  totalProtein: number;
  totalCarbs: number;
  totalFat: number;
}
