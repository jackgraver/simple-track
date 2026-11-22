import type { QueryClient } from "@tanstack/vue-query";
import type { RouteLocationNormalized } from "vue-router";

/** Max time a navigation waits for a route's primary queries (DESIGN.md "Prefetch"). */
export const NAV_BUDGET_MS = 150;

/** A `queryOptions(...)` result (vue-query's union type is not worth modelling here). */
export type PrefetchQuery = object;

/** What prefetch reads from a location: both normalized and resolved locations fit. */
export type RouteLike = Pick<RouteLocationNormalized, "meta" | "query" | "params" | "matched">;

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

/**
 * The leaf record's own `queries`. Not `route.meta`: vue-router merges parent meta into children,
 * so `/gym/plans` would otherwise warm (and create) the gym day it never reads.
 */
function leafQueries(route: RouteLike): RouteQueries | undefined {
    return route.matched[route.matched.length - 1]?.meta.queries;
}

function safeQueries(route: RouteLike): PrefetchQuery[] {
    const fn = leafQueries(route);
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

type NavLike = RouteLike & Pick<RouteLocationNormalized, "name" | "fullPath">;

/**
 * `beforeResolve` warmup. A query-only change on the same page (day nav, week strip) is not held:
 * the page shows its own loading state for a cold day, so it only prefetches. Other navigations wait
 * up to the budget (`warmRoute`).
 */
export async function warmNavigation(qc: Prefetcher, to: NavLike, from: NavLike): Promise<void> {
    if (to.name === "signin" || to.fullPath === from.fullPath) return;
    if (to.name === from.name) {
        prefetchRoute(qc, to);
        return;
    }
    await warmRoute(qc, to);
}
