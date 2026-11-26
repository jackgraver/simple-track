import { reactive } from "vue";

// Live tuning for motion (and any other numbers worth feeling out), dev server only.
//
// Add `?tune` to any URL under `npm run dev` and a Tweakpane panel opens with a slider per knob.
// Dragging a slider changes the value the running page reads, so the next animation already uses
// it: no reload, no edit. "Copy values" puts the numbers on the clipboard to hand over; they then
// become the new defaults in code. `?tune=0` hides the panel again. The choice and the tweaked
// values survive reloads (session/local storage), so HMR and navigation don't lose them.
//
// In a production build `tunable` returns the defaults as a plain object and Tweakpane is never
// bundled: the branch that loads it is behind `import.meta.env.DEV`, which Vite compiles to false.

export type Knob = { value: number; min: number; max: number; step: number };
export type Knobs = Record<string, Knob>;
export type TuneGroup = { name: string; knobs: Knobs; values: Record<string, number> };

const STORAGE_KEY = "tune:values";
// Kept on globalThis so hot reloads of this file (or of the files declaring knobs) reuse the same
// groups instead of registering them again.
const store = globalThis as { __tuneGroups?: Map<string, TuneGroup> };
const groups = (store.__tuneGroups ??= new Map<string, TuneGroup>());

/** Every registered group, in declaration order (the panel shows one folder each). */
export const tuneGroups = (): TuneGroup[] => [...groups.values()];

function enabled(): boolean {
    try {
        const q = new URLSearchParams(location.search).get("tune");
        if (q !== null) sessionStorage.setItem("tune", q === "0" ? "" : "1");
        return sessionStorage.getItem("tune") === "1";
    } catch {
        return false;
    }
}

function saved(name: string): Record<string, number> {
    try {
        return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? "{}")[name] ?? {};
    } catch {
        return {};
    }
}

/** Writes every group's current values; the panel calls it on each change and reset. */
export function saveTuning(): void {
    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(Object.fromEntries([...groups].map(([n, g]) => [n, g.values]))));
    } catch {
        // Storage blocked: tuning still works for this page load.
    }
}

/**
 * A named group of tunable numbers. Read the returned object's fields at the moment they are used
 * (inside a function, a computed or the template), not once at module load, so a tweak takes effect.
 */
export function tunable<K extends Knobs>(name: string, knobs: K): Readonly<{ [P in keyof K]: number }> {
    const defaults = Object.fromEntries(Object.entries(knobs).map(([k, v]) => [k, v.value])) as { [P in keyof K]: number };
    if (!import.meta.env.DEV || !enabled()) return defaults;

    // A hot-reloaded module calls this again: keep the same values and panel folder.
    const existing = groups.get(name);
    if (existing) return existing.values as { [P in keyof K]: number };

    const stored = saved(name);
    const values = reactive<Record<string, number>>({ ...defaults });
    for (const k of Object.keys(defaults)) if (typeof stored[k] === "number") values[k] = stored[k]!;
    const group: TuneGroup = { name, knobs, values };
    groups.set(name, group);
    void import("./tunePanel").then((panel) => panel.render());
    return values as { [P in keyof K]: number };
}
