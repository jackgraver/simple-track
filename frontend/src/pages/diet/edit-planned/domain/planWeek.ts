import type { DietLogsTodayResponse } from "~/api/diet/api";
import type { PlannedMeal, SavedMeal } from "~/types/diet";
import { formatInt, mealMacros, type DietWeekDay, type Macros } from "~/pages/diet/home/dietDay";

/**
 * How a day's plan sits against its kcal target.
 * - empty: nothing planned
 * - under: planned below ON_TARGET_LOW of the target
 * - on: within the band
 * - over: above ON_TARGET_HIGH
 * - none: meals planned but the day has no kcal target
 */
export type PlanStatus = "empty" | "under" | "on" | "over" | "none";

/** Within 90–105% of the kcal target reads as planned. */
export const ON_TARGET_LOW = 0.9;
export const ON_TARGET_HIGH = 1.05;

export type PlanMealRow = {
    id: number;
    name: string;
    macros: Macros;
};

export type PlanDay = DietWeekDay & {
    isPast: boolean;
    meals: PlanMealRow[];
    planned: Macros;
    target: Macros;
    /** Logged kcal (only meaningful for today and past days). */
    loggedKcal: number;
    status: PlanStatus;
    /** Planned kcal as a share of the target, 0–100 for the bar. */
    fill: number;
};

const ZERO: Macros = { calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0 };

export function sumMacros(list: Macros[]): Macros {
    return list.reduce(
        (a, m) => ({
            calories: a.calories + m.calories,
            protein: a.protein + m.protein,
            carbs: a.carbs + m.carbs,
            fat: a.fat + m.fat,
            fiber: a.fiber + m.fiber,
        }),
        { ...ZERO },
    );
}

export function savedMealMacros(sm: SavedMeal): Macros {
    const m: Macros = { ...ZERO };
    for (const item of sm.items ?? []) {
        const a = Number(item.amount) || 0;
        m.calories += (item.food?.calories ?? 0) * a;
        m.protein += (item.food?.protein ?? 0) * a;
        m.carbs += (item.food?.carbs ?? 0) * a;
        m.fat += (item.food?.fat ?? 0) * a;
        m.fiber += (item.food?.fiber ?? 0) * a;
    }
    return m;
}

export function planStatus(planned: number, target: number, mealCount: number): PlanStatus {
    if (mealCount === 0) return "empty";
    if (target <= 0) return "none";
    if (planned < target * ON_TARGET_LOW) return "under";
    if (planned > target * ON_TARGET_HIGH) return "over";
    return "on";
}

export function sortPlanned(planned: PlannedMeal[]): PlannedMeal[] {
    return [...planned].sort((a, b) => (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID);
}

export function plannedKcal(planned: PlannedMeal[]): number {
    return planned.reduce((s, p) => s + mealMacros(p.meal).calories, 0);
}

export function buildPlanDay(day: DietWeekDay, data: DietLogsTodayResponse | undefined): PlanDay {
    const meals = sortPlanned(data?.day.plannedMeals ?? []).map((p) => ({
        id: p.ID,
        name: p.meal?.name ?? "Meal",
        macros: mealMacros(p.meal),
    }));
    const planned = sumMacros(meals.map((m) => m.macros));
    const plan = data?.day.plan;
    const target: Macros = {
        calories: plan?.calories ?? 0,
        protein: plan?.protein ?? 0,
        carbs: plan?.carbs ?? 0,
        fat: plan?.fat ?? 0,
        fiber: plan?.fiber ?? 0,
    };
    return {
        ...day,
        isPast: day.offset > 0,
        meals,
        planned,
        target,
        loggedKcal: data?.totalCalories ?? 0,
        status: planStatus(planned.calories, target.calories, meals.length),
        fill: target.calories > 0 ? Math.min(100, (planned.calories / target.calories) * 100) : meals.length ? 100 : 0,
    };
}

export type WeekSummary = {
    /** Today and later with nothing planned. */
    unplanned: number;
    /** Today and later planned below the band. */
    under: number;
    /** Today and later inside the band. */
    onTarget: number;
    upcoming: number;
};

export function weekSummary(days: PlanDay[]): WeekSummary {
    const upcoming = days.filter((d) => !d.isPast);
    return {
        unplanned: upcoming.filter((d) => d.status === "empty").length,
        under: upcoming.filter((d) => d.status === "under").length,
        onTarget: upcoming.filter((d) => d.status === "on").length,
        upcoming: upcoming.length,
    };
}

function plural(n: number, word: string): string {
    return `${n} ${word}${n === 1 ? "" : "s"}`;
}

/** One line for the header: what still needs doing this week. Null for a fully past week. */
export function formatWeekSummary(s: WeekSummary): { text: string; tone: "warn" | "success" | "neutral" } | null {
    if (s.upcoming === 0) return null;
    const parts: string[] = [];
    if (s.unplanned) parts.push(`${plural(s.unplanned, "day")} with nothing planned`);
    if (s.under) parts.push(`${s.under} under target`);
    if (parts.length) return { text: parts.join(" · "), tone: "warn" };
    if (s.onTarget === s.upcoming) return { text: "Every day ahead is on target", tone: "success" };
    return { text: "Every day ahead has meals", tone: "neutral" };
}

/** "Sat 21" */
export function shortDayLabel(d: Date): string {
    return `${d.toLocaleDateString(undefined, { weekday: "short" })} ${d.getDate()}`;
}

/** "Nov 16 – 22" or "Nov 30 – Dec 6" */
export function weekRangeLabel(first: Date, last: Date): string {
    const m = (d: Date) => d.toLocaleDateString(undefined, { month: "short" });
    return m(first) === m(last)
        ? `${m(first)} ${first.getDate()} – ${last.getDate()}`
        : `${m(first)} ${first.getDate()} – ${m(last)} ${last.getDate()}`;
}

/** "1,010 / 2,500 kcal" or "1,010 kcal" without a target. */
export function kcalLine(planned: number, target: number): string {
    return target > 0 ? `${formatInt(planned)} / ${formatInt(target)} kcal` : `${formatInt(planned)} kcal`;
}

/** Short status for screen readers and badges. */
export function statusText(day: PlanDay): string {
    switch (day.status) {
        case "empty":
            return "Nothing planned";
        case "under":
            return `${formatInt(day.target.calories - day.planned.calories)} kcal to plan`;
        case "over":
            return `${formatInt(day.planned.calories - day.target.calories)} kcal over`;
        case "on":
            return "On target";
        default:
            return "";
    }
}

/** Swap two neighbours; null when the move would leave the list. */
export function moveId(ids: number[], index: number, dir: -1 | 1): number[] | null {
    const j = index + dir;
    if (index < 0 || index >= ids.length || j < 0 || j >= ids.length) return null;
    const next = [...ids];
    [next[index], next[j]] = [next[j]!, next[index]!];
    return next;
}
