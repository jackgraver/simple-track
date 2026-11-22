import { watch, type MaybeRefOrGetter, toValue } from "vue";
import { useQueryClient, type QueryClient } from "@tanstack/vue-query";
import { dietQueries } from "~/api/diet/queries";
import { workoutQueries } from "~/api/workout/queries";
import { offsetToDate } from "~/api/workout/dates";

type DayOptions = (offset: number) => object;
type Prefetcher = Pick<QueryClient, "prefetchQuery">;

/**
 * Day offsets next to `offset`: older (+1) and newer (-1), as the day views swipe. Future days
 * (offset < 0) are left out: the day GETs create the row with the weekday's current plan, so an
 * early fetch would pin tomorrow's plan before the user changes their program or weekday plans.
 */
export function adjacentOffsets(offset: number): number[] {
    return [offset + 1, offset - 1].filter((o) => o >= 0);
}

/** Diet and gym day queries by offset (gym keys are dates, diet derives its date itself). */
export const dietDayOptions: DayOptions = (o) => dietQueries.day(o);
export const gymDayOptions: DayOptions = (o) => workoutQueries.day(offsetToDate(o));

/** Fire-and-forget prefetch of `offset` ± 1. Honours `staleTime`; failures are swallowed. */
export function prefetchAdjacentDays(qc: Prefetcher, offset: number, day: DayOptions): void {
    for (const o of adjacentOffsets(offset)) {
        void qc.prefetchQuery(day(o) as Parameters<Prefetcher["prefetchQuery"]>[0]);
    }
}

/**
 * Day views: once the viewed day has data, warm the days either side so a swipe or week-strip tap
 * renders from cache. `ensureQueryData` returns cached data or joins the page's in-flight request
 * (same key), so this adds no request for the viewed day itself.
 */
export function useAdjacentDayPrefetch(
    offset: MaybeRefOrGetter<number>,
    day: DayOptions,
    enabled: MaybeRefOrGetter<boolean> = true,
): void {
    const qc = useQueryClient();
    watch(
        () => (toValue(enabled) ? toValue(offset) : null),
        (o) => {
            if (o === null) return;
            const ready = qc.ensureQueryData(day(o) as Parameters<QueryClient["ensureQueryData"]>[0]);
            ready.then(
                () => {
                    if (toValue(offset) === o) prefetchAdjacentDays(qc, o, day);
                },
                () => undefined,
            );
        },
        { immediate: true },
    );
}
