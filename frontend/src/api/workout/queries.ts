import { infiniteQueryOptions, queryOptions, useInfiniteQuery, useMutation, useMutationState, useQuery, useQueryClient, type QueryClient } from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import axios from 'axios';
import { apiDELETE, apiGET, apiPATCH, apiPOST } from '~/api/client';
import { isClientError } from '~/api/queryClient';
import { useOptimisticMutation, usePendingIds, useServerMutation } from '~/api/optimistic';
import { activateWorkoutProgram, createWorkoutPlan, createWorkoutProgram, getWorkoutPrograms, renameWorkoutProgram, switchWorkoutPlan, type WorkoutActivityResponse, WorkoutLogsPreviousResponse } from '~/api/workout/api';
import { dayParams, offsetToDate } from '~/api/workout/dates';
import { workoutKeys } from '~/api/workout/keys';
import {
    addExerciseGroup,
    applyAddedExercise,
    applyCardioResponse,
    applyMobilityResponse,
    removeExerciseGroup,
    restoreCardio,
    restoreExerciseGroup,
    restoreMobility,
    setCardio,
    setMobilityChecked,
    type AddExerciseVars,
    type CardioVars,
    type MobilitySlot,
    type MobilityVars,
    type RemoveExerciseVars,
} from '~/api/workout/dayCache';
import {
    appendSet,
    dropRemovedExercise,
    findGroup,
    mergeLoggedExercise,
    patchSet,
    removeSetByKey,
    restoreSet,
    setNotes,
    setRef,
    type DeleteSetResponse,
    type DeleteSetVars,
    type LocalSetIntents,
    type LogSetResponse,
    type LogSetVars,
    type SaveNotesVars,
    type SetIntentState,
    type UpdateSetResponse,
    type UpdateSetVars,
} from '~/api/workout/setsCache';
import type { Cardio, Exercise, LoggedExercise, MobilityLogged, WorkoutPlan } from '~/types/workout';

export type WorkoutActivityQueryOpts = {
    mode: 'year' | 'rolling';
    weeks: number | null;
    days: number | null;
};

export function normalizeExercisesListPayload(value: unknown): Exercise[] {
    if (!value) return [];
    if (Array.isArray(value)) return value as Exercise[];
    if (
        typeof value === 'object' &&
        value !== null &&
        'exercises' in value &&
        Array.isArray((value as { exercises: unknown }).exercises)
    ) {
        return (value as { exercises: Exercise[] }).exercises;
    }
    const firstArray = Object.values(value as object).find((v) => Array.isArray(v));
    return (firstArray as Exercise[]) ?? [];
}

export async function fetchWorkoutExercisesAll(): Promise<unknown> {
    return apiGET<unknown>('/workout/exercises/all');
}

type ExercisesPageResponse = {
    exercises: Exercise[];
    has_next: boolean;
    total: number;
    page: number;
    page_size: number;
};

const STALE_HOT = 1000 * 60 * 2;
const STALE_COLD = 1000 * 60 * 10;

function activityParams(o: WorkoutActivityQueryOpts) {
    const params: Record<string, string | number> = { mode: o.mode };
    if (o.mode === 'rolling') {
        if (o.days != null && o.days > 0) {
            params.days = o.days;
        } else if (o.weeks != null && o.weeks > 0) {
            params.weeks = o.weeks;
        }
    }
    return params;
}

