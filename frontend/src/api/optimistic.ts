/**
 * Shared optimistic-mutation helper (DESIGN.md §2). Every optimistic hook uses `useOptimisticMutation`;
 * wait-for-server hooks use `useServerMutation` for the same error/settle conventions.
 *
 * Concurrency model:
 * - Each mutation records which cache keys it patches (via `meta`), so helpers can ask "is anybody
 *   else still writing to this key?" across different hooks (`siblingCount`).
 * - Rollback restores the snapshot only when nobody else is writing the key and the cache still holds
 *   exactly the value this mutation wrote (a sibling that already settled, failed or succeeded, may
 *   have changed it since); otherwise it runs the patch's `revert`, or invalidates when there is none.
 * - Settle-time invalidation is skipped while siblings on the key are pending and queued instead; the
 *   last one to settle flushes the queue, whatever hook it belongs to.
 * - A success that settles while siblings are pending may have skipped its reconcile, so the key is
 *   remembered; if the last writer then fails and rolls back, it invalidates that key.
 * - Per-entity ordering (`scope`) is a promise chain per scope id, not TanStack's `scope` option,
 *   because TanStack's scope is fixed per `useMutation` instance (DECISIONS.md, F6). The cache patch
 *   still applies immediately; only the request waits for the previous same-scope mutation to settle.
 */
import {
    hashKey,
    partialMatchKey,
    useMutation,
    useMutationState,
    useQueryClient,
    type Mutation,
    type MutationFunctionContext,
    type MutationScope,
    type MutationStatus,
    type QueryClient,
    type QueryKey,
    type UseMutationReturnType,
} from "@tanstack/vue-query";
import { computed, type Ref } from "vue";

import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";

/** One cache entry a mutation patches. `TCache` is the cached value's type at `queryKey`. */
export type CachePatch<TVars, TCache = any> = {
    /** Exact key of the entry (no prefixes: cancel, snapshot and restore are per entry). */
    queryKey: QueryKey;
    /** Pure: the next cache value. Not called when the entry is not cached (nothing on screen to patch). */
    apply: (old: TCache, vars: TVars) => TCache;
    /**
     * Pure inverse, used on rollback when another mutation also patched this key since our snapshot,
     * so restoring the snapshot would wipe its change. Usually "remove/restore entity by client_id".
     * Without it the helper invalidates the key instead.
     */
    revert?: (current: TCache, vars: TVars, snapshot: TCache) => TCache;
};

export type FailureMode = "rollback" | "mark";

/** Passed to `reconcile` so it can decide between replacing an aggregate and merging one entity. */
export type ReconcileHelpers = {
    /** Other in-flight mutations (not this one) patching `queryKey` or a key overlapping it by prefix. */
    siblingsPending: (queryKey: QueryKey) => number;
};

type BaseOptions<TVars, TData> = {
    /** Stable key, e.g. `workoutKeys.mutations.logSet()`. Used by `usePendingIds` and sibling checks. */
    mutationKey: QueryKey;
    mutationFn: (vars: TVars) => Promise<TData>;
    /**
     * Write the server's answer into the cache (preferred over invalidation). While siblings are
     * pending, merge only this mutation's entity (by `client_id`); replace aggregates only when
     * `siblingsPending(key) === 0`.
     */
    reconcile?: (data: TData, vars: TVars, qc: QueryClient, helpers: ReconcileHelpers) => void;
    /** Keys to invalidate after settle. Skipped per key while a sibling mutation on it is still pending. */
    invalidate?: (vars: TVars, data: TData | undefined) => QueryKey[];
    /**
     * Serialise requests that must not reorder (create a set, then edit it): same id → requests run one
     * after another, across hooks. Optimistic patches still apply immediately.
     */
    scope?: (vars: TVars) => MutationScope;
    /**
     * Retries for idempotent writes (client_id creates, upserts, PATCH). Default 0. A function can
     * skip errors that won't change on retry (4xx), so a scope isn't held up by backoff for nothing.
     */
    retry?: number | ((failureCount: number, error: unknown) => boolean);
    retryDelay?: number;
    /** Use this client instead of the injected one (tests, code outside component setup). */
    queryClient?: QueryClient;
};

