import {
    useMutation,
    useQueryClient,
    type QueryClient,
} from '@tanstack/vue-query';
import {
    logPlannedMeal,
    deleteLoggedMeal,
    editLoggedMeal,
    deletePlannedMeal,
    addPlannedMealFromSaved,
    reorderPlannedMeals,
} from '~/api/diet/api';
import { homeKeys } from './keys';
import type { Meal } from '~/types/diet';
import { toValue, type MaybeRefOrGetter } from 'vue';

function invalidateMonthPlannedSummaries(queryClient: QueryClient) {
    queryClient.invalidateQueries({
        queryKey: homeKeys.diet.monthPlannedSummaryPrefix,
    });
}

export function useLogPlannedMeal(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (mealId: number) => logPlannedMeal(mealId),
        onSuccess: (day) => {
            queryClient.setQueryData(
                homeKeys.diet.today(toValue(offset)),
                day,
            );
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}

export function useDeleteLoggedMeal(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ mealId, dayId }: { mealId: number; dayId: number }) =>
            deleteLoggedMeal(mealId, dayId),
        onSuccess: (day) => {
            queryClient.setQueryData(
                homeKeys.diet.today(toValue(offset)),
                day,
            );
        },
    });
}

export function useEditLoggedMeal(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ meal, oldMealId }: { meal: Meal; oldMealId: number }) =>
            editLoggedMeal(meal, oldMealId),
        onSuccess: (day) => {
            queryClient.setQueryData(
                homeKeys.diet.today(toValue(offset)),
                day,
            );
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
            invalidateMonthPlannedSummaries(queryClient);
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
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}

export function useReorderPlannedMeals(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (plannedMealIds: number[]) =>
            reorderPlannedMeals(plannedMealIds, toValue(offset)),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(toValue(offset)),
            });
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}