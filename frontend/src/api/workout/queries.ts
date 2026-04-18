import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { apiDELETE, apiGET, apiPOST } from '~/api/client';
import { switchWorkoutPlan, type WorkoutActivityResponse, WorkoutLogsPreviousResponse } from '~/api/workout/api';
import { liveworkoutKeys } from '~/api/workout/keys';
import { homeKeys } from '~/pages/home/queries/keys';
import type { Cardio, Exercise, LoggedExercise, MobilityLogged, WorkoutPlan } from '~/types/workout';

export type WorkoutActivityQueryOpts = {
    mode: 'year' | 'rolling';
    weeks: number | null;
    days: number | null;
};

export const workoutExercisesAllQueryKey = ['workout', 'exercises', 'all'] as const;

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

export function useWorkoutExercisesAllQuery(enabled: MaybeRefOrGetter<boolean> = true) {
    return useQuery({
        queryKey: workoutExercisesAllQueryKey,
        queryFn: fetchWorkoutExercisesAll,
        enabled: computed(() => toValue(enabled)),
    });
}

function invalidateWorkoutActivityQueries(queryClient: ReturnType<typeof useQueryClient>) {
    queryClient.invalidateQueries({ queryKey: liveworkoutKeys.workouts.activityPrefix() });
}

export function useWorkoutActivity(opts: MaybeRefOrGetter<WorkoutActivityQueryOpts>) {
    return useQuery(
        computed(() => {
            const o = toValue(opts);
            return {
                queryKey: liveworkoutKeys.workouts.activity(o),
                queryFn: async () => {
                    const params: Record<string, string | number> = { mode: o.mode };
                    if (o.mode === 'rolling') {
                        if (o.days != null && o.days > 0) {
                            params.days = o.days;
                        } else if (o.weeks != null && o.weeks > 0) {
                            params.weeks = o.weeks;
                        }
                    }
                    return apiGET<WorkoutActivityResponse>('/workout/logs/activity', { params });
                },
                staleTime: 1000 * 60 * 2,
            };
        }),
    );
}

export function useWorkoutLogToday(
    offset: MaybeRefOrGetter<number> = 0,
    enabled: MaybeRefOrGetter<boolean> = true,
) {
    return useQuery(
        computed(() => ({
            queryKey: liveworkoutKeys.workouts.day(toValue(offset)),
            queryFn: async () => {
                const response = await apiGET<WorkoutLogsPreviousResponse>(
                    '/workout/logs/previous',
                    { params: { offset: toValue(offset) } },
                );
                return response.day;
            },
            staleTime: 1000 * 60 * 2,
            enabled: toValue(enabled),
        })),
    );
}

export function useWorkoutLogsPrevious(offset: MaybeRefOrGetter<number> = 0) {
    return useQuery(
        computed(() => ({
            queryKey: liveworkoutKeys.workouts.previous(toValue(offset)),
            queryFn: () =>
                apiGET<WorkoutLogsPreviousResponse>('/workout/logs/previous', {
                    params: { offset: toValue(offset) },
                }),
            staleTime: 1000 * 60 * 2,
        })),
    );
}

export function useHomeWorkoutLogsPrevious(offset: number) {
    return useQuery({
        queryKey: homeKeys.workouts.previous(offset),
        queryFn: () =>
            apiGET<WorkoutLogsPreviousResponse>('/workout/logs/previous', {
                params: { offset },
            }),
        staleTime: 1000 * 60 * 2,
    });
}

export function useWorkoutPlansAll() {
    return useQuery({
        queryKey: ['workout', 'plans', 'all'] as const,
        queryFn: () => apiGET<{ plans: WorkoutPlan[] }>('/workout/plans/all'),
        staleTime: 1000 * 60 * 10,
    });
}

export function useSwitchWorkoutPlan(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (planId: number | null) => {
            return switchWorkoutPlan(toValue(offset), planId);
        },
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: homeKeys.workouts.previous(currentOffset),
            });
        },
    });
}

export function useAllExercises() {
    return useQuery({
        queryKey: liveworkoutKeys.exercises.allList(),
        queryFn: async () => {
            const body = await apiGET<{ exercises: Exercise[] }>('/workout/exercises/all');
            return body.exercises;
        },
        staleTime: 1000 * 60 * 10,
    });
}

type ExercisesPageResponse = {
    exercises: Exercise[];
    has_next: boolean;
    total: number;
    page: number;
    page_size: number;
};

export function useExercisesPaginated(search: MaybeRefOrGetter<string> = '', pageSize = 10) {
    return useInfiniteQuery(
        computed(() => ({
            queryKey: liveworkoutKeys.exercises.paginated(toValue(search)),
            queryFn: async ({ pageParam }: { pageParam: number }) => {
                const params: Record<string, string | number> = { page: pageParam, page_size: pageSize };
                const s = toValue(search);
                if (s) params.search = s;
                return apiGET<ExercisesPageResponse>('/workout/exercises/all', { params });
            },
            getNextPageParam: (lastPage: ExercisesPageResponse) =>
                lastPage.has_next ? lastPage.page + 1 : undefined,
            initialPageParam: 1,
            staleTime: 1000 * 60 * 10,
        })),
    );
}

export function useLogExercise(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async ({
            exercise,
            type,
        }: {
            exercise: LoggedExercise;
            type: 'logged' | 'previous';
        }) => {
            const body = await apiPOST<{ exercise: LoggedExercise }>('/workout/exercises/log', {
                exercise,
                type,
            });
            return body.exercise;
        },
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            invalidateWorkoutActivityQueries(queryClient);
        },
    });
}

export function useAddExerciseToWorkout(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (exerciseId: number) => {
            const body = await apiPOST<{ exercise: LoggedExercise }>(
                '/workout/exercises/add',
                { exercise_id: exerciseId },
                { params: { offset: toValue(offset) } },
            );
            return body.exercise;
        },
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            invalidateWorkoutActivityQueries(queryClient);
        },
    });
}

export function useDeleteLoggedSet(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (setId: number) => apiDELETE(`/workout/exercises/sets/${setId}`),
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            invalidateWorkoutActivityQueries(queryClient);
        },
    });
}

export function useUpsertMobilityPre(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (checked: string[]) => {
            const body = await apiPOST<{ mobility: MobilityLogged }>(
                '/workout/logs/mobility/pre',
                { checked },
                { params: { offset: toValue(offset) } },
            );
            return body.mobility;
        },
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            invalidateWorkoutActivityQueries(queryClient);
        },
    });
}

export function useUpsertMobilityPost(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (checked: string[]) => {
            const body = await apiPOST<{ mobility: MobilityLogged }>(
                '/workout/logs/mobility/post',
                { checked },
                { params: { offset: toValue(offset) } },
            );
            return body.mobility;
        },
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            invalidateWorkoutActivityQueries(queryClient);
        },
    });
}

export function useUpsertCardio(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (payload: { minutes: number; type?: string; notes?: string }) => {
            const body = await apiPOST<{ cardio: Cardio }>(
                '/workout/logs/cardio',
                {
                    minutes: payload.minutes,
                    ...(payload.type !== undefined && payload.type !== ''
                        ? { type: payload.type }
                        : {}),
                    ...(payload.notes !== undefined ? { notes: payload.notes } : {}),
                },
                { params: { offset: toValue(offset) } },
            );
            return body.cardio;
        },
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
            invalidateWorkoutActivityQueries(queryClient);
        },
    });
}
