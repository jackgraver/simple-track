import { useMutation, useQueryClient } from '@tanstack/vue-query';
import {
    logPlannedMeal,
    deleteLoggedMeal,
    editLoggedMeal,
    deletePlannedMeal,
    addPlannedMealFromSaved,
} from '~/api/diet/api';
import { homeKeys } from './keys';
import type { Meal } from '~/types/diet';
import { toValue, type MaybeRefOrGetter } from 'vue';

export function useLogPlannedMeal(offset: number) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (mealId: number) => logPlannedMeal(mealId),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(offset)
            });
        },
    });
}

export function useDeleteLoggedMeal(offset: number) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ mealId, dayId }: { mealId: number; dayId: number }) =>
            deleteLoggedMeal(mealId, dayId),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(offset)
            });
        },
    });
}

export function useEditLoggedMeal(offset: number) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ meal, oldMealId }: { meal: Meal; oldMealId: number }) =>
            editLoggedMeal(meal, oldMealId),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(offset)
            });
        },
    });
}

export function useDeletePlannedMeal(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (plannedMealId: number) =>
            deletePlannedMeal(plannedMealId, toValue(offset)),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(toValue(offset)),
            });
        },
    });
}

export function useAddPlannedFromSaved(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (savedMealId: number) =>
            addPlannedMealFromSaved(savedMealId, toValue(offset)),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(toValue(offset)),
            });
        },
    });
}