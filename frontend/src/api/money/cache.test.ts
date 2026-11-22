import { describe, expect, it, vi } from "vitest";

import { useDeleteBalance, useDeleteTransaction } from "./queries";
import { compareBalances, compareTransactions, removeById, restoreById } from "./cache";
import type { AccountBalance, AccountTransaction } from "./types";

const tx = (ID: number, date: string): AccountTransaction => ({
    ID,
    account_id: 1,
    kind: "deposit",
    amount: 10,
    date,
    note: "",
});
const bal = (ID: number, date: string): AccountBalance => ({ ID, account_id: 1, date, balance: 100 });

describe("money cache patches", () => {
    it("removeById drops one row and keeps identity of the rest", () => {
        const a = tx(1, "2026-01-02");
        const b = tx(2, "2026-01-01");
        const out = removeById([a, b], 1);
        expect(out).toEqual([b]);
        expect(out[0]).toBe(b);
    });

    it("removeById returns the same array when the id is absent", () => {
        const list = [tx(1, "2026-01-02")];
        expect(removeById(list, 9)).toBe(list);
    });

    it("restoreById puts the row back in server order, keeping a sibling's removal", () => {
        const a = tx(3, "2026-03-01");
        const b = tx(2, "2026-02-01");
        const c = tx(1, "2026-01-01");
        const snapshot = [a, b, c];
        // we deleted b; a sibling deleted c
        const current = [a];
        expect(restoreById(current, snapshot, 2, compareTransactions)).toEqual([a, b]);
    });

    it("restoreById is a no-op when present or unknown", () => {
        const a = bal(1, "2026-01-01");
        const list = [a];
        expect(restoreById(list, [a], 1, compareBalances)).toBe(list);
        expect(restoreById(list, [a], 7, compareBalances)).toBe(list);
    });

    it("orders newest date first, then newest id", () => {
        const rows = [bal(1, "2026-01-01"), bal(3, "2026-02-01"), bal(2, "2026-02-01")];
        expect([...rows].sort(compareBalances).map((r) => r.ID)).toEqual([3, 2, 1]);
    });
});

// The hooks use useMutation, which needs a component; exercise the optimistic config directly.
vi.mock("~/api/optimistic", () => ({
    useOptimisticMutation: (opts: unknown) => opts,
    useServerMutation: (opts: unknown) => opts,
}));
vi.mock("./api", () => ({ deleteTransaction: vi.fn(), deleteBalance: vi.fn() }));

describe("delete hooks (partial optimistic)", () => {
    it("transaction: patches only the transactions list; totals refetch on settle", () => {
        const opts = useDeleteTransaction() as any;
        const [patch, ...rest] = opts.optimistic({ accountId: 5, id: 2 });
        expect(rest).toEqual([]);
        expect(patch.queryKey).toEqual(["money", "account", 5, "transactions"]);
        expect(patch.apply([tx(1, "2026-01-01"), tx(2, "2026-01-02")], { accountId: 5, id: 2 }).map((t: AccountTransaction) => t.ID)).toEqual([1]);
        expect(opts.invalidate()).toEqual([["money"]]);
    });

    it("balance: patches only the balances list; revert restores the row", () => {
        const opts = useDeleteBalance() as any;
        const [patch] = opts.optimistic({ accountId: 5, id: 2 });
        expect(patch.queryKey).toEqual(["money", "account", 5, "balances"]);
        const snap = [bal(2, "2026-02-01"), bal(1, "2026-01-01")];
        expect(patch.revert([bal(1, "2026-01-01")], { accountId: 5, id: 2 }, snap).map((b: AccountBalance) => b.ID)).toEqual([2, 1]);
    });
});

