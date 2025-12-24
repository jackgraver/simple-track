import { useQuery } from '@tanstack/vue-query';
import { getWorkoutLogsPrevious } from '../api/workouts';
import { homeKeys } from './keys';

export function useWorkoutLogsPrevious(offset: number) {
    return useQuery({
        queryKey: homeKeys.workouts.previous(offset),
        queryFn: () => getWorkoutLogsPrevious(offset),
        staleTime: 1000 * 60 * 2, // 2 minutes
    });
}

