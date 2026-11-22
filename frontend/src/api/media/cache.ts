import type { MediaItem, MediaStatus } from "./api";

/** Fields a quick (instant) edit may change. Dates are server-filled from the status, so never patched. */
export type MediaQuickFields = {
    status?: MediaStatus;
    favorite?: boolean;
    rating?: number | null;
};

/** The item with `fields` applied. Only keys present in `fields` change. */
export function patchItem(item: MediaItem, fields: MediaQuickFields): MediaItem {
    return { ...item, ...definedOnly(fields) };
}

/** A list with item `id` patched in place (other rows keep their identity). Missing id: unchanged. */
export function patchList(list: MediaItem[], id: number, fields: MediaQuickFields): MediaItem[] {
    return list.map((it) => (it.ID === id ? patchItem(it, fields) : it));
}

/**
 * Rollback when a sibling also patched this entry: put back only the fields we changed, to the values
 * the snapshot had, leaving the sibling's other fields alone.
 */
export function revertItem(current: MediaItem, fields: MediaQuickFields, snapshot: MediaItem): MediaItem {
    const restore: Record<string, unknown> = {};
    for (const key of Object.keys(definedOnly(fields)) as (keyof MediaQuickFields)[]) restore[key] = snapshot[key];
    return { ...current, ...restore };
}

export function revertList(list: MediaItem[], id: number, fields: MediaQuickFields, snapshot: MediaItem[]): MediaItem[] {
    const before = snapshot.find((it) => it.ID === id);
    if (!before) return list;
    return list.map((it) => (it.ID === id ? revertItem(it, fields, before) : it));
}

/** Replace the server's version of an item in a list. */
export function replaceInList(list: MediaItem[], item: MediaItem): MediaItem[] {
    return list.map((it) => (it.ID === item.ID ? item : it));
}

function definedOnly(fields: MediaQuickFields): MediaQuickFields {
    return Object.fromEntries(Object.entries(fields).filter(([, v]) => v !== undefined));
}
