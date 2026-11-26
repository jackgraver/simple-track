// The Tweakpane panel behind `?tune` (see tune.ts). Only ever loaded by a dynamic import in dev.
import { Pane } from "tweakpane";
import { saveTuning, type TuneGroup } from "./tune";

const added: TuneGroup[] = [];
let pane: Pane | null = null;

function format(): string {
    return added.map((g) => `${g.name}: ${Object.entries(g.values).map(([k, v]) => `${k} ${v}`).join(", ")}`).join("\n");
}

function root(): Pane {
    if (pane) return pane;
    pane = new Pane({ title: "Tune" });
    const copy = pane.addButton({ title: "Copy values" });
    copy.on("click", () => {
        void navigator.clipboard.writeText(format()).then(
            () => (copy.title = "Copied"),
            () => (copy.title = "Copy failed (see console)"),
        );
        console.info(`[tune]\n${format()}`);
        setTimeout(() => (copy.title = "Copy values"), 1500);
    });
    pane.on("change", saveTuning);
    return pane;
}

export function addGroup(group: TuneGroup): void {
    added.push(group);
    const folder = root().addFolder({ title: group.name });
    for (const [key, knob] of Object.entries(group.knobs)) {
        folder.addBinding(group.values, key, { min: knob.min, max: knob.max, step: knob.step });
    }
    folder.addButton({ title: "Reset to code defaults" }).on("click", () => {
        for (const [key, knob] of Object.entries(group.knobs)) group.values[key] = knob.value;
        folder.refresh();
        saveTuning();
    });
}
