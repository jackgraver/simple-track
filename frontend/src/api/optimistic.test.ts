import type { QueryClient } from "@tanstack/vue-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, type EffectScope } from "vue";

import { createQueryClient } from "./queryClient";
import {
    siblingCount,
    useOptimisticMutation,
    usePendingIds,
    useServerMutation,
    type CachePatch,
    type OptimisticMutationOptions,
} from "./optimistic";

const push = vi.hoisted(() => vi.fn());
vi.mock("~/composables/toast/useToast", () => ({ toast: { push } }));

type Row = { client_id: string; ID: number; name: string };
type Vars = { client_id: string; name: string };
const LIST = ["rows", "list"] as const;

function deferred<T>() {
    let resolve!: (v: T) => void;
    let reject!: (e: unknown) => void;
    const promise = new Promise<T>((res, rej) => {
        resolve = res;
        reject = rej;
    });
    return { promise, resolve, reject };
}

const flush = () => new Promise((r) => setTimeout(r, 0));

/** Upsert by client_id: idempotent, as 'mark' mode requires. */
const addRow: CachePatch<Vars, Row[]> = {
    queryKey: LIST,
    apply: (old, v) => [...old.filter((r) => r.client_id !== v.client_id), { client_id: v.client_id, ID: 0, name: v.name }],
    revert: (cur, v) => cur.filter((r) => r.client_id !== v.client_id),
};

let scope: EffectScope;
let qc: QueryClient;
let notify: ReturnType<typeof vi.fn<(m: string) => void>>;

beforeEach(() => {
    push.mockClear();
    notify = vi.fn<(m: string) => void>();
    qc = createQueryClient(notify);
    scope = effectScope();
});
afterEach(() => scope.stop());

/** A controllable server: each call returns a deferred the test settles. */
function server() {
    const calls: { vars: Vars; d: ReturnType<typeof deferred<Row>> }[] = [];
    const fn = vi.fn((vars: Vars) => {
        const d = deferred<Row>();
        calls.push({ vars, d });
        return d.promise;
    });
    return { fn, calls };
}

function useAdd(
    srv: ReturnType<typeof server>,
    extra: Partial<OptimisticMutationOptions<Vars, Row>> = {},
    patch: CachePatch<Vars, Row[]> = addRow,
) {
    return scope.run(() =>
        useOptimisticMutation<Vars, Row>({
            mutationKey: ["rows", "add"],
            mutationFn: srv.fn,
            optimistic: () => [patch],
            reconcile: (saved, v, client) =>
                client.setQueryData<Row[]>(LIST, (rows) => rows?.map((r) => (r.client_id === v.client_id ? saved : r))),
            queryClient: qc,
            ...extra,
        }),
    )!;
}

const rows = () => qc.getQueryData<Row[]>(LIST)!;
const seed = () => qc.setQueryData<Row[]>(LIST, [{ client_id: "s", ID: 1, name: "seed" }]);