/** Single source for hooks, prefetch and `ensureQueryData`. Keys are absolute dates. */
export const workoutQueries = {
    /** The offset is derived from the date when the request is sent, so it never goes stale. */
    day: (date: string) =>
        queryOptions({
            queryKey: workoutKeys.days.day(date),
            queryFn: () =>
                apiGET<WorkoutLogsPreviousResponse>('/workout/logs/previous', {
                    params: dayParams(date),
                }),
            staleTime: STALE_HOT,
        }),
    activity: (o: WorkoutActivityQueryOpts) =>
        queryOptions({
            queryKey: workoutKeys.activity.list(o),
            queryFn: () =>
                apiGET<WorkoutActivityResponse>('/workout/logs/activity', { params: activityParams(o) }),
            staleTime: STALE_HOT,
        }),
    exercisesAll: () =>
        queryOptions({
            queryKey: workoutKeys.exercises.list(),
            queryFn: fetchWorkoutExercisesAll,
        }),
    exercisesPaginated: (search: string, pageSize = 10) =>
        infiniteQueryOptions({
            queryKey: workoutKeys.exercises.paginated(search),
            queryFn: async ({ pageParam }: { pageParam: number }) => {
                const params: Record<string, string | number> = { page: pageParam, page_size: pageSize };
                if (search) params.search = search;
                return apiGET<ExercisesPageResponse>('/workout/exercises/all', { params });
            },
            getNextPageParam: (lastPage: ExercisesPageResponse) =>
                lastPage.has_next ? lastPage.page + 1 : undefined,
            initialPageParam: 1,
            staleTime: STALE_COLD,
        }),
};

export function useWorkoutExercisesAllQuery(enabled: MaybeRefOrGetter<boolean> = true) {
    return useQuery(
        computed(() => ({ ...workoutQueries.exercisesAll(), enabled: toValue(enabled) })),
    );
}

export function useWorkoutActivity(opts: MaybeRefOrGetter<WorkoutActivityQueryOpts>) {
    return useQuery(computed(() => workoutQueries.activity(toValue(opts))));
}

/** Offset in, date key out: the offset is converted here, at the edge. */
export function useWorkoutLogsPrevious(offset: MaybeRefOrGetter<number> = 0) {
    return useQuery(computed(() => workoutQueries.day(offsetToDate(toValue(offset)))));
}

/** Same cache entry as `useWorkoutLogsPrevious`, narrowed with `select`. */
export function useWorkoutLogToday(
    offset: MaybeRefOrGetter<number> = 0,
    enabled: MaybeRefOrGetter<boolean> = true,
) {
    return useQuery(
        computed(() => ({
            ...workoutQueries.day(offsetToDate(toValue(offset))),
            select: (r: WorkoutLogsPreviousResponse) => r.day,
            enabled: toValue(enabled),
        })),
    );
}

export function useWorkoutPlansAll() {
    return useQuery({
        queryKey: workoutKeys.plans.list(),
        queryFn: () => apiGET<{ plans: WorkoutPlan[] }>('/workout/plans/all'),
        staleTime: STALE_COLD,
    });
}

export function useWorkoutPrograms() {
    return useQuery({
        queryKey: workoutKeys.programs,
        queryFn: getWorkoutPrograms,
        staleTime: STALE_COLD,
    });
}

/** The pages catch these and toast in context, so the global mutation toast stays quiet. */
const CALLER_TOASTS = { handled: true } as const;

function invalidateWorkoutPrograms(queryClient: ReturnType<typeof useQueryClient>) {
    queryClient.invalidateQueries({ queryKey: workoutKeys.programs });
    queryClient.invalidateQueries({ queryKey: workoutKeys.plans.all });
}

export function useCreateWorkoutProgram() {
    const queryClient = useQueryClient();
    return useMutation({
        meta: CALLER_TOASTS,
        mutationFn: (name: string) => createWorkoutProgram(name),
        onSuccess: () => invalidateWorkoutPrograms(queryClient),
    });
}

export function useRenameWorkoutProgram() {
    const queryClient = useQueryClient();
    return useMutation({
        meta: CALLER_TOASTS,
        mutationFn: ({ id, name }: { id: number; name: string }) => renameWorkoutProgram(id, name),
        onSuccess: () => invalidateWorkoutPrograms(queryClient),
    });
}

export function useActivateWorkoutProgram() {
    const queryClient = useQueryClient();
    return useMutation({
        meta: CALLER_TOASTS,
        mutationFn: (id: number) => activateWorkoutProgram(id),
        onSuccess: () => {
            invalidateWorkoutPrograms(queryClient);
            queryClient.invalidateQueries({ queryKey: workoutKeys.days.all });
        },
    });
}

