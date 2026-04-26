import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { createFood } from '~/api/diet/food';
import type { Food } from '~/types/diet';
import { toast } from '~/composables/toast/useToast';

export function useCreateFood() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { food: Food; relatedFoodId?: number }) =>
            createFood(vars.food, vars.relatedFoodId),
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ["searchList"] });
            toast.push("Food Created Successfully!", "success");
        },
        onError: (error: any) => {
            toast.push("Create Food Failed! " + (error.message || ""), "error");
        },
    });
}

