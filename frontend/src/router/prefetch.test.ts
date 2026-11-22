import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { QueryClient } from "@tanstack/vue-query";
import { createMemoryHistory, createRouter, type RouteLocationNormalized } from "vue-router";

import { NAV_BUDGET_MS, prefetchRoute, settleWithin, warmNavigation, warmRoute } from "./prefetch";

const routeWith = (queries?: (r: RouteLocationNormalized) => never[] | object[]) =>
    ({ meta: { queries }, matched: [{ meta: { queries } }], query: {}, params: {} }) as unknown as RouteLocationNormalized;

describe("settleWithin", () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it("is true when the promise settles inside the budget", async () => {
        const p = settleWithin(new Promise((r) => setTimeout(r, 100)), 150);
        await vi.advanceTimersByTimeAsync(100);
        expect(await p).toBe(true);
    });

    it("is false at the budget when the promise is still pending", async () => {
        const p = settleWithin(new Promise((r) => setTimeout(r, 1000)), 150);
        await vi.advanceTimersByTimeAsync(150);
        expect(await p).toBe(false);
    });

    it("treats a rejection as settled (navigation must not fail)", async () => {
        expect(await settleWithin(Promise.reject(new Error("x")), 150)).toBe(true);
    });
});

describe("warmRoute", () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it("returns immediately with no queries", async () => {
        const qc = new QueryClient();
        expect(await warmRoute(qc, routeWith())).toBe(true);
    });

    it("resolves from cache without fetching, even when stale", async () => {
        const qc = new QueryClient();
        qc.setQueryData(["a"], 1);
        const queryFn = vi.fn(async () => 2);
        const ok = await warmRoute(qc, routeWith(() => [{ queryKey: ["a"], queryFn, staleTime: 0 }]));
        expect(ok).toBe(true);
        expect(queryFn).not.toHaveBeenCalled();
    });

    it("gives up after the budget but keeps the request running", async () => {
        const qc = new QueryClient();
        const queryFn = vi.fn(() => new Promise<number>((r) => setTimeout(() => r(7), 1000)));
        const p = warmRoute(qc, routeWith(() => [{ queryKey: ["slow"], queryFn }]));
        await vi.advanceTimersByTimeAsync(NAV_BUDGET_MS);
        expect(await p).toBe(false);
        await vi.advanceTimersByTimeAsync(1000);
        expect(qc.getQueryData(["slow"])).toBe(7);
        expect(queryFn).toHaveBeenCalledTimes(1);
    });

    it("does not throw or hang on a failing query or a throwing meta fn", async () => {
        const qc = new QueryClient();
        const failing = vi.fn(async () => {
            throw new Error("boom");
        });
        expect(await warmRoute(qc, routeWith(() => [{ queryKey: ["bad"], queryFn: failing, retry: false }]))).toBe(true);
        expect(
            await warmRoute(
                qc,
                routeWith(() => {
                    throw new Error("meta");
                }),
            ),
        ).toBe(true);
    });
});

describe("prefetchRoute", () => {
    it("prefetches every query and swallows errors", async () => {
        const qc = new QueryClient();
        const a = vi.fn(async () => 1);
        const b = vi.fn(async () => {
            throw new Error("x");
        });
        prefetchRoute(
            qc,
            routeWith(() => [
                { queryKey: ["a"], queryFn: a },
                { queryKey: ["b"], queryFn: b, retry: false },
            ]),
        );
        await vi.waitFor(() => expect(qc.getQueryData(["a"])).toBe(1));
        expect(b).toHaveBeenCalled();
    });

    it("skips fresh data", async () => {
        const qc = new QueryClient();
        qc.setQueryData(["a"], 1);
        const a = vi.fn(async () => 2);
        prefetchRoute(qc, routeWith(() => [{ queryKey: ["a"], queryFn: a, staleTime: 60_000 }]));
        await Promise.resolve();
        expect(a).not.toHaveBeenCalled();
    });
});

describe("leaf-only route queries (nfix #5)", () => {
    const Empty = { render: () => null };
    const dayFn = vi.fn(async () => "day");
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [
            {
                path: "/gym",
                name: "gym",
                component: Empty,
                meta: { queries: (to) => [{ queryKey: ["day", to.query.offset ?? "0"], queryFn: dayFn }] },
                children: [{ path: "plans", name: "gym-plans", component: Empty }],
            },
        ],
    });

    it("a child route does not warm the parent's queries (vue-router merges meta)", async () => {
        const qc = new QueryClient();
        const plans = router.resolve("/gym/plans");
        expect(plans.meta.queries).toBeDefined(); // inherited
        prefetchRoute(qc, plans);
        expect(await warmRoute(qc, plans)).toBe(true);
        expect(dayFn).not.toHaveBeenCalled();
        prefetchRoute(qc, router.resolve("/gym"));
        await vi.waitFor(() => expect(dayFn).toHaveBeenCalledTimes(1));
    });

    it("a query-only change on the same route prefetches without waiting", async () => {
        vi.useFakeTimers();
        const qc = new QueryClient();
        const slow = vi.fn(() => new Promise((r) => setTimeout(r, 1000)));
        const r = (o: string) => ({
            ...routeWith(() => [{ queryKey: ["d", o], queryFn: slow }]),
            name: "gym",
            fullPath: `/gym?offset=${o}`,
        });
        let done = false;
        void warmNavigation(qc, r("2"), r("1")).then(() => (done = true));
        await vi.advanceTimersByTimeAsync(0);
        expect(done).toBe(true);
        expect(slow).toHaveBeenCalledTimes(1);
        vi.useRealTimers();
    });
});
