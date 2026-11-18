// Remembers the last visited route so a PWA relaunched by the OS (which loads
// the manifest start_url "/") can resume where the user left off.
const KEY = "st:last-route";
export const LAST_ROUTE_MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

export function saveLastRoute(fullPath: string, now = Date.now()): void {
    try {
        localStorage.setItem(KEY, JSON.stringify({ path: fullPath, at: now }));
    } catch {
        // storage unavailable: resuming is best-effort
    }
}

export function loadLastRoute(now = Date.now()): string | null {
    try {
        const raw = localStorage.getItem(KEY);
        if (!raw) return null;
        const { path, at } = JSON.parse(raw) as { path?: unknown; at?: unknown };
        if (typeof path !== "string" || typeof at !== "number") return null;
        if (!path.startsWith("/") || path.startsWith("//")) return null;
        if (now - at > LAST_ROUTE_MAX_AGE_MS) return null;
        return path;
    } catch {
        return null;
    }
}
