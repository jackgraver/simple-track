import type { WorkoutPlan, WorkoutProgram } from "~/types/workout";

/**
 * Pure cache patches for routine (workout plan) edits, applied to the cached
 * `GET /workout/programs` response. Each has a matching `revert*` used when another mutation also
 * patched the entry and a snapshot restore would wipe its change.
 */
export type ProgramsData = { programs: WorkoutProgram[] };

/** Days a routine is on. Older rows only have `day_of_week`. */
function daysOf(p: WorkoutPlan): number[] {
    if (p.assigned_days?.length) return p.assigned_days;
    return p.day_of_week === null || p.day_of_week === undefined ? [] : [p.day_of_week];
}

function withDays(p: WorkoutPlan, days: number[]): WorkoutPlan {
    const sorted = [...new Set(days)].sort((a, b) => a - b);
    return { ...p, assigned_days: sorted, day_of_week: sorted[0] ?? null };
}

function mapPlans(data: ProgramsData, fn: (p: WorkoutPlan, program: WorkoutProgram) => WorkoutPlan): ProgramsData {
    return { ...data, programs: data.programs.map((prog) => ({ ...prog, plans: prog.plans.map((p) => fn(p, prog)) })) };
}

/** Patch one routine. */
export function patchRoutine(data: ProgramsData, planId: number, patch: (p: WorkoutPlan) => WorkoutPlan): ProgramsData {
    return mapPlans(data, (p) => (p.ID === planId ? patch(p) : p));
}

/** Put a routine on `dow`; the program's other routines lose that day (the backend does the same). */
export function assignDay(data: ProgramsData, planId: number, dow: number): ProgramsData {
    return mapPlans(data, (p, prog) => {
        if (!prog.plans.some((x) => x.ID === planId)) return p;
        if (p.ID === planId) return withDays(p, [...daysOf(p), dow]);
        return daysOf(p).includes(dow) ? withDays(p, daysOf(p).filter((d) => d !== dow)) : p;
    });
}

export function unassignDay(data: ProgramsData, planId: number, dow: number): ProgramsData {
    return patchRoutine(data, planId, (p) => withDays(p, daysOf(p).filter((d) => d !== dow)));
}

/** Copy `fields` of the routines `ids` from `snapshot` into `current` (leaves everything else alone). */
export function restoreFields(
    current: ProgramsData,
    snapshot: ProgramsData,
    ids: number[],
    fields: (keyof WorkoutPlan)[],
): ProgramsData {
    const before = new Map(snapshot.programs.flatMap((prog) => prog.plans).map((p) => [p.ID, p]));
    return mapPlans(current, (p) => {
        const old = before.get(p.ID);
        if (!old || !ids.includes(p.ID)) return p;
        const next = { ...p } as Record<string, unknown>;
        for (const f of fields) next[f] = old[f];
        return next as unknown as WorkoutPlan;
    });
}

/** Routines affected by assigning `dow` to `planId`: itself plus whoever held the day in its program. */
export function assignAffected(snapshot: ProgramsData, planId: number, dow: number): number[] {
    const prog = snapshot.programs.find((g) => g.plans.some((p) => p.ID === planId));
    return [planId, ...(prog?.plans ?? []).filter((p) => daysOf(p).includes(dow)).map((p) => p.ID)];
}

export const ASSIGN_FIELDS: (keyof WorkoutPlan)[] = ["assigned_days", "day_of_week"];

/** New order for `exerciseIds`; ids not in the routine are ignored, exercises not listed are dropped
 *  (the request replaces the order, as the backend does). */
export function reorderExercises(p: WorkoutPlan, exerciseIds: number[]): WorkoutPlan {
    return { ...p, exercises: exerciseIds.flatMap((id) => p.exercises.filter((e) => e.ID === id)) };
}

export function removeExercise(p: WorkoutPlan, exerciseId: number): WorkoutPlan {
    return { ...p, exercises: p.exercises.filter((e) => e.ID !== exerciseId) };
}
