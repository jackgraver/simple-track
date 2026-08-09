import { computed, ref } from "vue";

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

export function markAuthenticated(name: string) {
    username.value = name;
    authStatus.value = "authenticated";
}

export function markUnauthorized() {
    username.value = null;
    authStatus.value = "unauthenticated";
}

export function setEnvironment(value: string | undefined) {
    environment.value = value ?? null;
}

export function notifyUnauthorized() {
    unauthorizedRedirect?.();
}
