import type { DayLog, Meal, PlannedMeal } from "~/types/diet";

export type Macros = {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
};

export type MacroKey = keyof Macros;

/** Sum of a meal's items. `amount` is a multiple of the food's serving. */
export function mealMacros(meal: Meal): Macros {
    const m: Macros = { calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0 };
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

export type Remaining = {
    /** Target minus eaten, never negative. */
    left: number;
    /** Eaten minus target, never negative. */
    over: number;
    /** Bar fill, 0–100. */
    fill: number;
    /** No target set (0), so "left" means nothing. */
    noTarget: boolean;
};

export function remaining(eaten: number, target: number): Remaining {
    const e = Math.max(0, eaten || 0);
    const t = Math.max(0, target || 0);
    if (t === 0) return { left: 0, over: 0, fill: e > 0 ? 100 : 0, noTarget: true };
    return {
        left: Math.max(0, t - e),
        over: Math.max(0, e - t),
        fill: Math.min(100, (e / t) * 100),
        noTarget: false,
    };
}

/** Over by less than this share of the target reads as "on target", not a warning. */
const OVER_TOLERANCE = 0.05;

export function isMeaningfullyOver(eaten: number, target: number): boolean {
    return target > 0 && eaten > target * (1 + OVER_TOLERANCE);
}

export type PlannedRow = {
    planned: PlannedMeal;
    logged: boolean;
};

/**
 * Planned meals in display order, each marked logged when a day log points at the same meal.
 * Logging a planned meal creates a DayLog with the planned `meal_id`, so the match is exact.
 * (The `logged` flag on PlannedMeal is not reset when the log is deleted, so it isn't used.)
 */
export function plannedRows(planned: PlannedMeal[], logged: DayLog[]): PlannedRow[] {
    const loggedIds = new Set(logged.map((l) => l.meal_id));
    return [...planned]
        .sort((a, b) => (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID)
        .map((p) => ({ planned: p, logged: loggedIds.has(p.meal_id) }));
}

const intFmt = new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 });

/** 1726.4 -> "1,726" */
export function formatInt(n: number): string {
    return intFmt.format(Math.round(n || 0));
}

/** "774 kcal · P 69 · C 70 · F 21" */
export function macroLine(m: Macros): string {
    return `${formatInt(m.calories)} kcal · P ${formatInt(m.protein)} · C ${formatInt(m.carbs)} · F ${formatInt(m.fat)}`;
}

// ---- Week -------------------------------------------------------------------------------------

export type DietWeekDay = {
    key: string; // YYYY-MM-DD
    date: Date;
    /** Diet `offset` query: 0 = today, positive = past, negative = future. */
    offset: number;
    isToday: boolean;
    isFuture: boolean;
    isSelected: boolean;
};

const DAY_MS = 86_400_000;

function startOfDay(d: Date): Date {
    return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

function addDays(d: Date, n: number): Date {
    return new Date(d.getFullYear(), d.getMonth(), d.getDate() + n);
}

export function dateKey(d: Date): string {
    const mo = String(d.getMonth() + 1).padStart(2, "0");
    const da = String(d.getDate()).padStart(2, "0");
    return `${d.getFullYear()}-${mo}-${da}`;
}

/** The Sunday-to-Saturday week holding the selected day (same week start as /gym). */
export function dietWeek(selectedOffset: number, now: Date = new Date()): DietWeekDay[] {
    const today = startOfDay(now);
    const selected = addDays(today, -selectedOffset);
    const start = addDays(selected, -selected.getDay());
    return Array.from({ length: 7 }, (_, i) => {
        const date = addDays(start, i);
        const offset = Math.round((today.getTime() - date.getTime()) / DAY_MS);
        return {
            key: dateKey(date),
            date,
            offset,
            isToday: offset === 0,
            isFuture: offset < 0,
            isSelected: offset === selectedOffset,
        };
    });
}

/** Average kcal over the days that have anything logged. Null when none do. */
export function averageLogged(kcal: (number | null)[]): number | null {
    const logged = kcal.filter((k): k is number => k != null && k > 0);
    if (logged.length === 0) return null;
    return logged.reduce((s, k) => s + k, 0) / logged.length;
}
