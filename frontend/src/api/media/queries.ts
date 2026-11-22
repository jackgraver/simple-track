import { useQuery } from "@tanstack/vue-query";
import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { getMedia, getMediaStats, listMedia, type MediaListParams } from "./api";
import { mediaKeys } from "./keys";

// Read hooks only. Pages still call api.ts directly until slice M1 migrates them.
export { mediaKeys };

export function useMediaList(params: MaybeRefOrGetter<MediaListParams>) {
    return useQuery({
        queryKey: computed(() => mediaKeys.list(toValue(params))),
        queryFn: () => listMedia(toValue(params)),
        placeholderData: (previous) => previous,
    });
}

export function useMediaItem(id: MaybeRefOrGetter<number>) {
    return useQuery({
        queryKey: computed(() => mediaKeys.detail(toValue(id))),
        queryFn: () => getMedia(toValue(id)),
        enabled: computed(() => Number.isFinite(toValue(id))),
    });
}

export function useMediaStats() {
    return useQuery({
        queryKey: mediaKeys.stats,
        queryFn: getMediaStats,
    });
}