export function useCreateWorkoutPlan() {
    const queryClient = useQueryClient();
    return useMutation({
        meta: CALLER_TOASTS,
        mutationFn: ({ programId, name, dayOfWeek }: { programId: number; name: string; dayOfWeek: number | null }) =>
            createWorkoutPlan(programId, name, dayOfWeek),
        onSuccess: () => invalidateWorkoutPrograms(queryClient),
    });
}

export function useExercisesPaginated(search: MaybeRefOrGetter<string> = '', pageSize = 10) {
    return useInfiniteQuery(computed(() => workoutQueries.exercisesPaginated(toValue(search), pageSize)));
}

// ── Per-set logging (optimistic-ui G2, endpoints from G1) ─────────────────────────────────────────

/**
 * All writes to one exercise's sets on one day (create, edit, delete, notes) run one after another.
 * An edit or delete of a set that is still being created waits for the create, and the notes save
 * (a full replace on the legacy endpoint) never races a set create. Cache patches apply at once.
 */
const setsScope = (v: { date: string; exercise_id: number }) => ({ id: `sets:${v.date}:${v.exercise_id}` });

type SetMutationVars = Partial<LogSetVars & UpdateSetVars & SaveNotesVars> & { date?: string };

/**
 * Set writes are idempotent, so transient failures retry twice. A 4xx (set deleted elsewhere,
 * client_id conflict, bad input) won't change on retry, and backing off would hold up every later
 * write in the exercise's scope, so it fails at once.
 */
export const retrySetWrite = (failureCount: number, error: unknown) => failureCount < 2 && !isClientError(error);

/**
 * Forget failed ('mark') creates/edits of a set once it is deleted, so `deriveSetRows` doesn't
 * rebuild the row from their variables. Done only after the delete succeeds: if it fails, the row is
 * restored and must still show "Didn't save" with Retry (it was never saved).
 */
function dismissFailedSetWrites(qc: QueryClient, v: DeleteSetVars) {
    const cache = qc.getMutationCache();
    for (const m of cache.findAll({ mutationKey: workoutKeys.mutations.sets, status: 'error' })) {
        const kind = m.options.mutationKey?.[3];
        const mv = m.state.variables as SetMutationVars | undefined;
        if ((kind === 'create' || kind === 'update') && mv?.client_id === v.client_id && mv.date === v.date) cache.remove(m);
    }
}

/**
 * What other set mutations still own on this day, so a reconcile doesn't overwrite it. `self` is
 * the reconciling call's variables (TanStack keeps the object identity), excluded from the scan.
 */
function localSetIntents(qc: QueryClient, date: string, exerciseId: number, self: unknown): LocalSetIntents {
    const keepFields = new Set<string>();
    const deleting = new Set<string>();
    let keepNotes = false;
    for (const m of qc.getMutationCache().findAll({ mutationKey: workoutKeys.mutations.sets })) {
        const v = m.state.variables as SetMutationVars | undefined;
        if (!v || v === self || v.date !== date) continue;
        const kind = m.options.mutationKey?.[3];
        const pending = m.state.status === 'pending';
        const failed = m.state.status === 'error';
        if ((kind === 'create' || kind === 'update') && (pending || failed) && v.client_id) keepFields.add(v.client_id);
        if (kind === 'delete' && pending && v.client_id) deleting.add(v.client_id);
        if (kind === 'notes' && pending && v.exercise_id === exerciseId) keepNotes = true;
    }
    return { keepFields, deleting, keepNotes };
}

function reconcileExercise(qc: QueryClient, v: { date: string; exercise_id: number }, exercise: LoggedExercise) {
    const intents = localSetIntents(qc, v.date, v.exercise_id, v);
    qc.setQueryData<WorkoutLogsPreviousResponse>(workoutKeys.days.day(v.date), (day) =>
        day ? mergeLoggedExercise(day, v.exercise_id, exercise, intents) : day,
    );
}

/**
 * Log one set. Appears at once; on failure the row stays, marked, with an inline retry ('mark').
 * Idempotent by `client_id`, so automatic retries and Retry never duplicate it. The activity
 * heatmap is invalidated only when this was (or, on a replay, may have been) the day's first set.
 */
