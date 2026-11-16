import { useQuery } from '@tanstack/vue-query';
import { computed, type Ref } from 'vue';
import { getMealById } from '~/api/diet/api';
import { dietKeys } from '~/api/diet/keys';

export function useMeal(id: Ref<number | null> | number | null) {
    const mealId = computed(() => {
        if (id && typeof id === 'object' && 'value' in id) {
            const idValue = id.value;
            return idValue !== null && idValue !== 0 ? idValue : null;
        }
        const idValue = id as number | null;
        return idValue !== null && idValue !== 0 ? idValue : null;
    });

    const enabled = computed(() => mealId.value !== null && mealId.value !== 0);
    const queryKey = computed(() =>
        mealId.value ? dietKeys.meals.detail(mealId.value) : dietKeys.meals.detail(0)
    );

    return useQuery({
        queryKey: queryKey,
        queryFn: () => getMealById(mealId.value!),
        enabled: enabled,
        staleTime: 1000 * 60 * 5, // 5 minutes
    });
}

