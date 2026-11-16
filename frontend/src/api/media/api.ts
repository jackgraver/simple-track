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
};

/** Fields accepted on create/update. Omit to leave unchanged; null clears. Dates are YYYY-MM-DD. */
export type MediaInput = Partial<
    Omit<MediaItem, "ID" | "CreatedAt" | "UpdatedAt" | "started_at" | "finished_at">
> & {
    started_at?: string | null;
    finished_at?: string | null;
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

export async function updateMedia(id: number, payload: MediaInput) {
    const response = await apiClient.patch<{ item: MediaItem }>(
        `/media/${id}`,
        payload,
    );
    return response.data.item;
}

export async function deleteMedia(id: number) {
    await apiClient.delete(`/media/${id}`);
}

export async function getMediaStats() {
    const response = await apiClient.get<{ stats: MediaStats }>("/media/stats");
    return response.data.stats;
}
