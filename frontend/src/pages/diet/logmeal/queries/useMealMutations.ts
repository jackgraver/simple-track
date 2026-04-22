import { useMutation, useQueryClient } from '@tanstack/vue-query';
import {
    createCompositeFood,
    createMeal,
    createSavedMeal,
    logEditedMeal,
    updateLoggedMeal,
    type SavedMealItemPayload,
} from '~/api/diet/api';
import { logmealKeys } from './keys';
import { homeKeys } from '~/pages/home/queries/keys';
import type { Meal } from '~/types/diet';
import { useRouter } from 'vue-router';

function invalidateDietQueries(
    queryClient: ReturnType<typeof useQueryClient>,
) {
    queryClient.invalidateQueries({ queryKey: logmealKeys.diet.today() });
    queryClient.invalidateQueries({ queryKey: homeKeys.diet.all });
    queryClient.invalidateQueries({ queryKey: ['savedMeals'] });
    queryClient.invalidateQueries({
        queryKey: ['searchList', '/diet/meals/saved-meal/all'],
    });
}

function afterMealLogNavigate(router: ReturnType<typeof useRouter>) {
    const name = router.currentRoute.value.name;
    if (name === 'diet-log') {
        router.push({ name: 'diet' });
        return;
    }
    if (name === 'home') {
        return;
    }
    router.push({ name: 'gym' });
}

export function useCreateSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: { name: string; items: SavedMealItemPayload[] }) =>
            createSavedMeal(payload),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['savedMeals'] });
            queryClient.invalidateQueries({
                queryKey: ['searchList', '/diet/meals/saved-meal/all'],
            });
        },
    });
}

export function useCreateCompositeFood() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: { name: string; items: { food_id: number; amount: number }[] }) =>
            createCompositeFood(payload),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['searchList'] });
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
        }: {
            meal: Meal;
            log: boolean;
            saveToLibrary?: boolean;
        }) => createMeal(meal, log, saveToLibrary),
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
        mutationFn: (meal: Meal) => logEditedMeal(meal),
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
        mutationFn: ({ meal, oldMealId }: { meal: Meal; oldMealId: number }) =>
            updateLoggedMeal(meal, oldMealId),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
            afterMealLogNavigate(router);
        },
    });
}

