import { useQuery } from '@tanstack/vue-query';
import { getDietLogsToday } from '../api/meals';
import { logmealKeys } from './keys';

export function useDietLogsToday() {
    return useQuery({
        queryKey: logmealKeys.diet.today(),
        queryFn: () => getDietLogsToday(),
        staleTime: 1000 * 60 * 2, // 2 minutes
    });
}

