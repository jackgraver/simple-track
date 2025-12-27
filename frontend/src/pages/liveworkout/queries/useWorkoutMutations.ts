import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { logExercise, addExerciseToWorkout, removeExerciseFromWorkout } from '../api/workouts';
import { liveworkoutKeys } from './keys';
import type { LoggedExercise } from '~/types/workout';

export function useLogExercise(offset: number = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ exercise, type }: { exercise: LoggedExercise; type: "logged" | "previous" }) =>
            logExercise(exercise, type),
        onSuccess: () => {
            queryClient.invalidateQueries({ 
                queryKey: liveworkoutKeys.workouts.previous(offset) 
            });
        },
    });
}

export function useAddExerciseToWorkout(offset: number = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (exerciseId: number) => addExerciseToWorkout(exerciseId),
        onSuccess: () => {
            queryClient.invalidateQueries({ 
                queryKey: liveworkoutKeys.workouts.previous(offset) 
            });
        },
    });
}

export function useRemoveExerciseFromWorkout(offset: number = 0) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (exerciseId: number) => removeExerciseFromWorkout(exerciseId),
        onSuccess: () => {
            queryClient.invalidateQueries({ 
                queryKey: liveworkoutKeys.workouts.previous(offset) 
            });
        },
    });
}

