import { infiniteQueryOptions, queryOptions, useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { apiDELETE, apiGET, apiPOST } from '~/api/client';
import { activateWorkoutProgram, createWorkoutPlan, createWorkoutProgram, getWorkoutPrograms, renameWorkoutProgram, switchWorkoutPlan, type WorkoutActivityResponse, WorkoutLogsPreviousResponse } from '~/api/workout/api';
import { dateToOffset, offsetToDate } from '~/api/workout/dates';
import { workoutKeys } from '~/api/workout/keys';
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
                    params: { offset: dateToOffset(date) },
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

function invalidateWorkoutActivityQueries(queryClient: ReturnType<typeof useQueryClient>) {
    queryClient.invalidateQueries({ queryKey: workoutKeys.activity.all });
}

/** After a write to one day: that day's view plus the activity heatmap. */
function invalidateDay(queryClient: ReturnType<typeof useQueryClient>, date: string) {
    queryClient.invalidateQueries({ queryKey: workoutKeys.days.day(date) });
    invalidateWorkoutActivityQueries(queryClient);
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

function invalidateWorkoutPrograms(queryClient: ReturnType<typeof useQueryClient>) {
    queryClient.invalidateQueries({ queryKey: workoutKeys.programs });
    queryClient.invalidateQueries({ queryKey: workoutKeys.plans.all });
}

export function useCreateWorkoutProgram() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (name: string) => createWorkoutProgram(name),
        onSuccess: () => invalidateWorkoutPrograms(queryClient),
    });
}

export function useRenameWorkoutProgram() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, name }: { id: number; name: string }) => renameWorkoutProgram(id, name),
        onSuccess: () => invalidateWorkoutPrograms(queryClient),
    });
}

export function useActivateWorkoutProgram() {
    const queryClient = useQueryClient();
    return useMutation({
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
        mutationFn: ({ programId, name, dayOfWeek }: { programId: number; name: string; dayOfWeek: number | null }) =>
            createWorkoutPlan(programId, name, dayOfWeek),
        onSuccess: () => invalidateWorkoutPrograms(queryClient),
    });
}

export function useSwitchWorkoutPlan() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ date, planId }: { date: string; planId: number | null }) =>
            switchWorkoutPlan(dateToOffset(date), planId),
        onSuccess: (_data, { date }) => {
            queryClient.invalidateQueries({ queryKey: workoutKeys.days.day(date) });
        },
    });
}

export function useExercisesPaginated(search: MaybeRefOrGetter<string> = '', pageSize = 10) {
    return useInfiniteQuery(computed(() => workoutQueries.exercisesPaginated(toValue(search), pageSize)));
}

export function useLogExercise() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async ({
            exercise,
            type,
        }: {
            date: string;
            exercise: LoggedExercise;
            type: 'logged' | 'previous';
        }) => {
            const body = await apiPOST<{ exercise: LoggedExercise }>('/workout/exercises/log', {
                exercise,
                type,
            });
            return body.exercise;
        },
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}

export function useAddExerciseToWorkout() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async ({ date, exerciseId }: { date: string; exerciseId: number }) => {
            const body = await apiPOST<{ exercise: LoggedExercise }>(
                '/workout/exercises/add',
                { exercise_id: exerciseId },
                { params: { offset: dateToOffset(date) } },
            );
            return body.exercise;
        },
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}

export function useRemoveExerciseFromWorkout() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async ({ date, exerciseId }: { date: string; exerciseId: number }) => {
            return apiDELETE<{ success?: boolean }>('/workout/exercises/remove', {
                data: { exercise_id: exerciseId },
                params: { offset: dateToOffset(date) },
            });
        },
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}

export function useDeleteLoggedSet() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ setId }: { date: string; setId: number }) => apiDELETE(`/workout/exercises/sets/${setId}`),
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}

export function useUpsertMobilityPre() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async ({ date, checked }: { date: string; checked: string[] }) => {
            const body = await apiPOST<{ mobility: MobilityLogged }>(
                '/workout/logs/mobility/pre',
                { checked },
                { params: { offset: dateToOffset(date) } },
            );
            return body.mobility;
        },
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}

export function useUpsertMobilityPost() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async ({ date, checked }: { date: string; checked: string[] }) => {
            const body = await apiPOST<{ mobility: MobilityLogged }>(
                '/workout/logs/mobility/post',
                { checked },
                { params: { offset: dateToOffset(date) } },
            );
            return body.mobility;
        },
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}

export function useUpsertCardio() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (payload: { date: string; minutes: number; type?: string; notes?: string }) => {
            const body = await apiPOST<{ cardio: Cardio }>(
                '/workout/logs/cardio',
                {
                    minutes: payload.minutes,
                    ...(payload.type !== undefined && payload.type !== ''
                        ? { type: payload.type }
                        : {}),
                    ...(payload.notes !== undefined ? { notes: payload.notes } : {}),
                },
                { params: { offset: dateToOffset(payload.date) } },
            );
            return body.cardio;
        },
        onSuccess: (_data, { date }) => invalidateDay(queryClient, date),
    });
}
