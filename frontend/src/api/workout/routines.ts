import { apiDELETE, apiPATCH, apiPOST, apiPUT } from "~/api/client";
import { useOptimisticMutation, useServerMutation, type CachePatch } from "~/api/optimistic";
import { workoutKeys } from "~/api/workout/keys";
import {
    ASSIGN_FIELDS,
    assignAffected,
    assignDay,
    patchRoutine,
    removeExercise,
    reorderExercises,
    restoreFields,
    unassignDay,
    type ProgramsData,
} from "~/api/workout/routinePatches";
import { searchListKeys } from "~/shared/searchListKeys";
import type { Exercise, ExerciseLoadType, WorkoutPlan } from "~/types/workout";

/**
 * Mutations for routines (a `WorkoutPlan` inside a program) and the exercises they hold.
 *
 * Optimistic edits patch the programs cache (`workoutKeys.programs`), roll back on error and
 * invalidate on settle. Routine edits change what today's workout is built from, so every settle
 * refreshes programs, the plans list and the day views (`days.all`). Requests for one routine are
 * serialised (`scope`) so rapid taps reach the server in order.
 */

const PROGRAMS = workoutKeys.programs;
const settleKeys = () => [PROGRAMS, workoutKeys.plans.all, workoutKeys.days.all];
const scope = (v: { planId: number }) => ({ id: `routine:${v.planId}` });

/** A patch of one routine in the programs cache whose revert restores `fields` from the snapshot. */
function routinePatch<V extends { planId: number }>(
    _vars: V,
    patch: (p: WorkoutPlan, v: V) => WorkoutPlan,
    fields: (keyof WorkoutPlan)[],
): CachePatch<V, ProgramsData> {
    return {
        queryKey: PROGRAMS,
        apply: (old, v) => patchRoutine(old, v.planId, (p) => patch(p, v)),
        revert: (cur, v, snap) => restoreFields(cur, snap, [v.planId], fields),
    };
}

export function useRenameRoutine() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "renameRoutine"],
        mutationFn: ({ planId, name }: { planId: number; name: string }) =>
            apiPATCH<{ plan: WorkoutPlan }>(`/workout/plans/${planId}`, { name }),
        optimistic: (v) => [routinePatch(v, (p) => ({ ...p, name: v.name }), ["name"])],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t rename the routine",
    });
}

/** Soft delete: the backend keeps the routine so `useRestoreRoutine` can undo it. Waits for the server. */
export function useDeleteRoutine() {
    return useServerMutation({
        mutationKey: ["workout", "mutations", "deleteRoutine"],
        mutationFn: (planId: number) => apiDELETE(`/workout/plans/${planId}`),
        invalidate: settleKeys,
        errorMessage: false,
    });
}

export function useRestoreRoutine() {
    return useServerMutation({
        mutationKey: ["workout", "mutations", "restoreRoutine"],
        mutationFn: (planId: number) => apiPOST<{ plan: WorkoutPlan }>(`/workout/plans/${planId}/restore`),
        invalidate: settleKeys,
        errorMessage: false,
    });
}

export function useReorderRoutineExercises() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "reorderRoutineExercises"],
        mutationFn: ({ planId, exerciseIds }: { planId: number; exerciseIds: number[] }) =>
            apiPUT(`/workout/plans/${planId}/exercises/reorder`, { exercise_ids: exerciseIds }),
        optimistic: (v) => [routinePatch(v, (p) => reorderExercises(p, v.exerciseIds), ["exercises"])],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t reorder the exercises",
    });
}

export function useRemoveRoutineExercise() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "removeRoutineExercise"],
        mutationFn: ({ planId, exerciseId }: { planId: number; exerciseId: number }) =>
            apiDELETE(`/workout/plans/${planId}/exercises/remove`, { data: { exercise_id: exerciseId } }),
        optimistic: (v) => [routinePatch(v, (p) => removeExercise(p, v.exerciseId), ["exercises"])],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t remove the exercise",
    });
}

/** Adds at the end of the routine (the backend appends). Needs the exercise row, so it waits for the server. */
export function useAddRoutineExercise() {
    return useServerMutation({
        mutationKey: ["workout", "mutations", "addRoutineExercise"],
        mutationFn: ({ planId, exerciseId }: { planId: number; exerciseId: number }) =>
            apiPOST(`/workout/plans/${planId}/exercises/add`, { exercise_id: exerciseId }),
        invalidate: settleKeys,
        scope,
        errorMessage: false,
    });
}