export function useLogSet() {
    const qc = useQueryClient();
    return useOptimisticMutation<LogSetVars, LogSetResponse>({
        mutationKey: workoutKeys.mutations.logSet(),
        mutationFn: (v) =>
            apiPOST<LogSetResponse>('/workout/sets', {
                client_id: v.client_id,
                date: v.date,
                exercise_id: v.exercise_id,
                reps: v.reps,
                weight: v.weight,
                weight_setup: v.weight_setup,
            }),
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: WorkoutLogsPreviousResponse) => appendSet(day, v),
                revert: (day: WorkoutLogsPreviousResponse) => removeSetByKey(day, v),
            },
        ],
        reconcile: (res, v) => reconcileExercise(qc, v, res.exercise),
        // A replay (`created: false`) can be the day's first set whose first answer was lost; the
        // server reports `first_set_of_day: false` then, so refresh the heatmap to be sure.
        invalidate: (_v, res) => (res && (res.first_set_of_day || !res.created) ? [workoutKeys.activity.all] : []),
        scope: setsScope,
        onFailure: 'mark',
        retry: retrySetWrite,
        queryClient: qc,
    });
}

/** Edit a set (by key). Marked on failure like a create; PATCH is idempotent. */
export function useUpdateSet() {
    const qc = useQueryClient();
    return useOptimisticMutation<UpdateSetVars, UpdateSetResponse>({
        mutationKey: workoutKeys.mutations.updateSet(),
        mutationFn: (v) =>
            apiPATCH<UpdateSetResponse>(`/workout/sets/${encodeURIComponent(setRef(v.client_id))}`, {
                reps: v.reps,
                weight: v.weight,
                weight_setup: v.weight_setup,
            }),
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: WorkoutLogsPreviousResponse) => patchSet(day, v),
            },
        ],
        reconcile: (res, v) => reconcileExercise(qc, v, res.exercise),
        scope: setsScope,
        onFailure: 'mark',
        retry: retrySetWrite,
        queryClient: qc,
    });
}

/**
 * Delete a set (by key). Removed at once; on failure it goes back where it was, with a Retry toast.
 * The server answers 200 for an already-deleted (or never-saved) set, so retries are safe, and an
 * unsaved row is deleted on the server too in case its failed create did commit. Failed writes of the
 * set are dismissed only on success, so a failed delete leaves an unsaved row marked as unsaved.
 */
export function useDeleteSet() {
    const qc = useQueryClient();
    return useOptimisticMutation<DeleteSetVars, DeleteSetResponse>({
        mutationKey: workoutKeys.mutations.deleteSet(),
        mutationFn: (v) => apiDELETE<DeleteSetResponse>(`/workout/sets/${encodeURIComponent(setRef(v.client_id))}`),
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: WorkoutLogsPreviousResponse) => removeSetByKey(day, v),
                revert: (day: WorkoutLogsPreviousResponse, _v, snapshot: WorkoutLogsPreviousResponse) =>
                    restoreSet(day, v, snapshot),
            },
        ],
        reconcile: (res, v) => {
            dismissFailedSetWrites(qc, v);
            if (!res.exercise_removed) return;
            qc.setQueryData<WorkoutLogsPreviousResponse>(workoutKeys.days.day(v.date), (day) =>
                day ? dropRemovedExercise(day, v.exercise_id) : day,
            );
        },
        // The day may have lost its last log; only then can the heatmap change.
        invalidate: (_v, res) => (res?.exercise_removed ? [workoutKeys.activity.all] : []),
        scope: setsScope,
        errorMessage: () => 'Couldn’t delete the set',
        retry: retrySetWrite,
        queryClient: qc,
    });
}

/**
 * Save an exercise's notes. There is no notes endpoint yet, so this uses the legacy full replace
 * (`POST /workout/exercises/log`, type "logged"), which deletes every stored set missing from the
 * payload. The payload's sets are therefore the server's own, fetched at send time, never the
 * cache's: the cache can lack a set the server has (a create whose answer was lost, a set logged
 * from another device) or hold values the server doesn't (a failed edit). It runs in the exercise's
 * scope, so every earlier set write of this tab has settled by then.
 */
