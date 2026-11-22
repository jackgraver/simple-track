import { describe, expect, it } from "vitest";

import type { GroceryItem } from "~/api/tracking/types";
import {
    addGroceryItem,
    completeGroceryItemInList,
    isTempGroceryId,
    reconcileGroceryItem,
    removeGroceryItemByClientId,
    restoreGroceryItem,
    tempGroceryId,
    groceryPathRef,
} from "./groceryCache";

const item = (ID: number, name: string, client_id?: string): GroceryItem => ({ ID, name, completed_at: null, client_id });
const base = [item(1, "milk"), item(2, "eggs"), item(3, "bread")];

describe("grocery cache patches", () => {
    it("temp ids are stable, negative and distinct per client_id", () => {
        expect(tempGroceryId("a")).toBe(tempGroceryId("a"));
        expect(isTempGroceryId(tempGroceryId("a"))).toBe(true);
        expect(tempGroceryId("a")).not.toBe(tempGroceryId("b"));
    });

    it("add appends once (idempotent for Retry) and revert removes it", () => {
        const v = { client_id: "c1", name: "rice" };
        const once = addGroceryItem(base, v);
        expect(once.slice(-1)[0]).toMatchObject({ name: "rice", client_id: "c1", completed_at: null });
        expect(addGroceryItem(once, v)).toHaveLength(4);
        expect(removeGroceryItemByClientId(once, v)).toEqual(base);
    });

    it("reconcile swaps the temp row for the server row, keeping position", () => {
        const withTemp = addGroceryItem(base, { client_id: "c1", name: "rice" });
        const server = item(9, "rice", "c1");
        const next = reconcileGroceryItem(withTemp, server, "c1");
        expect(next).toHaveLength(4);
        expect(next.slice(-1)[0]).toEqual(server);
        // Row already gone (ticked off while the create was in flight) -> stays gone; already real -> no duplicate.
        expect(reconcileGroceryItem(base, server, "c1")).toEqual(base);
        expect(reconcileGroceryItem(next, server, "c1")).toHaveLength(4);
    });

    it("complete removes the row; restore puts it back in place", () => {
        const after = completeGroceryItemInList(base, { id: 2 });
        expect(after.map((r) => r.ID)).toEqual([1, 3]);
        expect(restoreGroceryItem(after, { id: 2 }, base)).toEqual(base);
    });

    it("restore keeps other changes, and does not duplicate or invent rows", () => {
        const after = [...completeGroceryItemInList(base, { id: 2 }), item(-5, "tea", "c9")];
        expect(restoreGroceryItem(after, { id: 2 }, base).map((r) => r.ID)).toEqual([1, 2, 3, -5]);
        expect(restoreGroceryItem(base, { id: 2 }, base)).toEqual(base);
        expect(restoreGroceryItem(after, { id: 99 }, base)).toEqual(after);
    });

    it("a row still on its temp id is completed and restored by client_id", () => {
        const withTemp = addGroceryItem(base, { client_id: "c1", name: "rice" });
        const temp = withTemp[3];
        const v = { id: temp.ID, client_id: "c1" };
        const after = completeGroceryItemInList(withTemp, v);
        expect(after).toEqual(base);
        // The create reconciled meanwhile: the row is not re-added by the swap...
        expect(reconcileGroceryItem(after, item(9, "rice", "c1"), "c1")).toEqual(base);
        // ...and a failed complete restores it (matched by client_id).
        expect(restoreGroceryItem(after, v, withTemp).slice(-1)[0]).toEqual(temp);
        expect(groceryPathRef(v)).toBe("c1");
        expect(groceryPathRef({ id: 4 })).toBe("4");
    });
});
