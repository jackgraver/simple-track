import { useQuery } from '@tanstack/vue-query';
import { getWorkoutLogsPrevious } from '../api/workouts';
import { liveworkoutKeys } from './keys';

export function useWorkoutLogsPrevious(offset: number = 0) {
    return useQuery({
        queryKey: liveworkoutKeys.workouts.previous(offset),
        queryFn: () => getWorkoutLogsPrevious(offset),
        staleTime: 1000 * 60 * 2, // 2 minutes
    });
}