export type OptimisticMutationOptions<TVars, TData> = BaseOptions<TVars, TData> & {
    /** Cache entries to patch before the request. */
    optimistic: (vars: TVars) => CachePatch<TVars>[];
    /**
     * 'rollback' (default): undo the patch and toast with Retry.
     * 'mark': keep the patched row; the mutation stays in the cache with status 'error' (read it with
     * `usePendingIds(key, 'error')`) until `retry(vars)` or `dismiss(vars)`. In this mode `apply` must
     * be idempotent (upsert by client_id), because `retry` re-runs it on a row that is already there.
     */
    onFailure?: FailureMode;
    /** Toast text. Defaults to the API error body via `getApiErrorMessage`. */
    errorMessage?: (err: unknown, vars: TVars) => string;
};

export type ServerMutationOptions<TVars, TData> = BaseOptions<TVars, TData> & {
    /** Toast text, or `false` when the caller renders the error itself (inline form error). */
    errorMessage?: ((err: unknown, vars: TVars) => string) | false;
};

/** `next` is the value this mutation wrote, so rollback can tell whether anyone changed it since. */
type Snapshot = { patch: CachePatch<unknown>; prev: unknown; next?: unknown; applied: boolean };

/** onMutate result. Its identity marks "self" when counting siblings. */
export type OptimisticContext = {
    snapshots: Snapshot[];
    /** Set when this mutation reaches onSettled, so siblings stop counting it as pending. */
    settled: boolean;
    release?: () => void;
};

export type OptimisticMutationReturn<TVars, TData> = UseMutationReturnType<TData, unknown, TVars, OptimisticContext> & {
    /** Re-run `vars` (same client_id), dropping any failed mutation recorded for them first. */
    retry: (vars: TVars) => void;
    /** Forget a failed ('mark' mode) mutation for `vars`; the caller removes the row from the cache. */
    dismiss: (vars: TVars) => void;
};

const KEYS_META = "optimisticKeys";
const FAILURE_META = "onFailure";
const ERROR_TOAST_MS = 6000;

type Meta = { handled: true; [KEYS_META]?: (vars: unknown) => QueryKey[]; [FAILURE_META]?: FailureMode };
// The cache holds mutations of every hook's types.
type AnyMutation = Mutation<any, any, any, any>;

/** Keys overlap when either is a prefix of the other (an invalidate of one refetches the other). */
function keysOverlap(a: QueryKey, b: QueryKey): boolean {
    return partialMatchKey(a, b) || partialMatchKey(b, a);
}

function touchedKeys(m: AnyMutation): QueryKey[] {
    const fn = (m.options.meta as Meta | undefined)?.[KEYS_META];
    if (!fn || m.state.variables === undefined) return [];
    try {
        return fn(m.state.variables);
    } catch {
        return [];
    }
}

/**
 * Other mutations (any hook) patching `key`, excluding `self`.
 * Pending ones count until they reach onSettled. With `includeMarked`, failed 'mark' mutations count
 * too: their rows are still in the cache, so a snapshot restore would wipe them.
 */
export function siblingCount(
    qc: QueryClient,
    key: QueryKey,
    self?: OptimisticContext,
    { includeMarked = false } = {},
): number {
    return qc.getMutationCache().getAll().filter((m) => {
        const ctx = m.state.context as OptimisticContext | undefined;
        if (self && ctx === self) return false;
        const pending = m.state.status === "pending" && !ctx?.settled;
        const marked =
            includeMarked && m.state.status === "error" && (m.options.meta as Meta | undefined)?.[FAILURE_META] === "mark";
        if (!pending && !marked) return false;
        return touchedKeys(m).some((k) => keysOverlap(k, key));
    }).length;
}

// ── Per-entity scopes: a promise chain per scope id ──────────────────────────────────────────────

const scopeTails = new Map<string, Promise<void>>();
/** The gate each execution waits on. Keyed by TanStack's per-execution MutationFunctionContext object,
 *  which is the same object passed to onMutate and mutationFn (covered by a test). */
const gates = new WeakMap<object, Promise<void>>();

function enterScope(id: string, fnCtx: object): () => void {
    const prev = scopeTails.get(id) ?? Promise.resolve();
    let done!: () => void;
    const tail = new Promise<void>((r) => (done = r));
    scopeTails.set(id, tail);
    gates.set(fnCtx, prev);
    return () => {
        done();
        if (scopeTails.get(id) === tail) scopeTails.delete(id);
    };
}

// ── Shared core ──────────────────────────────────────────────────────────────────────────────────

type CoreOptions<TVars, TData> = BaseOptions<TVars, TData> & {
    optimistic?: (vars: TVars) => CachePatch<TVars>[];
    onFailure?: FailureMode;
    errorMessage?: ((err: unknown, vars: TVars) => string) | false;
    retryAction: boolean;
};

