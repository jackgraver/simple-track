import { useQuery } from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { getWorkoutLogsPrevious } from '~/api/workout/api';
import { liveworkoutKeys } from './keys';

export function useWorkoutLogToday(offset: MaybeRefOrGetter<number> = 0) {
    return useQuery(
        computed(() => ({
            queryKey: liveworkoutKeys.workouts.day(toValue(offset)),
            queryFn: async () => {
                const response = await getWorkoutLogsPrevious(toValue(offset));
                return response.day;
            },
            staleTime: 1000 * 60 * 2,
        }))
    );
}
