import type { QueryClient } from "@tanstack/vue-query";
import type { RouteLocationNormalized } from "vue-router";

/** Max time a navigation waits for a route's primary queries (DESIGN.md "Prefetch"). */
export const NAV_BUDGET_MS = 150;

/** A `queryOptions(...)` result (vue-query's union type is not worth modelling here). */
export type PrefetchQuery = object;

/** What prefetch reads from a location: both normalized and resolved locations fit. */
export type RouteLike = Pick<RouteLocationNormalized, "meta" | "query" | "params">;

/** `route.meta.queries`: the route's primary queries for the target location. */
export type RouteQueries = (route: RouteLike) => PrefetchQuery[];

type Prefetcher = Pick<QueryClient, "ensureQueryData" | "prefetchQuery">;

/** Resolves `true` when `p` settled within `ms`, `false` on timeout. Never rejects. */
export function settleWithin(p: Promise<unknown>, ms: number): Promise<boolean> {
    return new Promise((resolve) => {
        const timer = setTimeout(() => resolve(false), ms);
        p.then(
            () => {
                clearTimeout(timer);
                resolve(true);
            },
            () => {
                clearTimeout(timer);
                resolve(true);
            },
        );
    });
}

function safeQueries(route: RouteLike): PrefetchQuery[] {
    const fn = route.meta.queries;
    if (!fn) return [];
    try {
        return fn(route);
    } catch {
        return [];
    }
}

/**
 * Warm the cache for a navigation: `ensureQueryData` every query (cached data is returned at once,
 * even if stale; missing data is fetched) but wait at most `budgetMs`. Requests keep running after
 * the budget so the page finds them in flight. Never throws and never blocks on failure.
 * Resolves to `true` if everything was ready within the budget.
 */
export async function warmRoute(
    qc: Prefetcher,
    route: RouteLike,
    budgetMs: number = NAV_BUDGET_MS,
): Promise<boolean> {
    const queries = safeQueries(route);
    if (queries.length === 0) return true;
    const all = Promise.all(
        queries.map((q) => qc.ensureQueryData(q as Parameters<Prefetcher["ensureQueryData"]>[0])),
    );
    // A rejection must not become an unhandled one after the budget elapsed.
    all.catch(() => undefined);
    return settleWithin(all, budgetMs);
}

/** Intent prefetch (hover/touch/focus): fire and forget, honours `staleTime`. */
export function prefetchRoute(qc: Prefetcher, route: RouteLike): void {
    for (const q of safeQueries(route)) {
        void qc.prefetchQuery(q as Parameters<Prefetcher["prefetchQuery"]>[0]);
    }
}
