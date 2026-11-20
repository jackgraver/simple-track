import type { WorkoutPlan, WorkoutProgram } from "~/types/workout";

/** 0 = Sunday, matching `day_of_week` from the API and the /gym week strip. */
export const DAY_NAMES = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"] as const;
export const DAY_SHORT = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"] as const;

export type DaySlot = {
    dow: number;
    name: string;
    short: string;
    plan: WorkoutPlan | null;
    isToday: boolean;
};

/** Days a routine is on. Older rows only have `day_of_week`. */
export function assignedDays(plan: WorkoutPlan): number[] {
    if (plan.assigned_days?.length) return plan.assigned_days;
    return plan.day_of_week === null || plan.day_of_week === undefined ? [] : [plan.day_of_week];
}

/**
 * The program to show first: an active one that has routines, else any active one, else the first.
 * Several programs can come back active (see the backend note in the plans page), and /gym picks the
 * same way, so the two pages agree.
 */
export function pickDefaultProgram(programs: WorkoutProgram[]): WorkoutProgram | null {
    return (
        programs.find((p) => p.is_active && p.plans?.length) ??
        programs.find((p) => p.is_active) ??
        programs[0] ??
        null
    );
}

/** Seven slots, Sunday first. The first routine found wins a day (the backend keeps one per day). */
export function buildWeek(plans: WorkoutPlan[], todayDow: number): DaySlot[] {
    const byDay = new Map<number, WorkoutPlan>();
    for (const plan of plans) {
        for (const d of assignedDays(plan)) if (!byDay.has(d)) byDay.set(d, plan);
    }
    return DAY_NAMES.map((name, dow) => ({
        dow,
        name,
        short: DAY_SHORT[dow] ?? name.slice(0, 3),
        plan: byDay.get(dow) ?? null,
        isToday: dow === todayDow,
    }));
}

/** "Mon, Thu" in week order, or "Not on the week". A routine can sit on several days. */
export function daysLabel(plan: WorkoutPlan): string {
    const days = [...new Set(assignedDays(plan))].sort((a, b) => a - b);
    return days.length ? days.map((d) => DAY_SHORT[d]).join(", ") : "Not on the week";
}

export type WeekSummary = { sessions: number; exercises: number; restDays: number };

/** Exercises are counted per session, so a routine on two days counts twice. */
export function weekSummary(week: DaySlot[]): WeekSummary {
    const sessions = week.filter((s) => s.plan).length;
    const exercises = week.reduce((n, s) => n + (s.plan?.exercises.length ?? 0), 0);
    return { sessions, exercises, restDays: week.length - sessions };
}

export function plural(n: number, word: string): string {
    return `${n} ${word}${n === 1 ? "" : "s"}`;
}

export function formatWeekSummary(s: WeekSummary): string {
    if (s.sessions === 0) return "No training days scheduled yet";
    return [plural(s.sessions, "session"), plural(s.exercises, "exercise"), plural(s.restDays, "rest day")].join(" · ");
}

/** Short facts about a routine: "4 exercises", "Bike 15 min". */
export function planMeta(plan: WorkoutPlan): string[] {
    const parts = [plural(plan.exercises.length, "exercise")];
    const cardio = plan.planned_cardio_type?.trim();
    if (cardio) parts.push(plan.planned_cardio_minutes ? `${cardio} ${plan.planned_cardio_minutes} min` : cardio);
    return parts;
}
