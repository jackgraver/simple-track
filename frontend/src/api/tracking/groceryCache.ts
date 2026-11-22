/** Pure cache patches for the grocery list (`trackingKeys.groceryItems`). Used by the optimistic hooks. */
import type { GroceryItem } from "~/api/tracking/types";

/**
 * Stable negative id for a row that has no server id yet, derived from its client_id so `apply`
 * stays pure and idempotent. Real ids are positive, so the two never collide.
 */
export function tempGroceryId(clientId: string): number {
    let h = 0;
    for (let i = 0; i < clientId.length; i++) h = (Math.imul(h, 31) + clientId.charCodeAt(i)) | 0;
    return -(Math.abs(h) + 1);
}

export function isTempGroceryId(id: number): boolean {
    return id < 0;
}

/** Upsert by client_id: append a not-yet-saved row (no duplicate when re-applied). */
export function addGroceryItem(old: GroceryItem[], v: { client_id: string; name: string }): GroceryItem[] {
    const row: GroceryItem = {
        ID: tempGroceryId(v.client_id),
        client_id: v.client_id,
        name: v.name,
        completed_at: null,
    };
    return [...old.filter((r) => r.client_id !== v.client_id), row];
}

export function removeGroceryItemByClientId(cur: GroceryItem[], v: { client_id: string }): GroceryItem[] {
    return cur.filter((r) => r.client_id !== v.client_id);
}

/**
 * Swap the optimistic row for the server's (real ID, same client_id). A row that is gone (ticked
 * off while its create was in flight) stays gone: the settle-time refetch has the final say.
 */
export function reconcileGroceryItem(cur: GroceryItem[], item: GroceryItem, clientId: string): GroceryItem[] {
    return cur.map((r) => (r.client_id === clientId || r.ID === item.ID ? item : r));
}

/** Identifies the row to complete: its server id, or its client_id while the id is still temporary. */
export type GroceryItemRef = { id: number; client_id?: string | null };

function matches(r: GroceryItem, v: GroceryItemRef): boolean {
    return r.ID === v.id || (!!v.client_id && r.client_id === v.client_id);
}

/** Completed items leave the active list. */
export function completeGroceryItemInList(old: GroceryItem[], v: GroceryItemRef): GroceryItem[] {
    return old.filter((r) => !matches(r, v));
}

/** Inverse of completing: put the snapshot's row back at its old position unless something re-added it. */
export function restoreGroceryItem(cur: GroceryItem[], v: GroceryItemRef, snapshot: GroceryItem[]): GroceryItem[] {
    if (cur.some((r) => matches(r, v))) return cur;
    const index = snapshot.findIndex((r) => matches(r, v));
    if (index < 0) return cur;
    const next = [...cur];
    next.splice(Math.min(index, next.length), 0, snapshot[index]);
    return next;
}

/** Path ref for complete/delete: the client_id when known (works before the create returns), else the id. */
export function groceryPathRef(v: GroceryItemRef): string {
    return v.client_id ?? String(v.id);
}
