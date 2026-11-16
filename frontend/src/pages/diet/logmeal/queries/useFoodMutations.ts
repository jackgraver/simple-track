import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { createFood } from '~/api/diet/food';
import type { Food } from '~/types/diet';
import { toast } from '~/composables/toast/useToast';
import { getApiErrorMessage } from '~/api/httpError';
import { searchListKeys } from '~/shared/searchListKeys';

export function useCreateFood() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { food: Food; relatedFoodId?: number }) =>
            createFood(vars.food, vars.relatedFoodId),
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: searchListKeys.all });
            toast.push("Food Created Successfully!", "success");
        },
        onError: (error) => {
            toast.push(getApiErrorMessage(error, "Create food failed"), "error");
        },
    });
}

