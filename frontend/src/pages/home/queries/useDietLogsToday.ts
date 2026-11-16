import { useQuery } from "@tanstack/vue-query";
import { computed, toValue, type MaybeRefOrGetter } from "vue";
import { getDietLogsToday } from "~/api/diet/api";
import { homeKeys } from "./keys";

export function useDietLogsToday(offset: MaybeRefOrGetter<number>) {
    return useQuery(
        computed(() => {
            const o = toValue(offset);
            return {
                queryKey: homeKeys.diet.today(o),
                queryFn: () => getDietLogsToday(o),
                staleTime: 1000 * 60 * 2,
            };
        }),
    );
}