import { onScopeDispose, ref } from "vue";

/** True from the `lg` breakpoint (64rem), the one layout switch (PRINCIPLES §5). Reactive. */
export function useIsDesktop() {
    const query = typeof window !== "undefined" && window.matchMedia ? window.matchMedia("(min-width: 64rem)") : null;
    const isDesktop = ref(query?.matches ?? true);
    if (query) {
        const onChange = (e: MediaQueryListEvent) => (isDesktop.value = e.matches);
        query.addEventListener("change", onChange);
        onScopeDispose(() => query.removeEventListener("change", onChange));
    }
    return isDesktop;
}
