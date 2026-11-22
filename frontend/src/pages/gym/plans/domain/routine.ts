import type { WorkoutPlan, WorkoutProgram } from "~/types/workout";
import { DAY_NAMES, DAY_SHORT, assignedDays, plural } from "./week";

/**
 * The routine and its program for a route. `programId` may be null (old /gym/plans/:id links) or
 * wrong (the routine moved): the routine id wins and the program is whichever holds it.
 */
export function findRoutine(
    programs: WorkoutProgram[],
    programId: number | null,
    planId: number | null,
): { program: WorkoutProgram; plan: WorkoutPlan } | null {
    if (planId === null) return null;
    const ordered = programId === null ? programs : [...programs].sort((a, b) => Number(b.ID === programId) - Number(a.ID === programId));
    for (const program of ordered) {
        const plan = program.plans?.find((p) => p.ID === planId);
        if (plan) return { program, plan };
    }
    return null;
}

/** Route params are strings; anything that isn't a positive integer is "no id". */
export function parseId(raw: unknown): number | null {
    if (typeof raw !== "string" || !/^\d+$/.test(raw)) return null;
    const n = Number.parseInt(raw, 10);
    return n > 0 ? n : null;
}

export type DayToggle = {
    dow: number;
    short: string;
    name: string;
    on: boolean;
    /** Another routine in the same program that has this day, if any. */
    takenBy: WorkoutPlan | null;
};

/** The seven day toggles for a routine, Sunday first like the week board and /gym. */
export function dayToggles(plan: WorkoutPlan, programPlans: WorkoutPlan[]): DayToggle[] {
    const mine = new Set(assignedDays(plan));
    return [0, 1, 2, 3, 4, 5, 6].map((dow) => ({
        dow,
        short: DAY_SHORT[dow] ?? "",
        name: DAY_NAMES[dow] ?? "",
        on: mine.has(dow),
        takenBy: programPlans.find((p) => p.ID !== plan.ID && assignedDays(p).includes(dow)) ?? null,
    }));
}

/** "Mon, Tue" or "No days yet". */
export function formatDays(toggles: DayToggle[]): string {
    const on = toggles.filter((t) => t.on).map((t) => t.short);
    return on.length ? on.join(", ") : "No days yet";
}

/** Other routines' days, for context under the toggles: "Wed Pull · Fri Legs". */
export function otherDaysLine(toggles: DayToggle[]): string {
    return toggles
        .filter((t) => t.takenBy)
        .map((t) => `${t.short} ${t.takenBy?.name ?? ""}`)
        .join(" · ");
}

/** "Push Pull Legs · 4 exercises". */
export function routineMeta(program: WorkoutProgram, plan: WorkoutPlan): string {
    return [program.name, plural(plan.exercises.length, "exercise")].join(" · ");
}

/** Move one item; out-of-range moves return the list unchanged (same reference). */
export function moveItem<T>(list: T[], from: number, to: number): T[] {
    if (from === to || from < 0 || to < 0 || from >= list.length || to >= list.length) return list;
    const next = [...list];
    const [item] = next.splice(from, 1);
    if (item !== undefined) next.splice(to, 0, item);
    return next;
}

/** Add a warmup/stretch item: trimmed, ignored when blank or already there (any case). */
export function addListItem(list: string[], raw: string): string[] {
    const item = raw.trim().replace(/\s+/g, " ");
    if (!item || list.some((x) => x.toLowerCase() === item.toLowerCase())) return list;
    return [...list, item];
}

/** Cardio as typed: type trimmed, minutes a whole number ≥ 0. */
export function normaliseCardio(type: string, minutes: number | null | undefined): { type: string; minutes: number } {
    return { type: type.trim(), minutes: Math.max(0, Math.round(Number(minutes) || 0)) };
}

export function cardioChanged(plan: WorkoutPlan, next: { type: string; minutes: number }): boolean {
    return (plan.planned_cardio_type?.trim() ?? "") !== next.type || (plan.planned_cardio_minutes ?? 0) !== next.minutes;
}
