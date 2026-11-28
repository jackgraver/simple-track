import { readonly, ref } from "vue";

// Per-user browser storage. Everything that holds a user's data (drafts, rest timer, last route,
// hidden exercises) is keyed by the signed-in user's ID so a second person on the same device never
// sees the first person's. Device-wide UI preferences stay un-namespaced and survive sign-out:
//   - `st:theme` (localStorage): display preference, no user data.
//   - `st:chunk-reload-at` (sessionStorage): reload-loop guard timestamp.
//   - `tune`, `tune:values` (dev-only live-tuning panel): numbers, no user data.

const USER_PREFIX = "st:u:";
const OWNER_KEY = "st:owner";
export const LAST_ROUTE_COOKIE = "st_last_route";

/** Keys written before storage was per-user. Removed on sign-out and when ownership is unknown. */
const LEGACY_KEYS = ["st:last-route", "gym-rest-timer:global"];
const LEGACY_PREFIXES = [
    "simpletracker:workout:hidden-exercises:",
    "simpletracker:diet:logmeal-draft:",
    "gym-draft:",
];

const userId = ref<number | null>(null);

/** The signed-in user's ID, or null until /auth/me (or login) has told us. Reactive. */
export const currentUserId = readonly(userId);

/** `st:u:<id>:<name>`, or null while no user is known so callers read and write nothing. */
export function userKey(name: string): string | null {
    return userId.value === null ? null : `${USER_PREFIX}${userId.value}:${name}`;
}

function stores(): Storage[] {
    const out: Storage[] = [];
    try {
        out.push(window.localStorage);
    } catch {
        // storage unavailable
    }
    try {
        out.push(window.sessionStorage);
    } catch {
        // storage unavailable
    }
    return out;
}

function isUserOwned(key: string): boolean {
    return (
        key.startsWith(USER_PREFIX) ||
        key === OWNER_KEY ||
        LEGACY_KEYS.includes(key) ||
        LEGACY_PREFIXES.some((p) => key.startsWith(p))
    );
}

function expireLastRouteCookie(): void {
    try {
        document.cookie = `${LAST_ROUTE_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
    } catch {
        // cookies unavailable
    }
}

/**
 * Removes every user-scoped key (current and legacy) from localStorage and sessionStorage, the
 * last-route cookie, and forgets the current user. Call on logout and when the session expires.
 */
export function clearUserStorage(): void {
    for (const store of stores()) {
        try {
            const doomed: string[] = [];
            for (let i = 0; i < store.length; i++) {
                const key = store.key(i);
                if (key !== null && isUserOwned(key)) doomed.push(key);
            }
            for (const key of doomed) store.removeItem(key);
        } catch {
            // storage unavailable
        }
    }
    expireLastRouteCookie();
    userId.value = null;
}

/** Forget the user in memory only (reads return nothing); stored data is left for clearUserStorage. */
export function forgetUser(): void {
    userId.value = null;
}

/**
 * Record who is signed in. If the stored owner is a different user, or there is none (data from
 * before this existed has unknown ownership), the previous data is cleared first.
 */
export function setUserStorageOwner(id: number): void {
    let stored: string | null = null;
    try {
        stored = window.localStorage.getItem(OWNER_KEY);
    } catch {
        // storage unavailable: nothing persisted to leak either
    }
    if (stored !== String(id)) {
        clearUserStorage();
        try {
            window.localStorage.setItem(OWNER_KEY, String(id));
        } catch {
            // storage unavailable
        }
    }
    userId.value = id;
}