describe("useOptimisticMutation: apply and rollback", () => {
    it("patches the cache before the request and reconciles with the response", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv);
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a"]);
        expect(srv.fn).toHaveBeenCalledTimes(1);

        srv.calls[0].d.resolve({ client_id: "a", ID: 42, name: "A" });
        await flush();
        expect(rows()[1]).toEqual({ client_id: "a", ID: 42, name: "A" });
        expect(push).not.toHaveBeenCalled();
    });

    it("restores the snapshot on failure and toasts with a Retry that reuses the variables", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv);
        const vars = { client_id: "a", name: "A" };
        m.mutate(vars);
        await flush();
        srv.calls[0].d.reject(new Error("offline"));
        await flush();

        expect(rows().map((r) => r.client_id)).toEqual(["s"]);
        expect(push).toHaveBeenCalledTimes(1);
        const [message, tone, opts] = push.mock.calls[0];
        expect(message).toBe("offline");
        expect(tone).toBe("error");
        expect(opts.action.label).toBe("Retry");
        expect(opts.duration).toBe(6000);
        // The helper sets meta.handled, so the global MutationCache toast stays quiet.
        expect(notify).not.toHaveBeenCalled();

        opts.action.onClick();
        await flush();
        expect(srv.calls[1].vars).toBe(vars);
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a"]);
        // The failed attempt is dropped from the mutation cache on retry.
        expect(qc.getMutationCache().findAll({ status: "error" })).toHaveLength(0);
    });

    it("uses errorMessage for the toast", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { errorMessage: (_e, v) => `Couldn’t add ${v.name}` });
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(push.mock.calls[0][0]).toBe("Couldn’t add A");
    });

    it("does not call apply when the entry is not cached", async () => {
        const srv = server();
        const apply = vi.fn(addRow.apply);
        const m = useAdd(srv, {}, { ...addRow, apply });
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        expect(apply).not.toHaveBeenCalled();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(qc.getQueryData(LIST)).toBeUndefined();
    });

    it("cancels an in-flight refetch so it cannot overwrite the optimistic value", async () => {
        seed();
        const refetch = deferred<Row[]>();
        void qc.fetchQuery({ queryKey: LIST, queryFn: () => refetch.promise, staleTime: 0 }).catch(() => {});
        const srv = server();
        useAdd(srv).mutate({ client_id: "a", name: "A" });
        await flush();
        refetch.resolve([]);
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a"]);
    });
});

describe("rollback with a sibling pending", () => {
    it("reverts only its own entity when another mutation on the key is in flight", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv);
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a", "b"]);

        srv.calls[0].d.reject(new Error("x"));
        await flush();
        // A snapshot restore would also drop "b" (snapshot ["s"]); revert keeps it.
        expect(rows().map((r) => r.client_id)).toEqual(["s", "b"]);
    });

    it("counts siblings from other hooks patching the same key", async () => {
        seed();
        const srvA = server();
        const srvB = server();
        const a = useAdd(srvA);
        const b = useAdd(srvB, { mutationKey: ["rows", "other"] });
        a.mutate({ client_id: "a", name: "A" });
        b.mutate({ client_id: "b", name: "B" });
        await flush();
        expect(siblingCount(qc, LIST)).toBe(2);
        srvA.calls[0].d.reject(new Error("x"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "b"]);
    });

    it("invalidates instead when the patch has no revert", async () => {
        seed();
        const srv = server();
        const { revert: _drop, ...noRevert } = addRow;
        const m = useAdd(srv, {}, noRevert);
        const inv = vi.spyOn(qc, "invalidateQueries");
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(inv).toHaveBeenCalledWith({ queryKey: LIST, exact: true });
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a", "b"]);
    });

    it("restores the snapshot once the sibling has settled, if the cache still holds our value", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { reconcile: () => {} }); // b's success leaves the cache untouched
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s"]);
    });

    // Review finding 2: a later sibling that already succeeded must survive the earlier one failing last.
    it("keeps a sibling's saved row when an earlier mutation fails after it", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv);
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[1].d.resolve({ client_id: "b", ID: 7, name: "B" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        // "a" snapshotted ["s"]; restoring it would wipe the saved "b".
        expect(rows()).toEqual([
            { client_id: "s", ID: 1, name: "seed" },
            { client_id: "b", ID: 7, name: "B" },
        ]);
    });

    // Review finding 1: the last failure must not restore a snapshot holding an earlier, failed row.
    it("does not bring back a row that already failed when the next one fails too", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv);
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "b"]);
        srv.calls[1].d.reject(new Error("x"));
        await flush();
        // b's snapshot was ["s", "a"]; "a" was never saved.
        expect(rows().map((r) => r.client_id)).toEqual(["s"]);
    });

    it("still restores the snapshot when patches have no revert and nobody changed the entry", async () => {
        seed();
        const srv = server();
        const { revert: _drop, ...noRevert } = addRow;
        const m = useAdd(srv, {}, noRevert);
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s"]);
    });
});

