/**
 * Pure cache patches for one diet day (`dietKeys.days.day(date)`), used by the optimistic diet
 * mutations. Each `apply*` has an inverse `revert*` for rollback while other writes to the same day
 * are in flight. The server's full-day response replaces the day once nothing else is pending, so
 * rounding or missing food data in these patches is only ever visible for one round trip.
 *
 * Logging a planned meal moves it to `loggedMeals` (the server only returns planned meals not yet
 * logged); unlogging removes the `DayLog` and, like the server, does not re-plan the meal.
 */
import type { DietLogsTodayResponse } from './api';
import type { DayLog, Food, Meal, PlannedMeal, SavedMeal } from '~/types/diet';

export type Totals = Pick<
    DietLogsTodayResponse,
    'totalCalories' | 'totalProtein' | 'totalFiber' | 'totalCarbs' | 'totalFat'
>;

type MacroSet = { calories: number; protein: number; carbs: number; fat: number; fiber: number };

let tempSeq = 0;
/** Negative ids can never collide with a server row. Callers must not send them to the server. */
export function tempId(): number {
    return -(++tempSeq);
}

export function isTempId(id: number): boolean {
    return id < 0;
}

export function macrosOf(meal: Meal): MacroSet {
    const m: MacroSet = { calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0 };
    for (const item of meal.items ?? []) {
        const a = Number(item.amount) || 0;
        m.calories += (item.food?.calories ?? 0) * a;
        m.protein += (item.food?.protein ?? 0) * a;
        m.carbs += (item.food?.carbs ?? 0) * a;
        m.fat += (item.food?.fat ?? 0) * a;
        m.fiber += (item.food?.fiber ?? 0) * a;
    }
    return m;
}

function shift(old: DietLogsTodayResponse, m: MacroSet, sign: 1 | -1): DietLogsTodayResponse {
    const clamp = (n: number) => Math.max(0, n);
    return {
        ...old,
        totalCalories: clamp(old.totalCalories + sign * m.calories),
        totalProtein: clamp(old.totalProtein + sign * m.protein),
        totalFiber: clamp(old.totalFiber + sign * m.fiber),
        totalCarbs: clamp(old.totalCarbs + sign * m.carbs),
        totalFat: clamp(old.totalFat + sign * m.fat),
    };
}

function withDay(old: DietLogsTodayResponse, patch: Partial<DietLogsTodayResponse['day']>): DietLogsTodayResponse {
    return { ...old, day: { ...old.day, ...patch } };
}

const mealIdOf = (l: DayLog) => l.meal_id || l.meal?.ID;

// ---- Log / unlog -------------------------------------------------------------------------------

/**
 * Log a planned meal: it moves from the planned list to the logged list, as on the server (the day
 * response only returns planned meals not yet logged, and unlogging does not bring them back).
 */
export function applyLogPlanned(old: DietLogsTodayResponse, mealId: number): DietLogsTodayResponse {
    const logged = old.day.loggedMeals ?? [];
    if (logged.some((l) => mealIdOf(l) === mealId)) return old; // idempotent, like the server
    const plannedList = old.day.plannedMeals ?? [];
    const planned = plannedList.find((p) => (p.meal_id || p.meal?.ID) === mealId);
    if (!planned) return old;
    const now = new Date().toISOString();
    const row: DayLog = {
        ID: tempId(),
        created_at: now,
        updated_at: now,
        day_id: old.day.ID,
        meal_id: mealId,
        meal: planned.meal,
    };
    return shift(
        withDay(old, { loggedMeals: [...logged, row], plannedMeals: plannedList.filter((p) => p !== planned) }),
        macrosOf(planned.meal),
        1,
    );
}

/** Undo a log: drop the temp row and put the planned meal back (from the snapshot) if it is gone. */
export function revertLogPlanned(
    cur: DietLogsTodayResponse,
    mealId: number,
    snapshot?: DietLogsTodayResponse,
): DietLogsTodayResponse {
    const logged = cur.day.loggedMeals ?? [];
    const row = logged.find((l) => mealIdOf(l) === mealId && isTempId(l.ID));
    let next = cur;
    if (row) next = shift(withDay(cur, { loggedMeals: logged.filter((l) => l !== row) }), macrosOf(row.meal), -1);
    const plannedNow = next.day.plannedMeals ?? [];
    const before = snapshot?.day.plannedMeals?.find((p) => (p.meal_id || p.meal?.ID) === mealId);
    if (before && !plannedNow.some((p) => p.ID === before.ID)) {
        // display_order is kept on rows, so the readers' sort restores the position.
        next = withDay(next, { plannedMeals: [...plannedNow, before] });
    }
    return next;
}

