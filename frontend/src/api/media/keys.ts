import type { MediaListParams, MediaType } from "./api";

export const mediaKeys = {
    all: ["media"] as const,
    /** Prefix of every list query (any filter). */
    lists: ["media", "list"] as const,
    list: (params: MediaListParams) => ["media", "list", params] as const,
    details: ["media", "detail"] as const,
    detail: (id: number) => ["media", "detail", id] as const,
    stats: ["media", "stats"] as const,
    /** Prefix of every shared-catalog search. */
    catalogs: ["media", "catalog"] as const,
    catalog: (q: string, type?: MediaType) => ["media", "catalog", q, type ?? ""] as const,
};
