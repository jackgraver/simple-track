// Derived numbers for /diet/targets: weight trend, recent intake, observed maintenance and the macro split.
// Pure functions so the page template only formats and lays out.

export type MacroKey = "protein" | "carbs" | "fat";

export type Targets = {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
};

/** kcal per lb of body weight, the usual rule of thumb for energy balance. */
export const KCAL_PER_LB = 3500;

const DAY_MS = 24 * 60 * 60 * 1000;

/**
 * Least-squares slope of weight over time, in lb per week, from logs inside the last `windowDays`.
 * Null when there are fewer than 3 logs or they span less than a week (too noisy to mean anything).
 */
export function weightTrendLbPerWeek(
    logs: { date: string; weight_lbs: number }[] | undefined,
    now: Date = new Date(),
    windowDays = 28,
): number | null {
    if (!logs) return null;
    const cutoff = now.getTime() - windowDays * DAY_MS;
    const pts = logs
        .filter((l) => l.weight_lbs > 0)
        .map((l) => ({ t: new Date(l.date).getTime(), w: l.weight_lbs }))
        .filter((p) => Number.isFinite(p.t) && p.t >= cutoff && p.t <= now.getTime());
    if (pts.length < 3) return null;
    const ts = pts.map((p) => p.t);
    if (Math.max(...ts) - Math.min(...ts) < 7 * DAY_MS) return null;
    const n = pts.length;
    const meanT = ts.reduce((a, b) => a + b, 0) / n;
    const meanW = pts.reduce((a, p) => a + p.w, 0) / n;
    let num = 0;
    let den = 0;
    for (const p of pts) {
        num += (p.t - meanT) * (p.w - meanW);
        den += (p.t - meanT) ** 2;
    }
    if (den === 0) return null;
    return (num / den) * 7 * DAY_MS;
}

export type IntakeSummary = {
    /** Days in the window with anything logged. */
    loggedDays: number;
    /** Days asked about (logged or not). */
    totalDays: number;
    avgCalories: number | null;
    avgProtein: number | null;
};

/** Average intake over the days that have any calories logged. Empty days are skipped, not counted as zero. */
export function intakeSummary(days: { totalCalories: number; totalProtein: number }[]): IntakeSummary {
    const logged = days.filter((d) => d.totalCalories > 0);
    const n = logged.length;
    return {
        loggedDays: n,
        totalDays: days.length,
        avgCalories: n ? logged.reduce((a, d) => a + d.totalCalories, 0) / n : null,
        avgProtein: n ? logged.reduce((a, d) => a + d.totalProtein, 0) / n : null,
    };
}

/** Minimum logged days (of the last 7) before observed maintenance is shown. */
export const MIN_LOGGED_DAYS = 4;

/**
 * Maintenance implied by what you ate and how your weight moved: losing 0.5 lb/wk while eating 2,400
 * means you burn ~2,650. Rounded to 10 kcal; null without enough data.
 */
export function observedMaintenance(intake: IntakeSummary, lbPerWeek: number | null): number | null {
    if (intake.avgCalories == null || intake.loggedDays < MIN_LOGGED_DAYS || lbPerWeek == null) return null;
    const kcal = intake.avgCalories - (lbPerWeek * KCAL_PER_LB) / 7;
    return Math.round(kcal / 10) * 10;
}

const KCAL_PER_G: Record<MacroKey, number> = { protein: 4, carbs: 4, fat: 9 };

export type SplitSegment = { key: MacroKey; grams: number; kcal: number; pct: number };

/** Protein, carbs, fat as shares of the calories the macros add up to (so the bar always fills). */
export function splitSegments(t: Pick<Targets, MacroKey>): SplitSegment[] {
    const keys: MacroKey[] = ["protein", "carbs", "fat"];
    const kcal = keys.map((k) => Math.max(0, t[k]) * KCAL_PER_G[k]);
    const total = kcal.reduce((a, b) => a + b, 0);
    return keys.map((key, i) => ({
        key,
        grams: Math.max(0, t[key]),
        kcal: kcal[i]!,
        pct: total > 0 ? (kcal[i]! / total) * 100 : 0,
    }));
}

/** Grams per lb of body weight, 2 decimals. Null without a weight. */
export function gramsPerLb(grams: number, weightLbs: number): number | null {
    if (weightLbs <= 0) return null;
    return Math.round((grams / weightLbs) * 100) / 100;
}

export function targetsEqual(a: Targets, b: Targets): boolean {
    return (
        Math.round(a.calories) === Math.round(b.calories) &&
        Math.round(a.protein) === Math.round(b.protein) &&
        Math.round(a.carbs) === Math.round(b.carbs) &&
        Math.round(a.fat) === Math.round(b.fat) &&
        Math.round(a.fiber) === Math.round(b.fiber)
    );
}

/** Macro kcal minus the calorie target, worded: "matches target", "30 under", "45 over". */
export function calorieGapLabel(delta: number): string {
    if (delta === 0) return "matches target";
    return delta < 0 ? `${fmtInt(-delta)} under` : `${fmtInt(delta)} over`;
}

export function fmtInt(n: number): string {
    return Math.round(n).toLocaleString("en-US");
}

/** "↓0.75 lb/wk", "↑0.2 lb/wk", "steady". */
export function fmtTrend(lbPerWeek: number): string {
    const r = Math.round(lbPerWeek * 100) / 100;
    if (Math.abs(r) < 0.05) return "steady";
    return `${r < 0 ? "↓" : "↑"}${Math.abs(r)} lb/wk`;
}

/** "Oct 20" from an ISO date. Empty for missing/invalid. */
export function fmtSince(iso: string | undefined | null): string {
    if (!iso) return "";
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return "";
    return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

/** Plans newest first by effective date (undated last). */
export function sortPlansByEffective<T extends { ID: number; effective_from?: string }>(plans: T[]): T[] {
    const t = (p: T) => (p.effective_from ? new Date(p.effective_from).getTime() : -Infinity);
    return [...plans].sort((a, b) => t(b) - t(a) || b.ID - a.ID);
}

/**
 * History rows: newest first, one per effective day. Each save creates a new plan dated today, so
 * several saves in a day leave several rows; only the last one (highest ID) was ever in effect.
 */
export function planHistory<T extends { ID: number; effective_from?: string }>(plans: T[]): T[] {
    const seen = new Set<string>();
    return sortPlansByEffective(plans).filter((p) => {
        const key = p.effective_from ? new Date(p.effective_from).toDateString() : `undated-${p.ID}`;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
    });
}
