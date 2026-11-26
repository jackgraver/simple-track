// The Tweakpane panel behind `?tune` (see tune.ts). Only ever loaded by a dynamic import in dev.
import { Pane } from "tweakpane";
import { saveTuning, tuneGroups } from "./tune";

// One panel per page, whatever hot reloads do: the live one is kept on globalThis and replaced on
// each render, which rebuilds it from every registered group (one folder each).
const store = globalThis as { __tunePane?: Pane; __tuneCollapsed?: Set<string> };
const collapsed = (store.__tuneCollapsed ??= new Set<string>());
const PANE = "\u0000pane";

function format(): string {
    return tuneGroups()
        .map((g) => `${g.name}: ${Object.entries(g.values).map(([k, v]) => `${k} ${v}`).join(", ")}`)
        .join("\n");
}

export function render(): void {
    store.__tunePane?.dispose();
    const pane = new Pane({ title: "Tune", expanded: !collapsed.has(PANE) });
    store.__tunePane = pane;
    pane.on("fold", (e) => (e.expanded ? collapsed.delete(PANE) : collapsed.add(PANE)));
    pane.on("change", saveTuning);

    const copy = pane.addButton({ title: "Copy values" });
    copy.on("click", () => {
        console.info(`[tune]\n${format()}`);
        void navigator.clipboard.writeText(format()).then(
            () => (copy.title = "Copied"),
            () => (copy.title = "Copy failed (values in console)"),
        );
        setTimeout(() => (copy.title = "Copy values"), 1500);
    });

    for (const group of tuneGroups()) {
        const folder = pane.addFolder({ title: group.name, expanded: !collapsed.has(group.name) });
        folder.on("fold", (e) => (e.expanded ? collapsed.delete(group.name) : collapsed.add(group.name)));
        for (const [key, knob] of Object.entries(group.knobs)) {
            folder.addBinding(group.values, key, { min: knob.min, max: knob.max, step: knob.step });
        }
        folder.addButton({ title: "Reset to code defaults" }).on("click", () => {
            for (const [key, knob] of Object.entries(group.knobs)) group.values[key] = knob.value;
            folder.refresh();
            saveTuning();
        });
    }
}

// Rebuild when this file itself is hot-reloaded (tune.ts only calls render on a new group).
render();