export function useSaveExerciseNotes() {
    const qc = useQueryClient();
    return useOptimisticMutation<SaveNotesVars, LoggedExercise>({
        mutationKey: workoutKeys.mutations.notes(),
        mutationFn: async (v) => {
            const day = await apiGET<WorkoutLogsPreviousResponse>('/workout/logs/previous', { params: dayParams(v.date) });
            const logged = findGroup(day, v.exercise_id)?.logged;
            if (!logged || !(logged.ID > 0)) throw new Error('Log a set before saving notes');
            const body = await apiPOST<{ exercise: LoggedExercise }>('/workout/exercises/log', {
                type: 'logged',
                exercise: {
                    ...logged,
                    notes: v.notes,
                    sets: logged.sets.map(({ client_id: _c, ...s }) => s),
                },
            });
            return body.exercise;
        },
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: WorkoutLogsPreviousResponse) => setNotes(day, v),
                revert: (day: WorkoutLogsPreviousResponse, _v, snapshot: WorkoutLogsPreviousResponse) =>
                    setNotes(day, { exercise_id: v.exercise_id, notes: findGroup(snapshot, v.exercise_id)?.logged?.notes ?? '' }),
            },
        ],
        reconcile: (exercise, v) => reconcileExercise(qc, v, exercise),
        scope: setsScope,
        errorMessage: () => 'Couldn’t save notes',
        queryClient: qc,
    });
}

/**
 * Pending and failed set writes (all days), for `deriveSetRows`. Status comes from `usePendingIds`;
 * the variables of unsaved creates/edits come from `useMutationState`, so a row survives a refetch
 * that drops it from the cache.
 */
export function useSetIntents() {
    const qc = useQueryClient();
    const k = workoutKeys.mutations;
    const pendingCreate = usePendingIds(k.logSet(), 'pending', qc);
    const pendingUpdate = usePendingIds(k.updateSet(), 'pending', qc);
    const pendingDelete = usePendingIds(k.deleteSet(), 'pending', qc);
    const failedCreate = usePendingIds(k.logSet(), 'error', qc);
    const failedUpdate = usePendingIds(k.updateSet(), 'error', qc);
    const unsaved = useMutationState(
        {
            filters: {
                mutationKey: k.sets,
                predicate: (m) =>
                    (m.state.status === 'pending' || m.state.status === 'error') &&
                    (m.options.mutationKey?.[3] === 'create' || m.options.mutationKey?.[3] === 'update'),
            },
            select: (m) => ({ kind: m.options.mutationKey?.[3], vars: m.state.variables }),
        },
        qc,
    );
    return computed<SetIntentState>(() => ({
        pending: new Set([...pendingCreate.value, ...pendingUpdate.value, ...pendingDelete.value]),
        deleting: pendingDelete.value,
        failedCreate: failedCreate.value,
        failedUpdate: failedUpdate.value,
        creates: unsaved.value.filter((u) => u.kind === 'create').map((u) => u.vars as LogSetVars),
        updates: unsaved.value.filter((u) => u.kind === 'update').map((u) => u.vars as UpdateSetVars),
    }));
}

/** Variables of a failed set write, so Retry/Delete can re-run or dismiss exactly that call. */
export function findFailedSetVars(
    qc: QueryClient,
    key: string,
): { kind: 'create'; vars: LogSetVars } | { kind: 'update'; vars: UpdateSetVars } | null {
    const failed = qc
        .getMutationCache()
        .findAll({ mutationKey: workoutKeys.mutations.sets, status: 'error' })
        .filter((m) => (m.state.variables as { client_id?: string } | undefined)?.client_id === key);
    const last = failed[failed.length - 1];
    if (!last) return null;
    const kind = last.options.mutationKey?.[3];
    if (kind === 'create') return { kind, vars: last.state.variables as LogSetVars };
    if (kind === 'update') return { kind, vars: last.state.variables as UpdateSetVars };
    return null;
}

