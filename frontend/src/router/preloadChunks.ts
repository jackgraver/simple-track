import type { Router } from "vue-router";

type Loader = () => Promise<unknown>;

/** Lazy route components are loader functions; an already-loaded one is the component object. */
function loaderOf(component: unknown): Loader | null {
    return typeof component === "function" && !("__vccOpts" in component) && !("setup" in component)
        ? (component as Loader)
        : null;
}

/** `requestIdleCallback` where available (not Safari), else a short timeout. */
function whenIdle(fn: () => void): void {
    if (typeof globalThis.requestIdleCallback === "function") globalThis.requestIdleCallback(fn, { timeout: 3000 });
    else globalThis.setTimeout(fn, 1500);
}

/**
 * Download the code of the named routes once the browser is idle, so navigating there later
 * doesn't wait on chunk requests. Only fetches JS (no data, no render), and skips Save-Data
 * connections. Loaders resolve to the same module later, so this never loads anything twice.
 */
export function preloadRouteChunks(router: Router, names: string[]): void {
    type WithConnection = Navigator & { connection?: { saveData?: boolean } };
    const conn = (globalThis.navigator as WithConnection | undefined)?.connection;
    if (conn?.saveData) return;
    whenIdle(() => {
        for (const name of names) {
            const record = router.getRoutes().find((r) => r.name === name);
            const load = loaderOf(record?.components?.default);
            // A failed preload is harmless: the real navigation retries and reports errors.
            if (load) load().catch(() => undefined);
        }
    });
}
