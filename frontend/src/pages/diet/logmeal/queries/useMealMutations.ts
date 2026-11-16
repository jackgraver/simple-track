import { useMutation, useQueryClient } from '@tanstack/vue-query';
import {
    createCompositeFood,
    createMeal,
    createSavedMeal,
    deleteSavedMeal,
    logEditedMeal,
    quickLog,
    updateLoggedMeal,
    updateSavedMeal,
    type QuickLogPayload,
    type SavedMealItemPayload,
} from '~/api/diet/api';
import { dietKeys } from '~/api/diet/keys';
import { searchListKeys } from '~/shared/searchListKeys';
import { parseDietDayOffsetQuery } from '~/utils/dateUtil';
import type { Meal } from '~/types/diet';
import { useRouter } from 'vue-router';

type QueryClient = ReturnType<typeof useQueryClient>;

const SAVED_MEALS_SEARCH_PATH = '/diet/meals/saved-meal/all';

function invalidateSavedMeals(queryClient: QueryClient) {
    queryClient.invalidateQueries({ queryKey: dietKeys.savedMeals.all });
    queryClient.invalidateQueries({
        queryKey: searchListKeys.path(SAVED_MEALS_SEARCH_PATH),
    });
}

function invalidateDietQueries(queryClient: QueryClient) {
    queryClient.invalidateQueries({ queryKey: dietKeys.days.all });
    invalidateSavedMeals(queryClient);
}

function afterMealLogNavigate(router: ReturnType<typeof useRouter>) {
    const route = router.currentRoute.value;
    const name = route.name;
    if (name === 'diet-log') {
        const offset = parseDietDayOffsetQuery(route.query.offset);
        router.push({
            name: 'diet',
            query: offset === 0 ? {} : { offset: String(offset) },
        });
        return;
    }
    if (name === 'home') {
        return;
    }
    router.push({ name: 'gym' });
}

export function useQuickLog() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: QuickLogPayload) => quickLog(payload),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
        },
    });
}

export function useCreateSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: { name: string; items: SavedMealItemPayload[] }) =>
            createSavedMeal(payload),
        onSuccess: () => {
            invalidateSavedMeals(queryClient);
        },
    });
}

export function useDeleteSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (vars: { savedMealId: number; force?: boolean }) =>
            deleteSavedMeal(vars.savedMealId, { force: vars.force }),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
        },
    });
}

export function useCreateCompositeFood() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: { name: string; items: { food_id: number; amount: number }[] }) =>
            createCompositeFood(payload),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: searchListKeys.all });
        },
    });
}

export function useCreateMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: ({
            meal,
            log,
            saveToLibrary,
            offset,
        }: {
            meal: Meal;
            log: boolean;
            saveToLibrary?: boolean;
            offset?: number;
        }) => createMeal(meal, log, { saveToLibrary, offset }),
        onSuccess: (_, variables) => {
            invalidateDietQueries(queryClient);
            if (variables.log) {
                afterMealLogNavigate(router);
            }
        },
    });
}

export function useLogEditedMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: (vars: {
            meal: Meal;
            plannedSourceMealId?: number;
            dayId?: number;
        }) =>
            logEditedMeal(vars.meal, {
                plannedSourceMealId: vars.plannedSourceMealId,
                dayId: vars.dayId,
            }),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
            afterMealLogNavigate(router);
        },
    });
}

export function useUpdateLoggedMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: ({
            meal,
            oldMealId,
            dayId,
        }: {
            meal: Meal;
            oldMealId: number;
            dayId?: number;
        }) => updateLoggedMeal(meal, oldMealId, { dayId }),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
            afterMealLogNavigate(router);
        },
    });
}

export function useUpdateSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: {
            savedMealId: number;
            name: string;
            items: SavedMealItemPayload[];
        }) =>
            updateSavedMeal(payload.savedMealId, {
                name: payload.name,
                items: payload.items,
            }),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({
                queryKey: dietKeys.savedMeals.detail(variables.savedMealId),
            });
            invalidateSavedMeals(queryClient);
        },
    });
}