export function applyUnlog(old: DietLogsTodayResponse, mealId: number): DietLogsTodayResponse {
    const logged = old.day.loggedMeals ?? [];
    const row = logged.find((l) => mealIdOf(l) === mealId);
    if (!row) return old;
    return shift(withDay(old, { loggedMeals: logged.filter((l) => l !== row) }), macrosOf(row.meal), -1);
}

/** Put the removed row back (from the snapshot) at its original index. */
export function revertUnlog(
    cur: DietLogsTodayResponse,
    mealId: number,
    snapshot: DietLogsTodayResponse,
): DietLogsTodayResponse {
    const before = snapshot.day.loggedMeals ?? [];
    const idx = before.findIndex((l) => mealIdOf(l) === mealId);
    if (idx < 0) return cur;
    const row = before[idx];
    const logged = cur.day.loggedMeals ?? [];
    if (logged.some((l) => l.ID === row.ID)) return cur;
    const next = [...logged];
    next.splice(Math.min(idx, next.length), 0, row);
    return shift(withDay(cur, { loggedMeals: next }), macrosOf(row.meal), 1);
}

// ---- Quick log ---------------------------------------------------------------------------------

export type QuickLogEntry = {
    client_id: string;
    name: string;
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
    replace_meal_id?: number;
};

function quickMeal(e: QuickLogEntry): Meal {
    const now = new Date().toISOString();
    const food: Food = {
        ID: tempId(),
        created_at: now,
        updated_at: now,
        name: e.name,
        serving_type: 'serving',
        serving_amount: 1,
        calories: e.calories,
        protein: e.protein,
        fiber: e.fiber,
        carbs: e.carbs,
        fat: e.fat,
        quick_entry: true,
    };
    const mealId = tempId();
    return {
        ID: mealId,
        created_at: now,
        updated_at: now,
        name: e.name,
        client_id: e.client_id,
        items: [
            {
                ID: tempId(),
                created_at: now,
                updated_at: now,
                meal_id: mealId,
                food_id: food.ID,
                food,
                amount: 1,
            },
        ],
    };
}

/** Upsert by `client_id` (idempotent, so retry re-applies safely). Replacing removes the edited row first. */
export function applyQuickLog(old: DietLogsTodayResponse, e: QuickLogEntry): DietLogsTodayResponse {
    let next = old;
    const logged = old.day.loggedMeals ?? [];
    if (logged.some((l) => l.meal?.client_id === e.client_id)) return old;
    let at = logged.length;
    if (e.replace_meal_id) {
        const idx = logged.findIndex((l) => mealIdOf(l) === e.replace_meal_id);
        if (idx >= 0) {
            at = idx;
            next = applyUnlog(old, e.replace_meal_id);
        }
    }
    const meal = quickMeal(e);
    const now = new Date().toISOString();
    const row: DayLog = {
        ID: tempId(),
        created_at: now,
        updated_at: now,
        day_id: old.day.ID,
        meal_id: meal.ID,
        meal,
    };
    const list = [...(next.day.loggedMeals ?? [])];
    list.splice(Math.min(at, list.length), 0, row);
    return shift(withDay(next, { loggedMeals: list }), macrosOf(meal), 1);
}

export function revertQuickLog(
    cur: DietLogsTodayResponse,
    e: QuickLogEntry,
    snapshot: DietLogsTodayResponse,
): DietLogsTodayResponse {
    const logged = cur.day.loggedMeals ?? [];
    const row = logged.find((l) => l.meal?.client_id === e.client_id);
    let next = cur;
    if (row) next = shift(withDay(cur, { loggedMeals: logged.filter((l) => l !== row) }), macrosOf(row.meal), -1);
    if (e.replace_meal_id) next = revertUnlog(next, e.replace_meal_id, snapshot);
    return next;
}

// ---- Planned meals -----------------------------------------------------------------------------

function nextOrder(planned: PlannedMeal[]): number {
    return planned.reduce((max, p) => Math.max(max, p.display_order ?? 0), 0) + 1;
}