/** A notes save for this exercise is already in flight with these notes (avoid re-sending). */
export function notesSavePending(qc: QueryClient, v: SaveNotesVars): boolean {
    return qc
        .getMutationCache()
        .findAll({ mutationKey: workoutKeys.mutations.notes(), status: 'pending' })
        .some((m) => {
            const mv = m.state.variables as SaveNotesVars | undefined;
            return mv?.date === v.date && mv.exercise_id === v.exercise_id && mv.notes === v.notes;
        });
}

// ── Rest of logging (optimistic-ui G3): add/remove exercise, cardio, mobility, switch plan ───────

type Day = WorkoutLogsPreviousResponse;

/** Heatmap membership can change with any of these writes; it is a different key from the day. */
const activityKeys = () => [workoutKeys.activity.all];

/**
 * Add an exercise to the day. The group appears at once (needs `exercise` for its name; without it the
 * day is refetched on settle). Shares the exercise's set scope so a first set can't overtake it. The
 * response's saved row replaces the temporary one; `previous`/`max` are server-computed, so the day is
 * refetched once after settle (the single GET this slice keeps).
 */
export function useAddExerciseToWorkout() {
    const qc = useQueryClient();
    return useOptimisticMutation<AddExerciseVars, LoggedExercise>({
        mutationKey: workoutKeys.mutations.addExercise(),
        mutationFn: async (v) => {
            const body = await apiPOST<{ exercise: LoggedExercise }>(
                '/workout/exercises/add',
                { exercise_id: v.exerciseId },
                { params: dayParams(v.date) },
            );
            return body.exercise;
        },
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: Day, vars: AddExerciseVars) => addExerciseGroup(day, vars),
                revert: (day: Day, vars: AddExerciseVars, snapshot: Day) => removeAddedGroup(day, snapshot, vars),
            },
        ],
        reconcile: (logged, v) =>
            qc.setQueryData<Day>(workoutKeys.days.day(v.date), (day) => (day ? applyAddedExercise(day, v.exerciseId, logged) : day)),
        invalidate: (v) => [workoutKeys.days.day(v.date), ...activityKeys()],
        scope: (v) => setsScope({ date: v.date, exercise_id: v.exerciseId }),
        errorMessage: () => 'Couldn’t add the exercise',
        queryClient: qc,
    });
}

/** Undo an add when siblings changed the day: drop the group again unless it was already there. */
function removeAddedGroup(day: Day, snapshot: Day, v: AddExerciseVars): Day {
    return findGroup(snapshot, v.exerciseId)?.logged ? day : removeExerciseGroup(day, v.exerciseId);
}

/**
 * Remove an exercise's log from the day. It disappears at once; a 404 (nothing was logged) counts as
 * success, as before. Rollback restores the row; the day is refetched on settle.
 */
export function useRemoveExerciseFromWorkout() {
    const qc = useQueryClient();
    return useOptimisticMutation<RemoveExerciseVars, { success?: boolean }>({
        mutationKey: workoutKeys.mutations.removeExercise(),
        mutationFn: async (v) => {
            try {
                return await apiDELETE<{ success?: boolean }>('/workout/exercises/remove', {
                    data: { exercise_id: v.exerciseId },
                    params: dayParams(v.date),
                });
            } catch (err) {
                if (axios.isAxiosError(err) && err.response?.status === 404) return { success: true };
                throw err;
            }
        },
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: Day, vars: RemoveExerciseVars) => removeExerciseGroup(day, vars.exerciseId),
                revert: (day: Day, vars: RemoveExerciseVars, snapshot: Day) => restoreExerciseGroup(day, snapshot, vars.exerciseId),
            },
        ],
        // Removing a planned exercise's log makes the server rebuild it (previous/max), so refetch the day.
        invalidate: (v) => [workoutKeys.days.day(v.date), ...activityKeys()],
        scope: (v) => setsScope({ date: v.date, exercise_id: v.exerciseId }),
        errorMessage: () => 'Couldn’t remove the exercise',
        retry: 1,
        queryClient: qc,
    });
}

