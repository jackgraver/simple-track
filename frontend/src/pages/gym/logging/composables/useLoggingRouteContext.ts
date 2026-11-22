import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { parseDietDayOffsetQuery } from "~/utils/dateUtil";

export function buildLoggingListQuery(offset: number): Record<string, string> | Record<string, never> {
    return offset === 0 ? {} : { offset: String(offset) };
}

export function useLoggingRouteContext() {
    const route = useRoute();
    const router = useRouter();

    const offset = computed(() => parseDietDayOffsetQuery(route.query.offset));

    const loggingListQuery = computed(() => buildLoggingListQuery(offset.value));

    const goBackToLogging = () => {
        router.push({ name: "logging", query: loggingListQuery.value });
    };

    return {
        route,
        router,
        offset,
        loggingListQuery,
        goBackToLogging,
    };
}
