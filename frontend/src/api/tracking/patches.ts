/** Pure cache patch/revert functions for the optimistic tracking mutations (T1). */
import type { BodyWeightLog, DrinkSizePreset, StepLog, WaterLog } from "~/api/tracking/types";

export type AddWaterVars = {
    client_id: string;
    date: string;
    amountOz: number;
    presetId?: number;
    /** Cached preset, so the optimistic row can show it before the server answers. */
    preset?: DrinkSizePreset | null;
};
/** `client_id` (when the entry has one) lets a delete target an entry that is still being created. */
export type DeleteWaterVars = { id: number; date: string; client_id?: string | null };

const isWaterRow = (r: WaterLog, v: DeleteWaterVars) => r.ID === v.id || (!!v.client_id && r.client_id === v.client_id);

/** Path ref for `DELETE /tracking/water/:ref`: the client_id when known, else the id. */
export function waterPathRef(v: DeleteWaterVars): string {
    return v.client_id ?? String(v.id);
}

/** Negative, deterministic per client_id: unique list key that can never collide with a server id. */
export function tempWaterId(clientId: string): number {
    return -(Number.parseInt(clientId.replace(/-/g, "").slice(0, 8), 16) + 1) || -1;
}

export function optimisticWaterLog(v: AddWaterVars, now = new Date()): WaterLog {
    return {
        ID: tempWaterId(v.client_id),
        client_id: v.client_id,
        date: v.date,
        amount_oz: v.amountOz,
        preset_id: v.presetId ?? null,
        preset: v.preset ?? null,
        CreatedAt: now.toISOString(),
    };
}

/** Idempotent upsert by client_id (a retry re-runs it on a row that is already there). */
export function addWaterPatch(old: WaterLog[], v: AddWaterVars): WaterLog[] {
    if (old.some((r) => r.client_id === v.client_id)) return old;
    return [...old, optimisticWaterLog(v)];
}

export function revertAddWater(cur: WaterLog[], v: AddWaterVars): WaterLog[] {
    return cur.filter((r) => r.client_id !== v.client_id);
}

/** Swap the optimistic row for the server's (merging only this entity, siblings stay untouched). */
export function reconcileWater(cur: WaterLog[], v: AddWaterVars, saved: WaterLog): WaterLog[] {
    // A row that is gone (deleted while the create was in flight) stays gone; the settle refetch decides.
    return cur.map((r) => (r.client_id === v.client_id ? { ...saved, client_id: v.client_id } : r));
}

export function deleteWaterPatch(old: WaterLog[], v: DeleteWaterVars): WaterLog[] {
    return old.filter((r) => !isWaterRow(r, v));
}

/** Put the deleted row back (from the snapshot) without touching other pending changes. */
export function revertDeleteWater(cur: WaterLog[], v: DeleteWaterVars, snapshot: WaterLog[]): WaterLog[] {
    if (cur.some((r) => isWaterRow(r, v))) return cur;
    const row = snapshot.find((r) => isWaterRow(r, v));
    return row ? [...cur, row] : cur;
}

// ── weight / steps: upsert by calendar day ───────────────────────────────────────────────────────

/** `YYYY-MM-DD` of a stored date (`2026-11-22T00:00:00Z` or plain). Compared as text, never via Date. */
export function dayKey(date: string): string {
    return date.slice(0, 10);
}

/** Replace the row for `date` (keeping its other fields) or insert one; list stays newest-first. */
export function upsertByDay<T extends { date: string }>(old: T[], date: string, make: (existing: T | undefined) => T): T[] {
    const key = dayKey(date);
    const existing = old.find((r) => dayKey(r.date) === key);
    const next = existing ? old.map((r) => (r === existing ? make(existing) : r)) : [...old, make(undefined)];
    return existing ? next : next.sort((a, b) => dayKey(b.date).localeCompare(dayKey(a.date)));
}

/** Restore the day's row from the snapshot (or remove it if it had none) without touching other days. */
export function revertUpsertByDay<T extends { date: string }>(cur: T[], date: string, snapshot: T[]): T[] {
    const key = dayKey(date);
    const before = snapshot.find((r) => dayKey(r.date) === key);
    if (!before) return cur.filter((r) => dayKey(r.date) !== key);
    return cur.some((r) => dayKey(r.date) === key)
        ? cur.map((r) => (dayKey(r.date) === key ? before : r))
        : [...cur, before].sort((a, b) => dayKey(b.date).localeCompare(dayKey(a.date)));
}

export const saveWeightPatch = (old: BodyWeightLog[], v: { date: string; weightLbs: number }): BodyWeightLog[] =>
    upsertByDay(old, v.date, (e) => ({ ...(e ?? { ID: 0, date: v.date }), weight_lbs: v.weightLbs }));

export const saveStepsPatch = (old: StepLog[], v: { date: string; steps: number }): StepLog[] =>
    upsertByDay(old, v.date, (e) => ({ ...(e ?? { ID: 0, date: v.date }), steps: v.steps }));

/** Write the server row for the day over the optimistic one. */
export function reconcileByDay<T extends { date: string }>(cur: T[], saved: T): T[] {
    return upsertByDay(cur, saved.date, () => saved);
}
