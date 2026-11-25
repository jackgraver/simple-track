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

  // diet plan

  /** A saved meal of seed foods (amount = servings of the food's serving size). Returns its id. */
  async createSavedMeal(name: string, items: { foodId: number; amount: number }[]) {
    const r = await this.post<{ saved_meal_id: number }>('/diet/meals/saved-meal/new', {
      name,
      items: items.map((i) => ({ food_id: i.foodId, amount: i.amount })),
    });
    return r.saved_meal_id;
  }

  /** Plan a saved meal on a day (`offset` days in the past, negative = future), like the plan page does. */
  addPlannedFromSaved(savedMealId: number, offset: number) {
    return this.post('/diet/meals/planned/from-saved', {
      saved_meal_id: savedMealId,
      offset,
      client_id: crypto.randomUUID(),
    });
  }

  /** A day's planned meal names in display order, as the day view loads them. */
  async plannedMeals(offset: number): Promise<string[]> {
    const r = await this.get<any>(`/diet/logs/today?offset=${offset}`);
    return [...(r.day.plannedMeals ?? [])]
      .sort((a: any, b: any) => (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID)
      .map((m: any) => m.meal?.name ?? m.name);
  }
}
