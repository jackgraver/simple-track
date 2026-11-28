import { queryOptions, useQuery, useQueryClient, type QueryClient } from "@tanstack/vue-query";
import { computed, type MaybeRefOrGetter, toValue } from "vue";

import { getApiErrorMessage } from "~/api/httpError";
import { useOptimisticMutation, useServerMutation, type CachePatch } from "~/api/optimistic";
import { STALE } from "~/api/queryClient";
import {
    createMedia,
    deleteMedia,
    getMedia,
    getMediaStats,
    listMedia,
    saveCatalogItem,
    searchCatalog,
    updateCatalogItem,
    updateMedia,
    type MediaCatalogInput,
    type MediaInput,
    type MediaItem,
    type MediaListParams,
    type MediaPersonalInput,
    type MediaType,
} from "./api";
import { patchItem, patchList, replaceInList, revertItem, revertList, type MediaQuickFields } from "./cache";
import { mediaKeys } from "./keys";

export { mediaKeys };

export const mediaMutationKeys = {
    quickUpdate: ["media", "mutation", "quick-update"] as const,
    create: ["media", "mutation", "create"] as const,
    update: ["media", "mutation", "update"] as const,
    remove: ["media", "mutation", "remove"] as const,
};

/** Single source for hooks and route prefetch. */
export const mediaQueries = {
    list: (params: MediaListParams = {}) =>
        queryOptions({
            queryKey: mediaKeys.list(params),
            queryFn: () => listMedia(params),
            staleTime: STALE.default,
        }),
    stats: () =>
        queryOptions({
            queryKey: mediaKeys.stats,
            queryFn: getMediaStats,
            staleTime: STALE.default,
        }),
};

export function useMediaList(params: MaybeRefOrGetter<MediaListParams>) {
    return useQuery({
        queryKey: computed(() => mediaKeys.list(toValue(params))),
        queryFn: () => listMedia(toValue(params)),
        staleTime: STALE.default,
        placeholderData: (previous) => previous,
    });
}

/** The newest cached copy of an item from any list query, with when that list was fetched. */
function findInLists(qc: QueryClient, id: number): { item: MediaItem; updatedAt: number } | undefined {
    let best: { item: MediaItem; updatedAt: number } | undefined;
    for (const q of qc.getQueryCache().findAll({ queryKey: mediaKeys.lists })) {
        const item = (q.state.data as MediaItem[] | undefined)?.find((it) => it.ID === id);
        if (item && (!best || q.state.dataUpdatedAt > best.updatedAt)) best = { item, updatedAt: q.state.dataUpdatedAt };
    }
    return best;
}

/** Detail seeded from the list cache (shown instantly), then refetched in the background. */
export function useMediaItem(id: MaybeRefOrGetter<number>) {
    const qc = useQueryClient();
    return useQuery({
        queryKey: computed(() => mediaKeys.detail(toValue(id))),
        queryFn: () => getMedia(toValue(id)),
        enabled: computed(() => Number.isFinite(toValue(id))),
        staleTime: STALE.default,
        initialData: () => findInLists(qc, toValue(id))?.item,
        initialDataUpdatedAt: () => findInLists(qc, toValue(id))?.updatedAt,
    });
}

export function useMediaStats() {
    return useQuery({
        queryKey: mediaKeys.stats,
        queryFn: getMediaStats,
        staleTime: STALE.default,
    });
}

export type QuickUpdateVars = { id: number; fields: MediaQuickFields };

/**
 * Optimistic status / favourite / rating change (PATCH, idempotent).
 * Touches: detail(id) and every cached list that contains the item. Not patched: dates and stats
 * (server-derived), which refresh on settle. Rollback: snapshot, or field-wise revert when a sibling
 * also edited the entry. Per-item scope keeps two quick edits of one item in order.
 */
