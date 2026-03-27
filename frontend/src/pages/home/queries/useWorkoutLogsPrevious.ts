import { useQuery } from '@tanstack/vue-query';
import { getWorkoutLogsPrevious } from '~/api/workout/api';
import { homeKeys } from './keys';

export function useWorkoutLogsPrevious(offset: number) {
    return useQuery({
        queryKey: homeKeys.workouts.previous(offset),
        queryFn: () => getWorkoutLogsPrevious(offset),
        staleTime: 1000 * 60 * 2, // 2 minutes
    });
}

