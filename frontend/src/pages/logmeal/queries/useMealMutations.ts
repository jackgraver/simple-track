import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { createMeal, logEditedMeal, updateLoggedMeal } from '../api/meals';
import { logmealKeys } from './keys';
import { homeKeys } from '~/pages/home/queries/keys';
import type { Meal } from '~/types/diet';
import { useRouter } from 'vue-router';

export function useCreateMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: ({ meal, log }: { meal: Meal; log: boolean }) =>
            createMeal(meal, log),
        onSuccess: (_, variables) => {
            // Invalidate today's diet logs
            queryClient.invalidateQueries({ queryKey: logmealKeys.diet.today() });
            queryClient.invalidateQueries({ queryKey: homeKeys.diet.today(0) });
            
            if (variables.log) {
                router.push('/');
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
            // Invalidate today's diet logs
            queryClient.invalidateQueries({ queryKey: logmealKeys.diet.today() });
            queryClient.invalidateQueries({ queryKey: homeKeys.diet.today(0) });
            router.push('/');
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
            // Invalidate today's diet logs
            queryClient.invalidateQueries({ queryKey: logmealKeys.diet.today() });
            queryClient.invalidateQueries({ queryKey: homeKeys.diet.today(0) });
            router.push('/');
        },
    });
}