// Review finding 4: work skipped while siblings were pending must happen when the last one settles.
describe("deferred work when the last writer fails", () => {
    /** Like diet's reconcileDay: replace the entry only when nobody else is writing it, else skip. */
    const skipWhileSiblings: Partial<OptimisticMutationOptions<Vars, Row>> = {
        reconcile: (saved, v, client, h) => {
            if (h.siblingsPending(LIST) > 0) return;
            client.setQueryData<Row[]>(LIST, (r) => r?.map((x) => (x.client_id === v.client_id ? saved : x)));
        },
    };

    it("refetches the entry when a sibling's reconcile was skipped and the last writer rolls back", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, skipWhileSiblings);
        const inv = vi.spyOn(qc, "invalidateQueries");
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        await flush();
        // a's reconcile skipped: its row is still the optimistic one.
        expect(rows().map((r) => r.ID)).toEqual([1, 0, 0]);
        expect(inv).not.toHaveBeenCalled();

        srv.calls[1].d.reject(new Error("x"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a"]);
        expect(inv).toHaveBeenCalledWith({ queryKey: LIST, exact: true });
    });

    it("does not refetch when the last writer succeeds (its reconcile replaced the entry)", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, skipWhileSiblings);
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        await flush();
        srv.calls[1].d.resolve({ client_id: "b", ID: 20, name: "B" });
        await flush();
        const inv = vi.spyOn(qc, "invalidateQueries");
        // A later lone failure has nothing pending to catch up on.
        m.mutate({ client_id: "c", name: "C" });
        await flush();
        srv.calls[2].d.reject(new Error("x"));
        await flush();
        expect(inv).not.toHaveBeenCalled();
    });

    it("runs a sibling hook's skipped invalidation when another hook settles last", async () => {
        seed();
        const srvA = server();
        const srvB = server();
        const a = useAdd(srvA, { invalidate: () => [LIST] });
        const b = useAdd(srvB, { mutationKey: ["rows", "other"] }); // invalidates nothing itself
        const inv = vi.spyOn(qc, "invalidateQueries");
        a.mutate({ client_id: "a", name: "A" });
        b.mutate({ client_id: "b", name: "B" });
        await flush();
        srvA.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        await flush();
        expect(inv).not.toHaveBeenCalled();
        srvB.calls[0].d.resolve({ client_id: "b", ID: 20, name: "B" });
        await flush();
        expect(inv).toHaveBeenCalledTimes(1);
        expect(inv).toHaveBeenCalledWith({ queryKey: LIST });
    });
});

describe("reconcile", () => {
    it("merges each entity by client_id while siblings are pending", async () => {
        seed();
        const srv = server();
        const seen: number[] = [];
        const m = useAdd(srv, {
            reconcile: (saved, v, client, h) => {
                seen.push(h.siblingsPending(LIST));
                client.setQueryData<Row[]>(LIST, (r) => r?.map((x) => (x.client_id === v.client_id ? saved : x)));
            },
        });
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[1].d.resolve({ client_id: "b", ID: 20, name: "B" });
        await flush();
        expect(rows()).toEqual([
            { client_id: "s", ID: 1, name: "seed" },
            { client_id: "a", ID: 0, name: "A" },
            { client_id: "b", ID: 20, name: "B" },
        ]);
        srv.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        await flush();
        expect(rows().map((r) => r.ID)).toEqual([1, 10, 20]);
        // The first reconcile saw "a" in flight; the last one saw nobody (it may replace aggregates).
        expect(seen).toEqual([1, 0]);
    });
});

