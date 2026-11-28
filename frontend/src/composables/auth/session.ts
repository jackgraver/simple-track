import { computed, ref } from "vue";
import { forgetUser, setUserStorageOwner } from "~/utils/userStorage";

export type AuthStatus = "unknown" | "authenticated" | "unauthenticated";

export const authStatus = ref<AuthStatus>("unknown");
export const username = ref<string | null>(null);
/** The signed-in user is the admin (from the users row via /auth/me; the server re-checks every request). */
export const isAdmin = ref(false);
export const environment = ref<string | null>(null);

export const isAuthenticated = computed(() => authStatus.value === "authenticated");
export const isDevEnvironment = computed(
    () => environment.value === "dev" || environment.value === "development",
);

let unauthorizedRedirect: (() => void) | null = null;

export function setUnauthorizedRedirect(fn: () => void) {
    unauthorizedRedirect = fn;
}

/**
 * `id` scopes per-user storage (a different user than the stored owner clears it first, so a
 * setup link or another sign-in on a shared device never sees the previous user's data).
 */
export function markAuthenticated(name: string, id?: number, admin = false) {
    // Before the auth status flips, so nothing user-scoped reads storage under the wrong user.
    if (id !== undefined) setUserStorageOwner(id);
    else forgetUser();
    username.value = name;
    isAdmin.value = admin;
    authStatus.value = "authenticated";
}

export function markUnauthorized() {
    forgetUser();
    username.value = null;
    isAdmin.value = false;
    authStatus.value = "unauthenticated";
}

export function setEnvironment(value: string | undefined) {
    environment.value = value ?? null;
}

export function notifyUnauthorized() {
    unauthorizedRedirect?.();
}