function useUpsertMobility(slot: MobilitySlot) {
    const qc = useQueryClient();
    const key = slot === 'pre' ? workoutKeys.mutations.mobilityPre() : workoutKeys.mutations.mobilityPost();
    return useOptimisticMutation<Omit<MobilityVars, 'slot'>, MobilityLogged>({
        mutationKey: key,
        mutationFn: async (v) => {
            const body = await apiPOST<{ mobility: MobilityLogged }>(
                `/workout/logs/mobility/${slot}`,
                { checked: v.checked },
                { params: dayParams(v.date) },
            );
            return body.mobility;
        },
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: Day) => setMobilityChecked(day, slot, v.checked),
                revert: (day: Day, _v, snapshot: Day) => restoreMobility(day, snapshot, slot),
            },
        ],
        // The next tick has already patched the cache while an earlier call is in flight; let the last one win.
        reconcile: (mobility, v, client, { siblingsPending }) => {
            const dayKey = workoutKeys.days.day(v.date);
            if (siblingsPending(dayKey) > 0) return;
            client.setQueryData<Day>(dayKey, (day) => (day ? applyMobilityResponse(day, slot, mobility) : day));
        },
        invalidate: () => activityKeys(),
        scope: (v) => ({ id: `mobility:${v.date}:${slot}` }),
        errorMessage: () => 'Couldn’t save mobility',
        retry: 2,
        queryClient: qc,
    });
}

/** Tick pre-workout mobility items. Upsert by workout log, so retries are safe. */
export function useUpsertMobilityPre() {
    return useUpsertMobility('pre');
}

export function useUpsertMobilityPost() {
    return useUpsertMobility('post');
}

/** Save cardio. The logged cardio shows at once; the server upserts by workout log, so retries are safe. */
export function useUpsertCardio() {
    const qc = useQueryClient();
    return useOptimisticMutation<CardioVars, Cardio>({
        mutationKey: workoutKeys.mutations.cardio(),
        mutationFn: async (v) => {
            const body = await apiPOST<{ cardio: Cardio }>(
                '/workout/logs/cardio',
                {
                    minutes: v.minutes,
                    ...(v.type !== undefined && v.type !== '' ? { type: v.type } : {}),
                    ...(v.notes !== undefined ? { notes: v.notes } : {}),
                },
                { params: dayParams(v.date) },
            );
            return body.cardio;
        },
        optimistic: (v) => [
            {
                queryKey: workoutKeys.days.day(v.date),
                apply: (day: Day) => setCardio(day, v),
                revert: (day: Day, _v, snapshot: Day) => restoreCardio(day, snapshot),
            },
        ],
        reconcile: (cardio, v, client, { siblingsPending }) => {
            const dayKey = workoutKeys.days.day(v.date);
            if (siblingsPending(dayKey) > 0) return;
            client.setQueryData<Day>(dayKey, (day) => (day ? applyCardioResponse(day, cardio) : day));
        },
        invalidate: () => activityKeys(),
        scope: (v) => ({ id: `cardio:${v.date}` }),
        errorMessage: () => 'Couldn’t save cardio',
        retry: 2,
        queryClient: qc,
    });
}

/**
 * Switch the day's plan. Waits for the server (it rebuilds the planned exercises) and writes the full
 * day it returns into the cache, so no follow-up GET. If another write to the day is still pending the
 * response would erase its optimistic rows, so the day is only marked stale then (no refetch now).
 */
export function useSwitchWorkoutPlan() {
    const qc = useQueryClient();
    return useServerMutation<{ date: string; planId: number | null }, Day>({
        mutationKey: workoutKeys.mutations.switchPlan(),
        mutationFn: ({ date, planId }) => switchWorkoutPlan(date, planId),
        reconcile: (res, v, client, { siblingsPending }) => {
            const dayKey = workoutKeys.days.day(v.date);
            if (siblingsPending(dayKey) > 0) {
                void client.invalidateQueries({ queryKey: dayKey, exact: true, refetchType: 'none' });
                return;
            }
            client.setQueryData<Day>(dayKey, res);
        },
        invalidate: () => activityKeys(),
        errorMessage: () => 'Couldn’t switch the plan',
        queryClient: qc,
    });
}
