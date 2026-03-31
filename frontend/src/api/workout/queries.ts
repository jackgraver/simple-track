import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { apiDELETE, apiGET, apiPOST } from '~/api/client';
import type { WorkoutLogsPreviousResponse } from '~/api/workout/api';
import { liveworkoutKeys } from '~/api/workout/keys';
import { homeKeys } from '~/pages/home/queries/keys';
import type { Cardio, Exercise, LoggedExercise } from '~/types/workout';

export function useWorkoutLogToday(offset: MaybeRefOrGetter<number> = 0) {
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
        },
    });
}

export function useRemoveExerciseFromWorkout(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (exerciseId: number) =>
            apiDELETE('/workout/exercises/remove', {
                data: { exercise_id: exerciseId },
                params: { offset: toValue(offset) },
            }),
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset),
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset),
            });
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
        },
    });
}

export function useUpsertCardio(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (payload: { minutes: number; type?: string }) => {
            const body = await apiPOST<{ cardio: Cardio }>(
                '/workout/logs/cardio',
                {
                    minutes: payload.minutes,
                    ...(payload.type !== undefined && payload.type !== ''
                        ? { type: payload.type }
                        : {}),
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
        },
    });
}
