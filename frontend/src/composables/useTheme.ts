import { computed, ref } from "vue";

// Mirrors the bootstrap script in index.html (which sets data-theme before first paint).
export type ThemeMode = "dark" | "light" | "system";
export type ResolvedTheme = "dark" | "light";

const STORAGE_KEY = "st:theme";

function readMode(): ThemeMode {
    try {
        const v = localStorage.getItem(STORAGE_KEY);
        if (v === "light" || v === "system") return v;
    } catch {
        // storage unavailable: fall through to the default
    }
    return "dark";
}

const mode = ref<ThemeMode>(readMode());
const systemLight = ref(false);
let query: MediaQueryList | null = null;

const resolved = computed<ResolvedTheme>(() =>
    mode.value === "system" ? (systemLight.value ? "light" : "dark") : mode.value,
);

function apply() {
    const root = document.documentElement;
    root.setAttribute("data-theme", resolved.value);
    // The browser chrome follows the page canvas (read back from the token, no colour literals).
    document
        .querySelector('meta[name="theme-color"]')
        ?.setAttribute("content", getComputedStyle(root).backgroundColor);
}

/** Call once at startup: follows the OS setting while mode is "system". */
export function initTheme() {
    query = window.matchMedia("(prefers-color-scheme: light)");
    systemLight.value = query.matches;
    query.addEventListener("change", (e) => {
        systemLight.value = e.matches;
        apply();
    });
    apply();
}

export function setThemeMode(next: ThemeMode) {
    mode.value = next;
    try {
        localStorage.setItem(STORAGE_KEY, next);
    } catch {
        // not persisted, still applies for this session
    }
    apply();
}

export function useTheme() {
    return { mode, resolved, setThemeMode };
}
