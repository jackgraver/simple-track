import { useQuery } from '@tanstack/vue-query';
import { getWorkoutLogsToday } from '~/api/workout/api';
import { liveworkoutKeys } from './keys';

export function useWorkoutLogToday() {
    return useQuery({
        queryKey: liveworkoutKeys.workouts.today(),
        queryFn: () => getWorkoutLogsToday(),
        staleTime: 1000 * 60 * 2,
    });
}
