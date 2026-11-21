import {
    useMutation,
    useQueries,
    useQuery,
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
    copyPlannedDay,
    getDietLogsToday,
    getAllPlans,
} from './api';
import { dietKeys } from './keys';
import type { Meal } from '~/types/diet';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';

export function useDietLogsToday(offset: MaybeRefOrGetter<number>) {
    return useQuery(
        computed(() => {
            const o = toValue(offset);
            return {
                queryKey: dietKeys.days.day(o),
                queryFn: () => getDietLogsToday(o),
                staleTime: 1000 * 60 * 2,
            };
        }),
    );
}

/** The `n` days before today (offset 1..n), sharing cache entries with `useDietLogsToday`. */
export function useRecentDietDays(n: number) {
    return useQueries({
        queries: Array.from({ length: n }, (_, i) => ({
            queryKey: dietKeys.days.day(i + 1),
            queryFn: () => getDietLogsToday(i + 1),
            staleTime: 1000 * 60 * 2,
        })),
    });
}

export function usePlans() {
    return useQuery({ queryKey: dietKeys.plans.all, queryFn: getAllPlans });
}

/** Several days by offset (e.g. a week), sharing cache entries with `useDietLogsToday`. */
export function useDietDays(offsets: MaybeRefOrGetter<number[]>) {
    return useQueries({
        queries: computed(() =>
            toValue(offsets).map((o) => ({
                queryKey: dietKeys.days.day(o),
                queryFn: () => getDietLogsToday(o),
                staleTime: 1000 * 60 * 2,
            })),
        ),
    });
}

function invalidateMonthPlannedSummaries(queryClient: QueryClient) {
    queryClient.invalidateQueries({
        queryKey: dietKeys.days.monthPlannedSummaryAll,
    });
}

export function useLogPlannedMeal(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (mealId: number) => logPlannedMeal(mealId),
        onSuccess: (day) => {
            queryClient.setQueryData(
                dietKeys.days.day(toValue(offset)),
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
                dietKeys.days.day(toValue(offset)),
                day,
            );
        },
    });
}

export function useEditLoggedMeal(offset: MaybeRefOrGetter<number>) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            meal,
            oldMealId,
            dayId,
        }: {
            meal: Meal;
            oldMealId: number;
            dayId?: number;
        }) => editLoggedMeal(meal, oldMealId, { dayId }),
        onSuccess: (day) => {
            queryClient.setQueryData(
                dietKeys.days.day(toValue(offset)),
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
                queryKey: dietKeys.days.day(toValue(offset)),
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
                queryKey: dietKeys.days.day(toValue(offset)),
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
                queryKey: dietKeys.days.day(toValue(offset)),
            });
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}

// ---- Planning across days: the offset travels with each call (the week board edits 7 days) ----

function invalidateDay(queryClient: QueryClient, offset: number) {
    queryClient.invalidateQueries({ queryKey: dietKeys.days.day(offset) });
    invalidateMonthPlannedSummaries(queryClient);
}

export function useAddPlannedFromSavedAt() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ savedMealId, offset }: { savedMealId: number; offset: number }) =>
            addPlannedMealFromSaved(savedMealId, offset),
        onSuccess: (_d, { offset }) => invalidateDay(queryClient, offset),
    });
}

export function useDeletePlannedMealAt() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ plannedMealId, offset }: { plannedMealId: number; offset: number }) =>
            deletePlannedMeal(plannedMealId, offset),
        onSuccess: (_d, { offset }) => invalidateDay(queryClient, offset),
    });
}

export function useReorderPlannedMealsAt() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ ids, offset }: { ids: number[]; offset: number }) => reorderPlannedMeals(ids, offset),
        onSuccess: (_d, { offset }) => invalidateDay(queryClient, offset),
    });
}

export function useCopyPlannedDay() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ from, to }: { from: number; to: number[] }) => copyPlannedDay(from, to),
        onSuccess: (_d, { to }) => {
            for (const o of to) queryClient.invalidateQueries({ queryKey: dietKeys.days.day(o) });
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}