function sameCall<TVars>(a: unknown, b: TVars): boolean {
    if (a === b) return true;
    const ida = (a as { client_id?: unknown } | null)?.client_id;
    const idb = (b as { client_id?: unknown } | null)?.client_id;
    return ida !== undefined && ida === idb;
}

function useCoreMutation<TVars, TData>(opts: CoreOptions<TVars, TData>): OptimisticMutationReturn<TVars, TData> {
    const qc = opts.queryClient ?? useQueryClient();
    const mark = opts.onFailure === "mark";

    const meta: Meta = { handled: true, [FAILURE_META]: opts.onFailure ?? "rollback" };
    if (opts.optimistic) {
        const optimistic = opts.optimistic;
        meta[KEYS_META] = (vars) => optimistic(vars as TVars).map((p) => p.queryKey);
    }

    const mutation = useMutation<TData, unknown, TVars, OptimisticContext>(
        {
            mutationKey: opts.mutationKey,
            meta,
            retry: opts.retry ?? 0,
            ...(opts.retryDelay !== undefined && { retryDelay: opts.retryDelay }),
            // Failed 'mark' mutations are the failed-row state, so they must outlive the default 5 min.
            ...(mark && { gcTime: Infinity }),
            mutationFn: async (vars, fnCtx: MutationFunctionContext) => {
                const gate = gates.get(fnCtx);
                if (gate) await gate;
                return opts.mutationFn(vars);
            },
            onMutate: async (vars, fnCtx) => {
                const ctx: OptimisticContext = { snapshots: [], settled: false };
                // Synchronous, before any await, so scope order = call order.
                if (opts.scope) ctx.release = enterScope(String(opts.scope(vars).id), fnCtx);
                try {
                    const patches = (opts.optimistic?.(vars) ?? []) as CachePatch<unknown>[];
                    // Stop in-flight refetches from overwriting the optimistic value when they land.
                    await Promise.all(patches.map((p) => qc.cancelQueries({ queryKey: p.queryKey, exact: true })));
                    for (const patch of patches) {
                        const prev = qc.getQueryData(patch.queryKey);
                        const applied = prev !== undefined;
                        // What the cache stored (structural sharing may keep a different object than `apply` returned).
                        const next = applied ? qc.setQueryData(patch.queryKey, patch.apply(prev, vars)) : undefined;
                        ctx.snapshots.push({ patch, prev, next, applied });
                    }
                } catch (err) {
                    ctx.release?.();
                    throw err;
                }
                return ctx;
            },
            onError: (err, vars, ctx) => {
                if (!mark && ctx) rollback(qc, ctx, vars);
                if (mark || opts.errorMessage === false) return;
                const message = opts.errorMessage?.(err, vars) ?? getApiErrorMessage(err, "Couldn’t save");
                if (opts.retryAction) {
                    toast.push(message, "error", {
                        action: { label: "Retry", onClick: () => retry(vars) },
                        duration: ERROR_TOAST_MS,
                    });
                } else {
                    toast.push(message, "error");
                }
            },
            onSuccess: (data, vars, ctx) => {
                // A reconcile with siblings pending may only merge (or skip); if the last writer then
                // fails, nobody would write the server's answer, so remember the key for its rollback.
                for (const { patch } of ctx?.snapshots ?? []) {
                    if (siblingCount(qc, patch.queryKey, ctx) > 0) remember(unreconciled, qc, patch.queryKey);
                }
                opts.reconcile?.(data, vars, qc, { siblingsPending: (key) => siblingCount(qc, key, ctx) });
            },
            onSettled: (data, err, vars, ctx) => {
                if (ctx) ctx.settled = true;
                ctx?.release?.();
                if (!err) {
                    // The last writer succeeded: its reconcile ran with no siblings pending.
                    for (const { patch } of ctx?.snapshots ?? []) {
                        if (siblingCount(qc, patch.queryKey, ctx) === 0) forget(unreconciled, qc, patch.queryKey);
                    }
                }
                for (const key of opts.invalidate?.(vars, data) ?? []) {
                    // Refetching while siblings are pending would briefly drop their optimistic rows.
                    if (siblingCount(qc, key, ctx) === 0) {
                        forget(deferredInvalidations, qc, key);
                        void qc.invalidateQueries({ queryKey: key });
                    } else {
                        remember(deferredInvalidations, qc, key);
                    }
                }
                // Keys other hooks put off: the siblings may not invalidate them themselves.
                for (const key of drain(deferredInvalidations, qc, ctx)) void qc.invalidateQueries({ queryKey: key });
            },
        },
        qc,
    );

    function failedFor(vars: TVars): AnyMutation[] {
        return qc
            .getMutationCache()
            .findAll({ mutationKey: opts.mutationKey, exact: true, status: "error" })
            .filter((m) => sameCall(m.state.variables, vars)) as AnyMutation[];
    }

    function dismiss(vars: TVars) {
        for (const m of failedFor(vars)) qc.getMutationCache().remove(m);
    }

    function retry(vars: TVars) {
        dismiss(vars);
        mutation.mutate(vars);
    }

    return Object.assign(mutation, { retry, dismiss });
}

