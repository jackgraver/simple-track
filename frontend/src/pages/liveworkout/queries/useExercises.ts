import { useQuery } from '@tanstack/vue-query';
import { getAllExercises } from '../api/workouts';
import { liveworkoutKeys } from './keys';

export function useAllExercises() {
    return useQuery({
        queryKey: liveworkoutKeys.exercises.allList(),
        queryFn: () => getAllExercises(),
        staleTime: 1000 * 60 * 10, // 10 minutes - exercises don't change often
    });
}

