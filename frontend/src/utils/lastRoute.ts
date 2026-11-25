// Remembers the last visited route so a PWA relaunched by the OS can resume
// where the user left off. Stored in localStorage and in a cookie: iOS can drop
// script-writable storage for installed web apps, while cookies (which keep the
// session signed in) survive a swipe-kill.
const KEY = "st:last-route";
const COOKIE = "st_last_route";
export const LAST_ROUTE_MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

function encode(path: string, now: number): string {
    return JSON.stringify({ path, at: now });
}

export interface LastRoute {
    path: string;
    at: number;
}

function parse(raw: string | null | undefined, now: number): LastRoute | null {
    if (!raw) return null;
    try {
        const { path, at } = JSON.parse(raw) as { path?: unknown; at?: unknown };
        if (typeof path !== "string" || typeof at !== "number") return null;
        if (!path.startsWith("/") || path.startsWith("//")) return null;
        if (now - at > LAST_ROUTE_MAX_AGE_MS) return null;
        return { path, at };
    } catch {
        return null;
    }
}

export function saveLastRoute(fullPath: string, now = Date.now()): void {
    const value = encode(fullPath, now);
    try {
        localStorage.setItem(KEY, value);
    } catch {
        // storage unavailable: the cookie below is the fallback
    }
    try {
        const secure = location.protocol === "https:" ? "; Secure" : "";
        document.cookie = `${COOKIE}=${encodeURIComponent(value)}; Path=/; Max-Age=${LAST_ROUTE_MAX_AGE_MS / 1000}; SameSite=Lax${secure}`;
    } catch {
        // cookies unavailable: resuming is best-effort
    }
}

function readCookie(): string | null {
    try {
        for (const part of document.cookie.split("; ")) {
            if (part.startsWith(`${COOKIE}=`)) {
                return decodeURIComponent(part.slice(COOKIE.length + 1));
            }
        }
    } catch {
        // ignore
    }
    return null;
}

export function loadLastRoute(now = Date.now()): LastRoute | null {
    let local: string | null = null;
    try {
        local = localStorage.getItem(KEY);
    } catch {
        // ignore
    }
    return parse(local, now) ?? parse(readCookie(), now);
}

export const LOGGING_RESUME_WINDOW_MS = 3 * 60 * 60 * 1000;

export function isExpiredLoggingResume(entry: LastRoute, now = Date.now()): boolean {
    return entry.path.startsWith("/gym/logging") && now - entry.at > LOGGING_RESUME_WINDOW_MS;
}
