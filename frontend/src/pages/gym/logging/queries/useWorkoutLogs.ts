import { useQuery } from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { getWorkoutLogsPrevious } from '~/api/workout/api';
import { liveworkoutKeys } from './keys';

export function useWorkoutLogsPrevious(offset: MaybeRefOrGetter<number> = 0) {
    return useQuery(
        computed(() => ({
            queryKey: liveworkoutKeys.workouts.previous(toValue(offset)),
            queryFn: () => getWorkoutLogsPrevious(toValue(offset)),
            staleTime: 1000 * 60 * 2,
        }))
    );
}

