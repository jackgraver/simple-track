/**
 * Rollback for whole-value upserts (a mobility slot, cardio): every call sends the full value, so the
 * newest call wins. Stacked calls break snapshot rollback: the second call's snapshot already holds the
 * first call's optimistic value, which the server may never have saved. Instead:
 * - a call that fails while a newer call in the same chain is pending keeps the cache as is (the newer
 *   call carries the latest full value) and hands the chain the value the server still holds;
 * - the last call of a chain to fail puts back what the server holds: the latest success in the chain,
 *   else the first failed call's snapshot, else its own.
 * A chain is the run of calls for one mutation key and date while any of them is pending.
 */
import type { QueryClient, QueryKey } from "@tanstack/vue-query";

type Vars = { date: string };

/** Per client, per chain id: the value the server holds while the chain is still running. */
const serverValues = new WeakMap<QueryClient, Map<string, unknown>>();

function valuesFor(qc: QueryClient): Map<string, unknown> {
    let m = serverValues.get(qc);
    if (!m) serverValues.set(qc, (m = new Map()));
    return m;
}

/** Other calls of the chain still in flight (the failing/succeeding call itself is still 'pending'). */
function othersPending(qc: QueryClient, mutationKey: QueryKey, vars: Vars): boolean {
    return qc
        .getMutationCache()
        .findAll({ mutationKey, exact: true, status: "pending" })
        .some((m) => {
            const v = m.state.variables as Vars | undefined;
            const settled = (m.state.context as { settled?: boolean } | undefined)?.settled === true;
            return v !== vars && !settled && v?.date === vars.date;
        });
}

export function upsertChain<T>(qc: QueryClient, mutationKey: QueryKey) {
    const id = (vars: Vars) => `${JSON.stringify(mutationKey)}:${vars.date}`;
    return {
        /** Call from `reconcile` (before any early return) with the server's answer. */
        confirmed(vars: Vars, value: T): void {
            const values = valuesFor(qc);
            if (othersPending(qc, mutationKey, vars)) values.set(id(vars), value);
            else values.delete(id(vars));
        },
        /**
         * Call from `revert` with this call's snapshot value. Returns the value to put back, or
         * `keep` when a newer call of the chain is still pending.
         */
        rollback(vars: Vars, snapshotValue: T): { keep: true } | { keep: false; value: T } {
            const values = valuesFor(qc);
            const key = id(vars);
            if (othersPending(qc, mutationKey, vars)) {
                if (!values.has(key)) values.set(key, snapshotValue);
                return { keep: true };
            }
            const value = values.has(key) ? (values.get(key) as T) : snapshotValue;
            values.delete(key);
            return { keep: false, value };
        },
    };
}
