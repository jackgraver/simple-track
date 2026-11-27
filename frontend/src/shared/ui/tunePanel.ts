// The Tweakpane panel behind `?tune` (see tune.ts). Only ever loaded by a dynamic import in dev.
import { Pane, type FolderApi } from "tweakpane";
import { saveTuning, tuneGroups, type TuneGroup } from "./tune";

// One panel per page, whatever hot reloads do: the live one is kept on globalThis and replaced on
// each render, which rebuilds it from every registered group (one folder each).
const store = globalThis as { __tunePane?: Pane; __tuneCollapsed?: Set<string> };
const collapsed = (store.__tuneCollapsed ??= new Set<string>());
const PANE = "\u0000pane";

// Slider steps add floating-point noise (0.09999999999999999): round to the knob's step.
function round(g: TuneGroup, key: string, v: number): number {
    const decimals = (String(g.knobs[key]?.step ?? 1).split(".")[1] ?? "").length;
    return Number(v.toFixed(decimals));
}

function format(groups: TuneGroup[]): string {
    return groups
        .map((g) => `${g.name}: ${Object.entries(g.values).map(([k, v]) => `${k} ${round(g, k, v)}`).join(", ")}`)
        .join("\n");
}

/** A "Copy values" button: copies `groups()` (logged too, in case the clipboard is blocked). */
function addCopyButton(container: Pane | FolderApi, label: string, groups: () => TuneGroup[]) {
    const button = container.addButton({ title: label });
    button.on("click", () => {
        const text = format(groups());
        console.info(`[tune]\n${text}`);
        void navigator.clipboard.writeText(text).then(
            () => (button.title = "Copied"),
            () => (button.title = "Copy failed (values in console)"),
        );
        setTimeout(() => (button.title = label), 1500);
    });
}

// Hover a row to see what its knob changes. One tooltip element, beside the panel (left of the
// row when there's room, else under it).
function tooltip(): HTMLElement {
    let el = document.getElementById("tune-hint");
    if (!el) {
        el = document.createElement("div");
        el.id = "tune-hint";
        el.role = "tooltip";
        el.className =
            "pointer-events-none fixed hidden max-w-64 rounded-control border border-line bg-surface-3 px-2.5 py-1.5 text-xs/snug text-fg shadow-overlay";
        el.style.zIndex = "1001"; // above Tweakpane's own panel (dev only, not a z token)
        document.body.append(el);
    }
    return el;
}

function attachHint(row: HTMLElement, hint: string) {
    row.addEventListener("pointerenter", () => {
        const tip = tooltip();
        tip.textContent = hint;
        tip.classList.remove("hidden");
        const r = row.getBoundingClientRect();
        const room = r.left > 272;
        tip.style.left = room ? "" : `${r.left}px`;
        tip.style.right = room ? `${innerWidth - r.left + 8}px` : "";
        tip.style.top = room ? `${r.top}px` : `${r.bottom + 4}px`;
    });
    row.addEventListener("pointerleave", () => tooltip().classList.add("hidden"));
}

export function render(): void {
    tooltip().classList.add("hidden");
    store.__tunePane?.dispose();
    const pane = new Pane({ title: "Tune", expanded: !collapsed.has(PANE) });
    store.__tunePane = pane;
    pane.on("fold", (e) => (e.expanded ? collapsed.delete(PANE) : collapsed.add(PANE)));
    pane.on("change", saveTuning);

    addCopyButton(pane, "Copy all values", tuneGroups);

    for (const group of tuneGroups()) {
        const folder = pane.addFolder({ title: group.name, expanded: !collapsed.has(group.name) });
        folder.on("fold", (e) => (e.expanded ? collapsed.delete(group.name) : collapsed.add(group.name)));
        for (const [key, knob] of Object.entries(group.knobs)) {
            const binding = folder.addBinding(group.values, key, { min: knob.min, max: knob.max, step: knob.step });
            attachHint(binding.element, knob.hint);
        }
        addCopyButton(folder, "Copy values", () => [group]);
        folder.addButton({ title: "Reset to code defaults" }).on("click", () => {
            for (const [key, knob] of Object.entries(group.knobs)) group.values[key] = knob.value;
            folder.refresh();
            saveTuning();
        });
    }
}

// Rebuild when this file itself is hot-reloaded (tune.ts only calls render on a new group).
render();
