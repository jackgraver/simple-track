// Review finding 3: a day's request derives its offset from the date when it is sent, so a page left
// open past midnight neither refetches the new "today" under yesterday's key nor writes to the wrong day.
import { VueQueryPlugin, type QueryClient } from "@tanstack/vue-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createApp } from "vue";

import { createQueryClient } from "~/api/queryClient";

const api = vi.hoisted(() => ({
    getDietLogsToday: vi.fn(),
    addPlannedMealFromSaved: vi.fn(),
    deletePlannedMeal: vi.fn(),
    reorderPlannedMeals: vi.fn(),
    quickLog: vi.fn(),
}));
vi.mock("./api", () => api);
vi.mock("vue-router", () => ({ useRouter: () => ({}) }));
vi.mock("~/composables/toast/useToast", () => ({ toast: { push: vi.fn() } }));

import {
    dayRef,
    dietQueries,
    quickLogVars,
    useAddPlannedFromSavedAt,
    useDeletePlannedMealAt,
    useQuickLog,
    useReorderPlannedMealsAt,
} from "./queries";
import { dietKeys } from "./keys";
import type { SavedMeal } from "~/types/diet";

const BEFORE_MIDNIGHT = new Date(2026, 10, 22, 23, 59, 0);
const AFTER_MIDNIGHT = new Date(2026, 10, 23, 0, 1, 0);

let qc: QueryClient;
function withQueryClient<T>(fn: () => T): T {
    const app = createApp({});
    app.use(VueQueryPlugin, { queryClient: qc });
    return app.runWithContext(fn);
}

beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(BEFORE_MIDNIGHT);
    // Mutations answer nothing, so reconcile leaves the (empty) cache alone between calls.
    for (const fn of Object.values(api)) fn.mockReset().mockResolvedValue(undefined);
    api.getDietLogsToday.mockResolvedValue({});
    qc = createQueryClient(() => {});
});
afterEach(() => vi.useRealTimers());

describe("diet day across midnight", () => {
    it("a refetch after midnight asks for the day in its key, not the new today", async () => {
        const opts = dietQueries.day(0);
        expect(opts.queryKey).toEqual(dietKeys.days.day("2026-11-22"));
        vi.setSystemTime(AFTER_MIDNIGHT);
        await qc.fetchQuery(opts);
        expect(api.getDietLogsToday).toHaveBeenCalledWith(1);
    });

    it("day mutations send the offset of the day they were made for", async () => {
        const day = dayRef(0); // 2026-11-22, built before midnight
        const addPlanned = withQueryClient(useAddPlannedFromSavedAt);
        const deletePlanned = withQueryClient(useDeletePlannedMealAt);
        const reorder = withQueryClient(useReorderPlannedMealsAt);
        const quick = withQueryClient(useQuickLog);
        const vars = quickLogVars({ name: "Banana", calories: 100, protein: 1, carbs: 25, fat: 0, fiber: 3 }, 0);

        vi.setSystemTime(AFTER_MIDNIGHT);
        await addPlanned.mutateAsync({ ...day, savedMeal: { ID: 5, name: "Oats" } as SavedMeal, client_id: "c1" });
        await deletePlanned.mutateAsync({ ...day, plannedMealId: 9, client_id: "c2" });
        await reorder.mutateAsync({ ...day, ids: [1, 2], client_id: "c3" });
        await quick.mutateAsync(vars);

        expect(api.addPlannedMealFromSaved).toHaveBeenCalledWith(5, 1, "c1");
        expect(api.deletePlannedMeal).toHaveBeenCalledWith(9, 1);
        expect(api.reorderPlannedMeals).toHaveBeenCalledWith([1, 2], 1);
        expect(api.quickLog).toHaveBeenCalledWith(expect.objectContaining({ date: "2026-11-22", offset: 1 }));
    });
});
