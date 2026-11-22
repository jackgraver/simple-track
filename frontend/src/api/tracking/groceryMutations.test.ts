import { QueryClient } from "@tanstack/vue-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, type EffectScope } from "vue";

import { createQueryClient } from "~/api/queryClient";
import { trackingKeys } from "./keys";
import type { GroceryItem } from "./types";

const api = vi.hoisted(() => ({ createGroceryItem: vi.fn(), completeGroceryItem: vi.fn() }));
vi.mock("~/api/tracking/api", () => api);
vi.mock("~/composables/toast/useToast", () => ({ toast: { push: vi.fn() } }));
const qcRef = vi.hoisted(() => ({ qc: null as unknown }));
vi.mock("@tanstack/vue-query", async (orig) => ({
    ...(await orig<typeof import("@tanstack/vue-query")>()),
    useQueryClient: () => qcRef.qc,
}));

import { useCompleteGroceryItem, useCreateGroceryItem } from "./queries";

const row = (ID: number, name: string, client_id?: string): GroceryItem => ({ ID, name, completed_at: null, client_id });
const flush = () => vi.advanceTimersByTimeAsync(0);

let qc: QueryClient;
let scope: EffectScope;
beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    qc = createQueryClient(vi.fn());
    qcRef.qc = qc;
    scope = effectScope();
    qc.setQueryData(trackingKeys.groceryItems, [row(1, "milk"), row(2, "eggs")]);
});
afterEach(() => {
    scope.stop();
    vi.useRealTimers();
});

describe("useCreateGroceryItem", () => {
    it("shows the row instantly and swaps in the server row", async () => {
        let resolve!: (r: GroceryItem) => void;
        api.createGroceryItem.mockReturnValue(new Promise((r) => (resolve = r)));
        const m = scope.run(() => useCreateGroceryItem())!;
        m.mutate({ client_id: "c1", name: "rice" });
        await flush();
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.slice(-1)[0]).toMatchObject({ name: "rice", client_id: "c1" });
        expect(api.createGroceryItem).toHaveBeenCalledWith("rice", "c1");
        resolve(row(7, "rice", "c1"));
        await flush();
        const list = qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!;
        expect(list.filter((r) => r.client_id === "c1")).toEqual([row(7, "rice", "c1")]);
    });

    it("rolls back on failure", async () => {
        api.createGroceryItem.mockRejectedValue(new Error("boom"));
        const m = scope.run(() => useCreateGroceryItem())!;
        m.mutate({ client_id: "c1", name: "rice" });
        await vi.advanceTimersByTimeAsync(5000);
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.map((r) => r.ID)).toEqual([1, 2]);
    });
});

describe("useCompleteGroceryItem", () => {
    it("removes the row now, invalidates list and suggestions on settle", async () => {
        api.completeGroceryItem.mockResolvedValue({ ...row(2, "eggs"), completed_at: "now" });
        const spy = vi.spyOn(qc, "invalidateQueries");
        const m = scope.run(() => useCompleteGroceryItem())!;
        m.mutate({ id: 2 });
        await flush();
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.map((r) => r.ID)).toEqual([1]);
        await flush();
        const keys = spy.mock.calls.map((c) => (c[0] as { queryKey: unknown }).queryKey);
        expect(keys).toContainEqual(trackingKeys.groceryItems);
        expect(keys).toContainEqual(trackingKeys.grocerySuggestionsAll);
    });

    it("restores the row when the request fails", async () => {
        api.completeGroceryItem.mockRejectedValue(new Error("boom"));
        const m = scope.run(() => useCompleteGroceryItem())!;
        m.mutate({ id: 1 });
        await vi.advanceTimersByTimeAsync(5000);
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.map((r) => r.ID)).toEqual([1, 2]);
    });

    it("a row ticked before its create returns is completed by client_id after the create", async () => {
        let resolveCreate!: (r: GroceryItem) => void;
        api.createGroceryItem.mockReturnValue(new Promise((r) => (resolveCreate = r)));
        api.completeGroceryItem.mockResolvedValue({ ...row(9, "rice", "c1"), completed_at: "now" });
        const create = scope.run(() => useCreateGroceryItem())!;
        const complete = scope.run(() => useCompleteGroceryItem())!;
        create.mutate({ client_id: "c1", name: "rice" });
        await flush();
        const temp = qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.find((r) => r.client_id === "c1")!;
        complete.mutate({ id: temp.ID, client_id: "c1" });
        await flush();
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.map((r) => r.ID)).toEqual([1, 2]);
        expect(api.completeGroceryItem).not.toHaveBeenCalled();
        resolveCreate(row(9, "rice", "c1"));
        await flush();
        await flush();
        expect(api.completeGroceryItem).toHaveBeenCalledWith("c1");
        // The create's reconcile did not bring the ticked row back.
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.map((r) => r.ID)).toEqual([1, 2]);
    });

    it("does not resurrect a row whose create failed", async () => {
        api.createGroceryItem.mockRejectedValue(new Error("boom"));
        api.completeGroceryItem.mockRejectedValue(new Error("not found"));
        const create = scope.run(() => useCreateGroceryItem())!;
        const complete = scope.run(() => useCompleteGroceryItem())!;
        create.mutate({ client_id: "c1", name: "rice" });
        await flush();
        const temp = qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.find((r) => r.client_id === "c1")!;
        complete.mutate({ id: temp.ID, client_id: "c1" });
        await vi.advanceTimersByTimeAsync(20_000);
        expect(qc.getQueryData<GroceryItem[]>(trackingKeys.groceryItems)!.map((r) => r.ID)).toEqual([1, 2]);
    });
});
