import { useQueries } from "@tanstack/vue-query";
import { computed, toValue, type MaybeRefOrGetter } from "vue";
import { getDietLogsToday } from "~/api/diet/api";
import { homeKeys } from "~/pages/home/queries/keys";
import { dietWeekDays } from "~/utils/dateUtil";

export function useDietWeekLogs(weekOffset: MaybeRefOrGetter<number>) {
    const days = computed(() => dietWeekDays(toValue(weekOffset)));
    const queries = useQueries({
        queries: computed(() =>
            days.value.map((day) => ({
                queryKey: homeKeys.diet.today(day.offset),
                queryFn: () => getDietLogsToday(day.offset),
                staleTime: 1000 * 60 * 2,
            })),
        ),
    });
    return { days, queries };
}