describe("settle-time invalidation", () => {
    const OTHER = ["rows", "stats"] as const;

    it("is skipped while a sibling on the key is pending, and runs once when the last settles", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { invalidate: () => [LIST, OTHER] });
        const inv = vi.spyOn(qc, "invalidateQueries");
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();

        srv.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        await flush();
        // LIST has "b" in flight; OTHER is not patched by anyone, so it refetches.
        expect(inv.mock.calls.map((c) => c[0])).toEqual([{ queryKey: OTHER }]);

        srv.calls[1].d.resolve({ client_id: "b", ID: 20, name: "B" });
        await flush();
        expect(inv.mock.calls.filter((c) => c[0]?.queryKey === LIST)).toHaveLength(1);
    });

    it("invalidates exactly once when siblings settle in the same tick", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { invalidate: () => [LIST] });
        const inv = vi.spyOn(qc, "invalidateQueries");
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        srv.calls[1].d.resolve({ client_id: "b", ID: 20, name: "B" });
        await flush();
        expect(inv).toHaveBeenCalledTimes(1);
    });

    it("treats a prefix key as overlapping (invalidating ['rows'] would refetch the list)", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { invalidate: () => [["rows"]] });
        const inv = vi.spyOn(qc, "invalidateQueries");
        m.mutate({ client_id: "a", name: "A" });
        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[0].d.resolve({ client_id: "a", ID: 10, name: "A" });
        await flush();
        expect(inv).not.toHaveBeenCalled();
        srv.calls[1].d.resolve({ client_id: "b", ID: 20, name: "B" });
        await flush();
        expect(inv).toHaveBeenCalledTimes(1);
    });

    it("also runs after a failure", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { invalidate: () => [OTHER] });
        const inv = vi.spyOn(qc, "invalidateQueries");
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(inv).toHaveBeenCalledWith({ queryKey: OTHER });
    });
});

describe("mark mode", () => {
    it("keeps the row, skips the toast and exposes pending/failed ids", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { onFailure: "mark" });
        const pending = scope.run(() => usePendingIds(["rows", "add"], "pending", qc))!;
        const failed = scope.run(() => usePendingIds(["rows", "add"], "error", qc))!;

        const vars = { client_id: "a", name: "A" };
        m.mutate(vars);
        await flush();
        expect([...pending.value]).toEqual(["a"]);
        expect(failed.value.size).toBe(0);

        srv.calls[0].d.reject(new Error("offline"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a"]);
        expect(push).not.toHaveBeenCalled();
        expect(notify).not.toHaveBeenCalled();
        expect(pending.value.size).toBe(0);
        expect([...failed.value]).toEqual(["a"]);

        // Retry re-applies (idempotent upsert: no duplicate) and clears the failed entry.
        m.retry({ ...vars });
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "a"]);
        expect(failed.value.size).toBe(0);
        expect([...pending.value]).toEqual(["a"]);
        srv.calls[1].d.resolve({ client_id: "a", ID: 5, name: "A" });
        await flush();
        expect(rows()[1].ID).toBe(5);
        expect(pending.value.size).toBe(0);
    });

    it("dismiss forgets the failure", async () => {
        seed();
        const srv = server();
        const m = useAdd(srv, { onFailure: "mark" });
        const failed = scope.run(() => usePendingIds(["rows", "add"], "error", qc))!;
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(failed.value.size).toBe(1);
        m.dismiss({ client_id: "a", name: "A" });
        await flush();
        expect(failed.value.size).toBe(0);
    });

    it("failed marked rows survive a sibling's rollback (counted as siblings)", async () => {
        seed();
        const srvMark = server();
        const srvRoll = server();
        const roll = useAdd(srvRoll, { mutationKey: ["rows", "roll"] });
        const markM = useAdd(srvMark, { onFailure: "mark" });
        roll.mutate({ client_id: "r", name: "R" });
        await flush();
        markM.mutate({ client_id: "m", name: "M" });
        await flush();
        srvMark.calls[0].d.reject(new Error("x"));
        await flush();
        srvRoll.calls[0].d.reject(new Error("x"));
        await flush();
        expect(rows().map((r) => r.client_id)).toEqual(["s", "m"]);
    });
});

