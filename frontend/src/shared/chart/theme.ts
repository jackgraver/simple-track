import { color as parseColor } from "chart.js/helpers";
import { ref } from "vue";

// The one place chart code reads design tokens (docs/design/TOKENS.md, `chart-*`).
// Colours are resolved to rgb() through a probe element, so hsl()/hex/space-syntax tokens
// all come back in a form canvas and chart.js parse.

/** Bumped when <html data-theme> changes. Reading it inside a computed makes that computed re-run. */
const themeVersion = ref(0);
let observer: MutationObserver | null = null;

function ensureObserver() {
    if (observer || typeof document === "undefined") return;
    observer = new MutationObserver(() => themeVersion.value++);
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
}

let probe: HTMLElement | null = null;
function resolve(token: string): string {
    if (typeof document === "undefined") return "";
    if (!probe) {
        probe = document.createElement("span");
        probe.setAttribute("aria-hidden", "true");
        probe.style.display = "none";
        document.documentElement.appendChild(probe);
    }
    probe.style.color = `var(${token})`;
    return getComputedStyle(probe).color;
}

export type ChartTheme = {
    /** Axis tick and legend text. */
    muted: string;
    grid: string;
    /** Series colours in order (blue, green, orange, violet, pink). */
    series: [string, string, string, string, string];
    success: string;
    danger: string;
};

/** Reactive: call it inside a computed and the result follows the theme. */
export function chartTheme(): ChartTheme {
    ensureObserver();
    void themeVersion.value;
    return {
        muted: resolve("--color-chart-text"),
        grid: resolve("--color-chart-grid"),
        series: [1, 2, 3, 4, 5].map((n) => resolve(`--color-chart-${n}`)) as ChartTheme["series"],
        success: resolve("--color-success"),
        danger: resolve("--color-danger"),
    };
}

/** `rgb(...)` with an alpha, for area fills. */
export function withAlpha(colour: string, alpha: number): string {
    return parseColor(colour).alpha(alpha).rgbString();
}
