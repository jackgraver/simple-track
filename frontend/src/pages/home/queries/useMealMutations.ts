import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { logPlannedMeal, deleteLoggedMeal, editLoggedMeal } from '../api/diet';
import { homeKeys } from './keys';
import type { Meal } from '~/types/diet';

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

