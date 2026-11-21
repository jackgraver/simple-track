import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiDELETE, apiPATCH, apiPOST, apiPUT } from "~/api/client";
import { liveworkoutKeys, workoutKeys } from "~/api/workout/keys";
import type { WorkoutPlan, WorkoutProgram } from "~/types/workout";

/**
 * Mutations for one routine (a `WorkoutPlan` inside a program), used by the routine page.
 * Edits that the page shows straight away (order, removal, mobility, cardio) patch the programs
 * cache first and roll back on error; the rest refetch on success.
 */

type ProgramsData = { programs: WorkoutProgram[] };
type QueryClient = ReturnType<typeof useQueryClient>;

function invalidate(qc: QueryClient, today = false) {
    void qc.invalidateQueries({ queryKey: workoutKeys.programs });
    void qc.invalidateQueries({ queryKey: workoutKeys.plans.all });
    if (today) void qc.invalidateQueries({ queryKey: liveworkoutKeys.workouts.all });
}

/** Apply `patch` to the routine in the cached programs list; returns a rollback. */
async function patchRoutine(qc: QueryClient, planId: number, patch: (p: WorkoutPlan) => WorkoutPlan) {
    await qc.cancelQueries({ queryKey: workoutKeys.programs });
    const before = qc.getQueryData<ProgramsData>(workoutKeys.programs);
    if (before) {
        qc.setQueryData<ProgramsData>(workoutKeys.programs, {
            programs: before.programs.map((prog) => ({
                ...prog,
                plans: prog.plans.map((p) => (p.ID === planId ? patch(p) : p)),
            })),
        });
    }
    return () => {
        if (before) qc.setQueryData(workoutKeys.programs, before);
    };
}

function optimistic<V extends { planId: number }>(
    qc: QueryClient,
    mutationFn: (v: V) => Promise<unknown>,
    patch: (p: WorkoutPlan, v: V) => WorkoutPlan,
    today = false,
) {
    return useMutation({
        mutationFn,
        onMutate: async (v: V) => ({ rollback: await patchRoutine(qc, v.planId, (p) => patch(p, v)) }),
        onError: (_err, _v, ctx) => ctx?.rollback(),
        onSettled: () => invalidate(qc, today),
    });
}

export function useRenameRoutine() {
    const qc = useQueryClient();
    return optimistic(
        qc,
        ({ planId, name }: { planId: number; name: string }) => apiPATCH<{ plan: WorkoutPlan }>(`/workout/plans/${planId}`, { name }),
        (p, { name }) => ({ ...p, name }),
        true,
    );
}

/** Soft delete: the backend keeps the routine so `useRestoreRoutine` can undo it. */
export function useDeleteRoutine() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (planId: number) => apiDELETE(`/workout/plans/${planId}`),
        onSettled: () => invalidate(qc, true),
    });
}

export function useRestoreRoutine() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (planId: number) => apiPOST<{ plan: WorkoutPlan }>(`/workout/plans/${planId}/restore`),
        onSettled: () => invalidate(qc, true),
    });
}

export function useReorderRoutineExercises() {
    const qc = useQueryClient();
    return optimistic(
        qc,
        ({ planId, exerciseIds }: { planId: number; exerciseIds: number[] }) =>
            apiPUT(`/workout/plans/${planId}/exercises/reorder`, { exercise_ids: exerciseIds }),
        (p, { exerciseIds }) => ({
            ...p,
            exercises: exerciseIds.flatMap((id) => p.exercises.filter((e) => e.ID === id)),
        }),
    );
}

export function useRemoveRoutineExercise() {
    const qc = useQueryClient();
    return optimistic(
        qc,
        ({ planId, exerciseId }: { planId: number; exerciseId: number }) =>
            apiDELETE(`/workout/plans/${planId}/exercises/remove`, { data: { exercise_id: exerciseId } }),
        (p, { exerciseId }) => ({ ...p, exercises: p.exercises.filter((e) => e.ID !== exerciseId) }),
    );
}

/** Adds at the end of the routine (the backend appends). */
export function useAddRoutineExercise() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ planId, exerciseId }: { planId: number; exerciseId: number }) =>
            apiPOST(`/workout/plans/${planId}/exercises/add`, { exercise_id: exerciseId }),
        onSettled: () => invalidate(qc),
    });
}

export function useSetRoutineMobility() {
    const qc = useQueryClient();
    return optimistic(
        qc,
        ({ planId, pre, post }: { planId: number; pre: string[]; post: string[] }) =>
            apiPUT(`/workout/plans/${planId}/planned-mobility`, { pre_mobility_items: pre, post_mobility_items: post }),
        (p, { pre, post }) => ({ ...p, pre_mobility_items: pre, post_mobility_items: post }),
        true,
    );
}

export function useSetRoutineCardio() {
    const qc = useQueryClient();
    return optimistic(
        qc,
        ({ planId, type, minutes }: { planId: number; type: string; minutes: number }) =>
            apiPUT(`/workout/plans/${planId}/planned-cardio`, { type, minutes }),
        (p, { type, minutes }) => ({ ...p, planned_cardio_type: type, planned_cardio_minutes: minutes }),
        true,
    );
}