// ── Work put off while siblings were pending, per client ─────────────────────────────────────────

type KeySet = WeakMap<QueryClient, Map<string, QueryKey>>;
/** Patch keys where a success settled with siblings pending (its reconcile may have been skipped). */
const unreconciled: KeySet = new WeakMap();
/** Settle-time invalidations skipped because siblings were pending. */
const deferredInvalidations: KeySet = new WeakMap();

function remember(set: KeySet, qc: QueryClient, key: QueryKey) {
    let m = set.get(qc);
    if (!m) set.set(qc, (m = new Map()));
    m.set(hashKey(key), key);
}

/** Removes `key`; reports whether it was there. */
function forget(set: KeySet, qc: QueryClient, key: QueryKey): boolean {
    return set.get(qc)?.delete(hashKey(key)) ?? false;
}

/** Removes and returns the keys nobody (other than `self`) is still writing. */
function drain(set: KeySet, qc: QueryClient, self?: OptimisticContext): QueryKey[] {
    const m = set.get(qc);
    if (!m) return [];
    const out: QueryKey[] = [];
    for (const [h, key] of m) {
        if (siblingCount(qc, key, self) > 0) continue;
        m.delete(h);
        out.push(key);
    }
    return out;
}

function rollback<TVars>(qc: QueryClient, ctx: OptimisticContext, vars: TVars) {
    // Undo in reverse order so multiple patches on related keys unwind cleanly.
    for (const { patch, prev, next, applied } of [...ctx.snapshots].reverse()) {
        if (!applied) continue;
        const others = siblingCount(qc, patch.queryKey, ctx, { includeMarked: true });
        // A sibling that settled after our patch (reverted after its own failure, or reconciled) may
        // have changed the entry: the snapshot is stale then, so only undo our own change.
        const untouched = qc.getQueryData(patch.queryKey) === next;
        // We are the last writer and a sibling's success may never have reached the cache: refetch.
        const refetch = others === 0 && forget(unreconciled, qc, patch.queryKey);
        if (others === 0 && untouched) {
            qc.setQueryData(patch.queryKey, prev);
        } else if (patch.revert) {
            const revert = patch.revert;
            qc.setQueryData(patch.queryKey, (cur: unknown) => (cur === undefined ? cur : revert(cur, vars, prev)));
        } else {
            void qc.invalidateQueries({ queryKey: patch.queryKey, exact: true });
            continue;
        }
        if (refetch) void qc.invalidateQueries({ queryKey: patch.queryKey, exact: true });
    }
}

// ── Public hooks ─────────────────────────────────────────────────────────────────────────────────

/** Optimistic write: patch the cache now, reconcile with the server's answer, roll back or mark on failure. */
export function useOptimisticMutation<TVars, TData>(
    opts: OptimisticMutationOptions<TVars, TData>,
): OptimisticMutationReturn<TVars, TData> {
    return useCoreMutation({ ...opts, retryAction: true });
}

/**
 * Wait-for-server write with the same conventions: one error toast (no global duplicate),
 * `reconcile` before `invalidate`, sibling-aware invalidation and optional per-entity `scope`.
 * The toast has no Retry: forms keep their inputs and the user resubmits.
 */
export function useServerMutation<TVars, TData>(
    opts: ServerMutationOptions<TVars, TData>,
): OptimisticMutationReturn<TVars, TData> {
    return useCoreMutation({ ...opts, retryAction: false });
}

/**
 * `client_id`s of this mutation key's calls in `status` (default 'pending'). Pending state is derived
 * from the mutation cache, never stored in cache data. Use `'error'` for failed 'mark' rows.
 */
export function usePendingIds(
    mutationKey: QueryKey,
    status: MutationStatus = "pending",
    queryClient?: QueryClient,
): Readonly<Ref<Set<string>>> {
    const ids = useMutationState(
        {
            filters: { mutationKey, exact: true, status },
            select: (m) => (m.state.variables as { client_id?: string } | undefined)?.client_id,
        },
        queryClient,
    );
    return computed(() => new Set(ids.value.filter((id): id is string => typeof id === "string")));
}
