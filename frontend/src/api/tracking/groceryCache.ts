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

/** Swap the optimistic row for the server's (real ID, same client_id). Appends if it vanished. */
export function reconcileGroceryItem(cur: GroceryItem[], item: GroceryItem, clientId: string): GroceryItem[] {
    const known = cur.some((r) => r.client_id === clientId || r.ID === item.ID);
    if (!known) return [...cur, item];
    return cur.map((r) => (r.client_id === clientId || r.ID === item.ID ? item : r));
}

/** Completed items leave the active list. */
export function completeGroceryItemInList(old: GroceryItem[], v: { id: number }): GroceryItem[] {
    return old.filter((r) => r.ID !== v.id);
}

/** Inverse of completing: put the snapshot's row back at its old position unless something re-added it. */
export function restoreGroceryItem(cur: GroceryItem[], v: { id: number }, snapshot: GroceryItem[]): GroceryItem[] {
    if (cur.some((r) => r.ID === v.id)) return cur;
    const index = snapshot.findIndex((r) => r.ID === v.id);
    if (index < 0) return cur;
    const next = [...cur];
    next.splice(Math.min(index, next.length), 0, snapshot[index]);
    return next;
}