export function useQuickUpdateMedia() {
    const qc = useQueryClient();
    return useOptimisticMutation<QuickUpdateVars, MediaItem>({
        mutationKey: mediaMutationKeys.quickUpdate,
        mutationFn: ({ id, fields }) => updateMedia(id, fields),
        scope: ({ id }) => ({ id: `media:${id}` }),
        optimistic: ({ id }) => {
            const patches: CachePatch<QuickUpdateVars>[] = [
                {
                    queryKey: mediaKeys.detail(id),
                    apply: (old: MediaItem, v) => patchItem(old, v.fields),
                    revert: (cur: MediaItem, v, snap: MediaItem) => revertItem(cur, v.fields, snap),
                },
            ];
            for (const q of qc.getQueryCache().findAll({ queryKey: mediaKeys.lists })) {
                patches.push({
                    queryKey: q.queryKey,
                    apply: (old: MediaItem[], v) => patchList(old, v.id, v.fields),
                    revert: (cur: MediaItem[], v, snap: MediaItem[]) => revertList(cur, v.id, v.fields, snap),
                });
            }
            return patches;
        },
        reconcile: (item, { id }, client, { siblingsPending }) => {
            // Only replace wholesale when no other quick edit is in flight (it would be wiped).
            if (siblingsPending(mediaKeys.detail(id)) === 0) client.setQueryData(mediaKeys.detail(id), item);
            for (const q of client.getQueryCache().findAll({ queryKey: mediaKeys.lists })) {
                if (siblingsPending(q.queryKey) === 0) client.setQueryData<MediaItem[]>(q.queryKey, (l) => l && replaceInList(l, item));
            }
        },
        invalidate: () => [mediaKeys.stats, mediaKeys.lists],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to update"),
    });
}

export function useCreateMedia() {
    return useServerMutation<MediaInput, MediaItem>({
        mutationKey: mediaMutationKeys.create,
        mutationFn: createMedia,
        reconcile: (item, _v, qc) => qc.setQueryData(mediaKeys.detail(item.ID), item),
        invalidate: () => [mediaKeys.lists, mediaKeys.stats],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to save"),
    });
}

export function useUpdateMedia() {
    return useServerMutation<{ id: number; payload: MediaPersonalInput }, MediaItem>({
        mutationKey: mediaMutationKeys.update,
        mutationFn: ({ id, payload }) => updateMedia(id, payload),
        scope: ({ id }) => ({ id: `media:${id}` }),
        reconcile: (item, _v, qc) => qc.setQueryData(mediaKeys.detail(item.ID), item),
        invalidate: () => [mediaKeys.lists, mediaKeys.stats],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to save"),
    });
}

export function useDeleteMedia() {
    return useServerMutation<number, void>({
        mutationKey: mediaMutationKeys.remove,
        mutationFn: deleteMedia,
        reconcile: (_d, id, qc) => qc.removeQueries({ queryKey: mediaKeys.detail(id), exact: true }),
        invalidate: () => [mediaKeys.lists, mediaKeys.stats],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to delete"),
    });
}

/** Search the shared catalog (what everyone has added). Off until there is something to search for. */
export function useCatalogSearch(q: MaybeRefOrGetter<string>, type: MaybeRefOrGetter<MediaType | undefined>) {
    return useQuery({
        queryKey: computed(() => mediaKeys.catalog(toValue(q).trim(), toValue(type))),
        queryFn: () => searchCatalog(toValue(q).trim(), toValue(type)),
        enabled: computed(() => toValue(q).trim().length >= 2),
        staleTime: STALE.default,
        placeholderData: (previous) => previous,
    });
}

/** Add a shared catalog entry to your own list. */
export function useSaveCatalogItem() {
    return useServerMutation<number, MediaItem>({
        mutationKey: ["media", "mutation", "save-catalog"] as const,
        mutationFn: saveCatalogItem,
        reconcile: (item, _v, qc) => qc.setQueryData(mediaKeys.detail(item.ID), item),
        invalidate: () => [mediaKeys.lists, mediaKeys.stats, mediaKeys.catalogs],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to add"),
    });
}

/** Edit a catalog entry (its creator or an admin). Your list and the picker refresh. */
export function useUpdateCatalogItem() {
    return useServerMutation<{ id: number; payload: MediaCatalogInput }, unknown>({
        mutationKey: ["media", "mutation", "update-catalog"] as const,
        mutationFn: ({ id, payload }) => updateCatalogItem(id, payload),
        invalidate: () => [mediaKeys.lists, mediaKeys.details, mediaKeys.catalogs],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to save"),
    });
}
