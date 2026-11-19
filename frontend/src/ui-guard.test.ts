/**
 * UI guard (ui-overhaul project, see docs/projects/ui-overhaul/RULES.md).
 *
 * A ratchet: counts design-system violations per file and fails only when a
 * file has MORE than its entry in ui-guard.baseline.json (new files start at
 * 0). Counts going down always pass; lower the baseline in the same PR with
 *
 *   UI_GUARD_UPDATE=1 npx vitest run src/ui-guard.test.ts
 *
 * and check that the diff only goes down. Never raise a number to make a
 * change pass: use a token or primitive instead.
 */
import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, it } from "vitest";

const SRC = join(__dirname);
const ROOT = join(__dirname, "..");
const BASELINE_PATH = join(__dirname, "ui-guard.baseline.json");

type Metric = "colorLiteral" | "paletteClass" | "arbitraryValue" | "styleBlock";
type Counts = Partial<Record<Metric, number>>;

const PALETTE =
    "red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|slate|gray|zinc|neutral|stone|black|white";

const METRICS: Record<Metric, RegExp> = {
    // #fff, #1b1b1b, rgb(...), rgba(...), hsl(...), hsla(...)
    colorLiteral: /#[0-9a-fA-F]{3,8}\b|\b(?:rgba?|hsla?)\(/g,
    // Tailwind default palette classes: bg-zinc-800, text-white, ring-amber-700/40
    paletteClass: new RegExp(
        `\\b(?:bg|text|border|ring|ring-offset|outline|fill|stroke|from|via|to|divide|decoration|caret|accent|shadow|placeholder)-(?:${PALETTE})(?:-\\d{2,3})?(?:/\\d+)?\\b`,
        "g",
    ),
    // Arbitrary values for things tokens cover: text-[10px], bg-[rgb(...)], rounded-[5px]
    arbitraryValue:
        /\b(?:text|bg|border|ring|ring-offset|outline|fill|stroke|rounded|shadow|z)-\[[^\]]+\]/g,
    styleBlock: /<style\b/g,
};

function walk(dir: string): string[] {
    return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
        const p = join(dir, e.name);
        if (e.isDirectory()) return walk(p);
        if (!/\.(vue|ts)$/.test(e.name) || /\.(test|d)\.ts$/.test(e.name)) return [];
        return [p];
    });
}

function scan(): Record<string, Counts> {
    const out: Record<string, Counts> = {};
    for (const file of walk(SRC).sort()) {
        const text = readFileSync(file, "utf8");
        const counts: Counts = {};
        for (const [name, re] of Object.entries(METRICS) as [Metric, RegExp][]) {
            const n = text.match(re)?.length ?? 0;
            if (n > 0) counts[name] = n;
        }
        if (Object.keys(counts).length > 0) out[relative(ROOT, file)] = counts;
    }
    return out;
}

function loadBaseline(): Record<string, Counts> {
    try {
        return JSON.parse(readFileSync(BASELINE_PATH, "utf8"));
    } catch {
        return {};
    }
}

describe("ui guard", () => {
    const current = scan();

    if (process.env.UI_GUARD_UPDATE) {
        writeFileSync(BASELINE_PATH, JSON.stringify(current, null, 2) + "\n");
    }

    it("adds no colour literals, palette classes, arbitrary values or <style> blocks", () => {
        const baseline = loadBaseline();
        const regressions: string[] = [];
        for (const [file, counts] of Object.entries(current)) {
            for (const [metric, n] of Object.entries(counts) as [Metric, number][]) {
                const allowed = baseline[file]?.[metric] ?? 0;
                if (n > allowed) {
                    regressions.push(`${file}: ${metric} ${n} > baseline ${allowed}`);
                }
            }
        }
        expect(
            regressions,
            "Use tokens/primitives instead (docs/projects/ui-overhaul/RULES.md)",
        ).toEqual([]);
    });
});