export function applyAddPlanned(
    old: DietLogsTodayResponse,
    saved: SavedMeal,
    clientId: string,
): DietLogsTodayResponse {
    const planned = old.day.plannedMeals ?? [];
    if (planned.some((p) => p.client_id === clientId)) return old;
    const now = new Date().toISOString();
    const mealId = tempId();
    const meal: Meal = {
        ID: mealId,
        created_at: now,
        updated_at: now,
        name: saved.name,
        items: (saved.items ?? []).map((it) => ({
            ID: tempId(),
            created_at: now,
            updated_at: now,
            meal_id: mealId,
            food_id: it.food_id,
            food: it.food,
            amount: it.amount,
            group_id: it.group_id,
            group_label: it.group_label,
            composite_food_id: it.composite_food_id,
        })),
    };
    const row: PlannedMeal = {
        ID: tempId(),
        created_at: now,
        updated_at: now,
        day_id: old.day.ID,
        meal_id: mealId,
        meal,
        display_order: nextOrder(planned),
        saved_meal_id: saved.ID,
        client_id: clientId,
    };
    return withDay(old, { plannedMeals: [...planned, row] });
}

export function revertAddPlanned(cur: DietLogsTodayResponse, clientId: string): DietLogsTodayResponse {
    const planned = cur.day.plannedMeals ?? [];
    if (!planned.some((p) => p.client_id === clientId)) return cur;
    return withDay(cur, { plannedMeals: planned.filter((p) => p.client_id !== clientId) });
}

export function applyDeletePlanned(old: DietLogsTodayResponse, plannedMealId: number): DietLogsTodayResponse {
    const planned = old.day.plannedMeals ?? [];
    if (!planned.some((p) => p.ID === plannedMealId)) return old;
    return withDay(old, { plannedMeals: planned.filter((p) => p.ID !== plannedMealId) });
}

export function revertDeletePlanned(
    cur: DietLogsTodayResponse,
    plannedMealId: number,
    snapshot: DietLogsTodayResponse,
): DietLogsTodayResponse {
    const before = snapshot.day.plannedMeals ?? [];
    const row = before.find((p) => p.ID === plannedMealId);
    const planned = cur.day.plannedMeals ?? [];
    if (!row || planned.some((p) => p.ID === plannedMealId)) return cur;
    // display_order is kept on rows, so display order is restored by the readers' sort.
    return withDay(cur, { plannedMeals: [...planned, row] });
}

/** Rewrite `display_order` to the index in `ids`; rows not named keep their place after the named ones. */
export function applyReorder(old: DietLogsTodayResponse, ids: number[]): DietLogsTodayResponse {
    const planned = old.day.plannedMeals ?? [];
    const pos = new Map(ids.map((id, i) => [id, i + 1]));
    const rest = planned.length + 1;
    return withDay(old, {
        plannedMeals: planned.map((p) => ({ ...p, display_order: pos.get(p.ID) ?? rest + (p.display_order ?? 0) })),
    });
}

/** Restore the previous `display_order` of every row that still exists. */
export function revertReorder(
    cur: DietLogsTodayResponse,
    snapshot: DietLogsTodayResponse,
): DietLogsTodayResponse {
    const orders = new Map((snapshot.day.plannedMeals ?? []).map((p) => [p.ID, p.display_order]));
    return withDay(cur, {
        plannedMeals: (cur.day.plannedMeals ?? []).map((p) =>
            orders.has(p.ID) ? { ...p, display_order: orders.get(p.ID) } : p,
        ),
    });
}

// ---- Unlogging a row that is still saving -----------------------------------------------------

/**
 * The server id of the logged meal to delete. A quick log still saving has a temp id: its real id
 * is in the day the quick log's request returned (matched by the meal's `client_id`). `null` when
 * the quick log never saved (failed), so there is nothing to delete on the server.
 */
export function resolveLoggedMealId(
    mealId: number,
    mealClientId: string | null | undefined,
    savedDays: readonly (DietLogsTodayResponse | undefined)[],
): number | null {
    if (!isTempId(mealId)) return mealId;
    if (!mealClientId) return null;
    for (const day of savedDays) {
        const row = day?.day.loggedMeals?.find((l) => l.meal?.client_id === mealClientId);
        const id = row ? mealIdOf(row) : undefined;
        if (id && !isTempId(id)) return id;
    }
    return null;
}