describe("per-entity scope", () => {
    type EditVars = Vars;
    const scopeFn = (v: Vars) => ({ id: `row:${v.client_id}` });

    it("serialises requests for one entity across hooks while patches apply immediately", async () => {
        seed();
        const create = server();
        const edit = server();
        const c = useAdd(create, { scope: scopeFn });
        const e = scope.run(() =>
            useOptimisticMutation<EditVars, Row>({
                mutationKey: ["rows", "edit"],
                mutationFn: edit.fn,
                optimistic: () => [
                    {
                        queryKey: LIST,
                        apply: (old: Row[], v) => old.map((r) => (r.client_id === v.client_id ? { ...r, name: v.name } : r)),
                    },
                ],
                scope: scopeFn,
                queryClient: qc,
            }),
        )!;

        c.mutate({ client_id: "a", name: "A" });
        e.mutate({ client_id: "a", name: "A2" });
        c.mutate({ client_id: "b", name: "B" }); // other entity: not blocked
        await flush();
        expect(rows().map((r) => r.name)).toEqual(["seed", "A2", "B"]);
        expect(create.fn).toHaveBeenCalledTimes(2);
        expect(edit.fn).not.toHaveBeenCalled();

        create.calls[0].d.resolve({ client_id: "a", ID: 9, name: "A" });
        await flush();
        expect(edit.fn).toHaveBeenCalledTimes(1);
        expect(edit.calls[0].vars.name).toBe("A2");
        // Settle the rest: scope chains are module-level, so a hung request would block later tests.
        edit.calls[0].d.resolve({ client_id: "a", ID: 9, name: "A2" });
        create.calls[1].d.resolve({ client_id: "b", ID: 10, name: "B" });
        await flush();
    });

    it("runs the next request after the previous one fails (including its retries)", async () => {
        seed();
        const order: string[] = [];
        let attempts = 0;
        const m = scope.run(() =>
            useOptimisticMutation<Vars, Row>({
                mutationKey: ["rows", "add"],
                mutationFn: async (v) => {
                    order.push(v.name);
                    if (v.name === "first") {
                        attempts++;
                        throw new Error("x");
                    }
                    return { ...v, ID: 1 };
                },
                optimistic: () => [addRow],
                scope: scopeFn,
                retry: 2,
                retryDelay: 0,
                queryClient: qc,
            }),
        )!;
        m.mutate({ client_id: "a", name: "first" });
        m.mutate({ client_id: "a", name: "second" });
        await vi.waitFor(() => expect(order).toContain("second"));
        expect(attempts).toBe(3);
        expect(order).toEqual(["first", "first", "first", "second"]);
    });
});

describe("useServerMutation", () => {
    it("toasts without Retry, reconciles then invalidates", async () => {
        const calls: string[] = [];
        const srv = server();
        const m = scope.run(() =>
            useServerMutation<Vars, Row>({
                mutationKey: ["rows", "save"],
                mutationFn: srv.fn,
                reconcile: () => calls.push("reconcile"),
                invalidate: () => {
                    calls.push("invalidate");
                    return [LIST];
                },
                queryClient: qc,
            }),
        )!;
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.resolve({ client_id: "a", ID: 1, name: "A" });
        await flush();
        expect(calls).toEqual(["reconcile", "invalidate"]);

        m.mutate({ client_id: "b", name: "B" });
        await flush();
        srv.calls[1].d.reject(new Error("nope"));
        await flush();
        expect(push).toHaveBeenCalledWith("nope", "error");
        expect(notify).not.toHaveBeenCalled();
    });

    it("errorMessage: false leaves the error to the caller", async () => {
        const srv = server();
        const m = scope.run(() =>
            useServerMutation<Vars, Row>({
                mutationKey: ["rows", "save"],
                mutationFn: srv.fn,
                errorMessage: false,
                queryClient: qc,
            }),
        )!;
        m.mutate({ client_id: "a", name: "A" });
        await flush();
        srv.calls[0].d.reject(new Error("x"));
        await flush();
        expect(push).not.toHaveBeenCalled();
        expect(m.error.value).toBeInstanceOf(Error);
    });
});
