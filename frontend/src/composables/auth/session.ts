import { computed, ref } from "vue";

export type AuthStatus = "unknown" | "authenticated" | "unauthenticated";

export const authStatus = ref<AuthStatus>("unknown");
export const username = ref<string | null>(null);

export const isAuthenticated = computed(() => authStatus.value === "authenticated");

let unauthorizedRedirect: (() => void) | null = null;

export function setUnauthorizedRedirect(fn: () => void) {
    unauthorizedRedirect = fn;
}

export function markAuthenticated(name: string) {
    username.value = name;
    authStatus.value = "authenticated";
}

export function markUnauthorized() {
    username.value = null;
    authStatus.value = "unauthenticated";
}

export function notifyUnauthorized() {
    unauthorizedRedirect?.();
}
