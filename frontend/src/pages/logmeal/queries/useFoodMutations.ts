import { useMutation } from '@tanstack/vue-query';
import { createFood } from '../api/food';
import type { Food } from '~/types/diet';
import { toast } from '~/composables/toast/useToast';

export function useCreateFood() {
    return useMutation({
        mutationFn: (food: Food) => createFood(food),
        onSuccess: () => {
            toast.push("Food Created Successfully!", "success");
        },
        onError: (error: any) => {
            toast.push("Create Food Failed! " + (error.message || ""), "error");
        },
    });
}

