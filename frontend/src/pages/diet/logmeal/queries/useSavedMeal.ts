import { useQuery } from '@tanstack/vue-query';
import { computed, type Ref } from 'vue';
import { getSavedMealById } from '~/api/diet/api';
import { dietKeys } from '~/api/diet/keys';

export function useSavedMeal(id: Ref<number | null> | number | null) {
    const savedMealId = computed(() => {
        if (id && typeof id === 'object' && 'value' in id) {
            const idValue = id.value;
            return idValue !== null && idValue !== 0 ? idValue : null;
        }
        const idValue = id as number | null;
        return idValue !== null && idValue !== 0 ? idValue : null;
    });
    const enabled = computed(
        () => savedMealId.value !== null && savedMealId.value !== 0,
    );
    const queryKey = computed(() =>
        savedMealId.value
            ? dietKeys.savedMeals.detail(savedMealId.value)
            : dietKeys.savedMeals.detail(0),
    );
    return useQuery({
        queryKey,
        queryFn: () => getSavedMealById(savedMealId.value!),
        enabled,
        staleTime: 1000 * 60 * 5,
    });
}