export function useSetRoutineMobility() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "setRoutineMobility"],
        mutationFn: ({ planId, pre, post }: { planId: number; pre: string[]; post: string[] }) =>
            apiPUT(`/workout/plans/${planId}/planned-mobility`, { pre_mobility_items: pre, post_mobility_items: post }),
        optimistic: (v) => [
            routinePatch(v, (p) => ({ ...p, pre_mobility_items: v.pre, post_mobility_items: v.post }), [
                "pre_mobility_items",
                "post_mobility_items",
            ]),
        ],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t save the mobility",
    });
}

export function useSetRoutineCardio() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "setRoutineCardio"],
        mutationFn: ({ planId, type, minutes }: { planId: number; type: string; minutes: number }) =>
            apiPUT(`/workout/plans/${planId}/planned-cardio`, { type, minutes }),
        optimistic: (v) => [
            routinePatch(v, (p) => ({ ...p, planned_cardio_type: v.type, planned_cardio_minutes: v.minutes }), [
                "planned_cardio_type",
                "planned_cardio_minutes",
            ]),
        ],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t save cardio",
    });
}

type DayVars = { planId: number; dayOfWeek: number };

/** Put a routine on a weekday. The backend clears whichever routine had that day in the same program. */
export function useAssignPlanDay() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "assignPlanDay"],
        mutationFn: ({ planId, dayOfWeek }: DayVars) =>
            apiPOST<{ plan: WorkoutPlan }>(`/workout/plans/${planId}/assign-day`, { day_of_week: dayOfWeek }),
        optimistic: () => [
            {
                queryKey: PROGRAMS,
                apply: (old: ProgramsData, x: DayVars) => assignDay(old, x.planId, x.dayOfWeek),
                revert: (cur: ProgramsData, x: DayVars, snap: ProgramsData) =>
                    restoreFields(cur, snap, assignAffected(snap, x.planId, x.dayOfWeek), ASSIGN_FIELDS),
            } satisfies CachePatch<DayVars, ProgramsData>,
        ],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t change the day",
    });
}

export function useUnassignPlanDay() {
    return useOptimisticMutation({
        mutationKey: ["workout", "mutations", "unassignPlanDay"],
        mutationFn: ({ planId, dayOfWeek }: DayVars) =>
            apiDELETE<{ plan: WorkoutPlan }>(`/workout/plans/${planId}/assign-day`, { params: { day_of_week: dayOfWeek } }),
        optimistic: () => [
            {
                queryKey: PROGRAMS,
                apply: (old: ProgramsData, x: DayVars) => unassignDay(old, x.planId, x.dayOfWeek),
                revert: (cur: ProgramsData, x: DayVars, snap: ProgramsData) =>
                    restoreFields(cur, snap, [x.planId], ASSIGN_FIELDS),
            } satisfies CachePatch<DayVars, ProgramsData>,
        ],
        invalidate: settleKeys,
        scope,
        errorMessage: () => "Couldn’t change the day",
    });
}

// ── Exercises (library) ─────────────────────────────────────────────────────────────────────────

/** Everything that embeds or lists exercises: library lists, search pickers, routines, today's day. */
export function exerciseKeys() {
    return [workoutKeys.exercises.all, searchListKeys.all, PROGRAMS, workoutKeys.plans.all, workoutKeys.days.all];
}

export type ExerciseInput = { name: string; rep_rollover: number; cues: string; load_type?: ExerciseLoadType };

/** Unique names fail often, so this waits for the server; callers show the error inline. */
export function useCreateExercise() {
    return useServerMutation({
        mutationKey: ["workout", "mutations", "createExercise"],
        mutationFn: async (input: ExerciseInput) =>
            (await apiPOST<{ exercise: Exercise }>("/workout/exercises", input)).exercise,
        invalidate: exerciseKeys,
        errorMessage: false,
    });
}

export function useUpdateExercise() {
    return useServerMutation({
        mutationKey: ["workout", "mutations", "updateExercise"],
        mutationFn: async ({ id, ...input }: ExerciseInput & { id: number }) =>
            (await apiPUT<{ exercise: Exercise }>(`/workout/exercises/${id}`, input)).exercise,
        invalidate: exerciseKeys,
        errorMessage: false,
    });
}
