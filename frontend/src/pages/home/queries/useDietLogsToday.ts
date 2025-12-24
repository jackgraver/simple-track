import { useQuery } from '@tanstack/vue-query';
import { getDietLogsToday } from '../api/diet';
import { homeKeys } from './keys';

export function useDietLogsToday(offset: number) {
    return useQuery({
        queryKey: homeKeys.diet.today(offset),
        queryFn: () => getDietLogsToday(offset),
        staleTime: 1000 * 60 * 2, // 2 minutes
    });
}

