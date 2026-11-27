import { computed, ref } from "vue";
import { forgetUser, setUserStorageOwner } from "~/utils/userStorage";

export type AuthStatus = "unknown" | "authenticated" | "unauthenticated";

export const authStatus = ref<AuthStatus>("unknown");
export const username = ref<string | null>(null);
export const environment = ref<string | null>(null);

export const isAuthenticated = computed(() => authStatus.value === "authenticated");
export const isDevEnvironment = computed(
    () => environment.value === "dev" || environment.value === "development",
);

let unauthorizedRedirect: (() => void) | null = null;

export function setUnauthorizedRedirect(fn: () => void) {
    unauthorizedRedirect = fn;
}

export function markAuthenticated(name: string, id?: number) {
    // Before the auth status flips, so nothing user-scoped reads storage under the wrong user.
    if (id !== undefined) setUserStorageOwner(id);
    else forgetUser();
    username.value = name;
    authStatus.value = "authenticated";
}

export function markUnauthorized() {
    forgetUser();
    username.value = null;
    authStatus.value = "unauthenticated";
}

export function setEnvironment(value: string | undefined) {
    environment.value = value ?? null;
}

export function notifyUnauthorized() {
    unauthorizedRedirect?.();
}
