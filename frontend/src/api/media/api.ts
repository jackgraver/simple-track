import { apiClient } from "~/api/client";

export const MEDIA_TYPES = [
    "movie",
    "tv",
    "book",
    "game",
    "anime",
    "comic",
    "album",
    "podcast",
    "other",
] as const;
export type MediaType = (typeof MEDIA_TYPES)[number];

export const MEDIA_STATUSES = [
    "planned",
    "in_progress",
    "completed",
    "on_hold",
    "dropped",
] as const;
export type MediaStatus = (typeof MEDIA_STATUSES)[number];

export const MEDIA_TYPE_LABELS: Record<MediaType, string> = {
    movie: "Movie",
    tv: "TV",
    book: "Book",
    game: "Game",
    anime: "Anime",
    comic: "Comic",
    album: "Album",
    podcast: "Podcast",
    other: "Other",
};

export const MEDIA_STATUS_LABELS: Record<MediaStatus, string> = {
    planned: "Want to",
    in_progress: "In progress",
    completed: "Completed",
    on_hold: "On hold",
    dropped: "Dropped",
};

export type MediaItem = {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    type: MediaType;
    title: string;
    creator: string;
    release_year: number | null;
    genres: string;
    cover_url: string;
    synopsis: string;
    status: MediaStatus;
    rating: number | null;
    review: string;
    favorite: boolean;
    progress: number | null;
    length: number | null;
    started_at: string | null;
    finished_at: string | null;
    times_consumed: number;
    source: string;
    external_id: string;
    /** The shared catalog entry behind this save (several people can save the same one). */
    media_item_id: number;
    /** Who added the catalog entry. */
    created_by: number;
    /** True for the entry's creator and for admins: only they can change the catalog fields. */
    can_edit_catalog: boolean;
};

/** Catalog fields (shared by everyone): only the entry's creator or an admin may change them. */
export type MediaCatalogInput = Partial<
    Pick<
        MediaItem,
        | "type"
        | "title"
        | "creator"
        | "release_year"
        | "genres"
        | "cover_url"
        | "synopsis"
        | "length"
        | "source"
        | "external_id"
    >
>;

/** Your own fields of a save. Omit to leave unchanged; null clears. Dates are YYYY-MM-DD. */
export type MediaPersonalInput = Partial<
    Pick<MediaItem, "status" | "rating" | "review" | "favorite" | "progress" | "times_consumed">
> & {
    started_at?: string | null;
    finished_at?: string | null;
};

/** Create: a catalog entry (reused when source + external_id already exist) plus your save of it. */
export type MediaInput = MediaCatalogInput & MediaPersonalInput;

/** A shared catalog entry as the add picker lists it. */
export type CatalogEntry = Pick<
    MediaItem,
    | "type"
    | "title"
    | "creator"
    | "release_year"
    | "genres"
    | "cover_url"
    | "synopsis"
    | "length"
    | "source"
    | "external_id"
    | "created_by"
    | "can_edit_catalog"
    | "CreatedAt"
    | "UpdatedAt"
> & {
    ID: number;
    /** Your save of it (media id), or null if you haven't saved it. */
    saved_id: number | null;
};

export type MediaListParams = {
    type?: MediaType;
    status?: MediaStatus;
    q?: string;
    favorite?: boolean;
    sort?: "updated" | "created" | "title" | "rating" | "finished" | "release";
};

export type MediaStats = {
    total: number;
    by_type: Partial<Record<MediaType, number>>;
    by_status: Partial<Record<MediaStatus, number>>;
    average_rating: number | null;
    completed_this_year: number;
};

export async function listMedia(params: MediaListParams = {}) {
    const response = await apiClient.get<{ items: MediaItem[] }>("/media", {
        params,
    });
    return response.data.items;
}

export async function getMedia(id: number) {
    const response = await apiClient.get<{ item: MediaItem }>(`/media/${id}`);
    return response.data.item;
}

export async function createMedia(payload: MediaInput) {
    const response = await apiClient.post<{ item: MediaItem }>("/media", payload);
    return response.data.item;
}

export async function updateMedia(id: number, payload: MediaPersonalInput) {
    const response = await apiClient.patch<{ item: MediaItem }>(
        `/media/${id}`,
        payload,
    );
    return response.data.item;
}

export async function deleteMedia(id: number) {
    await apiClient.delete(`/media/${id}`);
}

export async function searchCatalog(q: string, type?: MediaType) {
    const response = await apiClient.get<{ items: CatalogEntry[] }>("/media/catalog", {
        params: { q, type },
    });
    return response.data.items;
}

/** Save an existing catalog entry to your list (status "planned"). Returns your save. */
export async function saveCatalogItem(catalogId: number) {
    const response = await apiClient.post<{ item: MediaItem }>(`/media/catalog/${catalogId}/save`);
    return response.data.item;
}

export async function updateCatalogItem(catalogId: number, payload: MediaCatalogInput) {
    const response = await apiClient.patch<{ item: CatalogEntry }>(`/media/catalog/${catalogId}`, payload);
    return response.data.item;
}

export async function getMediaStats() {
    const response = await apiClient.get<{ stats: MediaStats }>("/media/stats");
    return response.data.stats;
}

/** Media types that have an online lookup source. */
export const LOOKUP_TYPES: ReadonlySet<MediaType> = new Set<MediaType>([
    "movie",
    "tv",
    "book",
    "anime",
    "comic",
    "album",
    "podcast",
]);

export const LOOKUP_SOURCE_LABELS: Record<string, string> = {
    tmdb: "TMDB",
    openlibrary: "Open Library",
    anilist: "AniList",
    itunes: "Apple",
};

export type LookupResult = {
    source: string;
    external_id: string;
    title: string;
    creator: string;
    release_year: number | null;
    synopsis: string;
    genres: string;
    cover_url: string;
    length: number | null;
};

export async function lookupMedia(type: MediaType, q: string) {
    const response = await apiClient.get<{ results: LookupResult[] }>(
        "/media/lookup",
        { params: { type, q } },
    );
    return response.data.results;
}

export async function lookupMediaDetails(source: string, id: string) {
    const response = await apiClient.get<{ result: LookupResult }>(
        "/media/lookup/details",
        { params: { source, id } },
    );
    return response.data.result;
}
