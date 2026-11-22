import type { MediaListParams } from "./api";

export const mediaKeys = {
    all: ["media"] as const,
    /** Prefix of every list query (any filter). */
    lists: ["media", "list"] as const,
    list: (params: MediaListParams) => ["media", "list", params] as const,
    details: ["media", "detail"] as const,
    detail: (id: number) => ["media", "detail", id] as const,
    stats: ["media", "stats"] as const,
};
