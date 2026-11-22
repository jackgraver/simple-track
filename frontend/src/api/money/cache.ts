import type { AccountBalance, AccountTransaction } from "./types";

type Row = { ID: number };

/** The list without row `id`. Other rows keep their identity; a missing id returns the same array. */
export function removeById<T extends Row>(list: T[], id: number): T[] {
    return list.some((r) => r.ID === id) ? list.filter((r) => r.ID !== id) : list;
}

/**
 * Rollback when a sibling also changed the list: put the row from `snapshot` back at its
 * ordered place, leaving the sibling's other changes alone. Already present or not in the snapshot: unchanged.
 */
export function restoreById<T extends Row>(
    current: T[],
    snapshot: T[],
    id: number,
    compare: (a: T, b: T) => number,
): T[] {
    if (current.some((r) => r.ID === id)) return current;
    const row = snapshot.find((r) => r.ID === id);
    if (!row) return current;
    return [...current, row].sort(compare);
}

/** Server order for transactions and balances: newest date first, then newest id. */
export const compareTransactions = (a: AccountTransaction, b: AccountTransaction) => byDateDesc(a, b);
export const compareBalances = (a: AccountBalance, b: AccountBalance) => byDateDesc(a, b);

function byDateDesc(a: { date: string; ID: number }, b: { date: string; ID: number }) {
    if (a.date !== b.date) return a.date < b.date ? 1 : -1;
    return b.ID - a.ID;
}
