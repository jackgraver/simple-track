import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { toValue, type MaybeRefOrGetter } from 'vue';
import { logExercise, addExerciseToWorkout, removeExerciseFromWorkout, deleteLoggedSet } from '~/api/workout/api';
import { liveworkoutKeys } from './keys';
import type { LoggedExercise } from '~/types/workout';

export function useLogExercise(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ exercise, type }: { exercise: LoggedExercise; type: "logged" | "previous" }) =>
            logExercise(exercise, type),
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset)
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset)
            });
        },
    });
}

export function useAddExerciseToWorkout(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (exerciseId: number) =>
            addExerciseToWorkout(exerciseId, toValue(offset)),
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset)
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset)
            });
        },
    });
}

export function useRemoveExerciseFromWorkout(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (exerciseId: number) =>
            removeExerciseFromWorkout(exerciseId, toValue(offset)),
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset)
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset)
            });
        },
    });
}

export function useDeleteLoggedSet(offset: MaybeRefOrGetter<number> = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (setId: number) => deleteLoggedSet(setId),
        onSuccess: () => {
            const currentOffset = toValue(offset);
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.previous(currentOffset)
            });
            queryClient.invalidateQueries({
                queryKey: liveworkoutKeys.workouts.day(currentOffset)
            });
        },
    });
}

