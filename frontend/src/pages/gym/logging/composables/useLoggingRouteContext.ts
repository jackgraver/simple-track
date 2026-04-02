import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";

export function buildLoggingListQuery(offset: number): Record<string, string> | Record<string, never> {
    return offset === 0 ? {} : { offset: String(offset) };
}

export function useLoggingRouteContext() {
    const route = useRoute();
    const router = useRouter();

    const offset = computed(() => {
        const raw = route.query.offset;
        const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
        return Number.isNaN(value) ? 0 : value;
    });

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
