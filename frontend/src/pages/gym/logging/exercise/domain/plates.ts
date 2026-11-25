import type {
    Exercise,
    ExerciseLoadType,
    LoggedSet,
    PlateMode,
    PlateSetup,
} from "~/types/workout";

/**
 * Plates first: for plate-loaded exercises the plates are what you set, and the weight is what they
 * add up to (base + plates). Everything here is pure, so the weight-progression code can call it:
 * `platesForWeight` turns any target weight into plates (changing as few as possible from the
 * current ones), and `loadFromPlates` goes the other way.
 */

/** Plates in the gym, heaviest first. 2.5s are rarely used: the solver only reaches for them when
 *  nothing heavier lands on the target. */
export const AVAILABLE_PLATES_LBS: readonly number[] = [45, 35, 25, 10, 5, 2.5];

/** Plate weight ("45", "2.5") → count. Per side in per_side mode, in all in total mode. */
export type PlateCounts = Record<string, number>;

/** Where the plates go and what they sit on, from the exercise. */
export type PlateConfig = { base: number; mode: PlateMode };

const EPS = 1e-6;
const round2 = (n: number) => Math.round(n * 100) / 100;

/* ---------- load types and exercise settings ---------- */

/** Maps legacy values (plate_loaded_with_bar / _without_bar) and missing ones to the current set. */
export function normalizeLoadType(raw: string | null | undefined): ExerciseLoadType {
    switch (raw) {
        case "plate_loaded_total":
        case "weight_stack":
        case "free_weights":
            return raw;
        default:
            return "plate_loaded_per_side";
    }
}

export function plateModeFor(loadType: string | null | undefined): PlateMode | null {
    const lt = normalizeLoadType(loadType);
    if (lt === "plate_loaded_per_side") return "per_side";
    if (lt === "plate_loaded_total") return "total";
    return null;
}

export function isPlateLoaded(loadType: string | null | undefined): boolean {
    return plateModeFor(loadType) != null;
}

/** Base weight when the exercise doesn't set one: a 45 lb bar for plates per side, else 0. */
export function defaultBaseWeight(loadType: string | null | undefined): number {
    if (loadType === "plate_loaded_without_bar") return 0;
    return normalizeLoadType(loadType) === "plate_loaded_per_side" ? 45 : 0;
}

/**
 * How far the weight moves when the exercise doesn't set its own step: the automatic increase
 * after hitting the rollover, and +/- on the weight. 5 lb for every load type; set 10 on an
 * exercise (e.g. a barbell lift, +5 a side) to go up faster.
 */
export const DEFAULT_WEIGHT_STEP_LBS = 5;

type ExerciseWeights = Pick<Exercise, "load_type" | "base_weight" | "weight_step">;

const positive = (n: number | null | undefined): n is number =>
    typeof n === "number" && Number.isFinite(n) && n > 0;

/** The exercise's weight step, or the default (5 lb). */
export function weightStepFor(exercise: Partial<ExerciseWeights> | null | undefined): number {
    if (positive(exercise?.weight_step)) return exercise.weight_step;
    return DEFAULT_WEIGHT_STEP_LBS;
}

/** The exercise's base weight, or its load type's default. */
export function baseWeightFor(exercise: Partial<ExerciseWeights> | null | undefined): number {
    const b = exercise?.base_weight;
    if (typeof b === "number" && Number.isFinite(b) && b >= 0) return b;
    return defaultBaseWeight(exercise?.load_type);
}

/** Null for exercises that aren't plate-loaded. */
export function plateConfigFor(exercise: Partial<ExerciseWeights> | null | undefined): PlateConfig | null {
    const mode = plateModeFor(exercise?.load_type);
    if (!mode) return null;
    return { base: baseWeightFor(exercise), mode };
}

/* ---------- plate counts ---------- */

/** Drops zero, negative and non-numeric entries, and normalises keys ("45.0" → "45"). */
export function cleanPlates(plates: Record<string, number> | null | undefined): PlateCounts {
    const out: PlateCounts = {};
    for (const [key, count] of Object.entries(plates ?? {})) {
        const lbs = Number(key);
        const n = Math.round(Number(count));
        if (!(lbs > 0) || !(n > 0)) continue;
        const k = String(lbs);
        out[k] = (out[k] ?? 0) + n;
    }
    return out;
}

/** Sum of the plates as counted (one side in per_side mode). */
export function plateSum(plates: PlateCounts): number {
    let sum = 0;
    for (const [key, count] of Object.entries(plates)) {
        const lbs = Number(key);
        if (lbs > 0 && count > 0) sum += lbs * count;
    }
    return round2(sum);
}

const sideMultiplier = (mode: PlateMode) => (mode === "per_side" ? 2 : 1);

/** Total weight: base plus the plates (both sides in per_side mode). */
export function loadFromPlates(config: PlateConfig, plates: PlateCounts): number {
    return round2(config.base + sideMultiplier(config.mode) * plateSum(plates));
}

/** Adds (`delta` > 0) or removes plates of one weight; never goes below zero. */
export function addPlate(plates: PlateCounts, plate: number, delta: number): PlateCounts {
    const key = String(plate);
    const next = Math.max(0, (plates[key] ?? 0) + delta);
    const { [key]: _removed, ...rest } = plates;
    return next > 0 ? { ...rest, [key]: next } : rest;
}

/** One entry per plate, heaviest first (e.g. for drawing a bar). */
export function plateStack(plates: PlateCounts): number[] {
    const out: number[] = [];
    for (const key of sortedKeys(plates)) {
        for (let i = 0; i < (plates[key] ?? 0); i++) out.push(Number(key));
    }
    return out;
}

function sortedKeys(plates: PlateCounts): string[] {
    return Object.keys(plates)
        .filter((k) => (plates[k] ?? 0) > 0 && Number(k) > 0)
        .sort((a, b) => Number(b) - Number(a));
}

function mergePlates(a: PlateCounts, b: PlateCounts): PlateCounts {
    const out = { ...a };
    for (const [k, n] of Object.entries(b)) out[k] = (out[k] ?? 0) + n;
    return cleanPlates(out);
}

/* ---------- weight → plates ---------- */

export type PlatesForWeight = {
    plates: PlateCounts;
    /** What the plates add up to (base included). */
    load: number;
    /** False when the target can't be loaded exactly; `load` is then the nearest weight that can. */
    exact: boolean;
};

/** Heaviest-first fill of `amount`; `rest` is what couldn't be covered. */
function greedy(amount: number, available: readonly number[]): { plates: PlateCounts; rest: number } {
    const plates: PlateCounts = {};
    let rest = amount;
    for (const plate of [...available].sort((a, b) => b - a)) {
        const n = Math.floor((rest + EPS) / plate);
        if (n > 0) {
            plates[String(plate)] = n;
            rest = round2(rest - n * plate);
        }
    }
    return { plates, rest };
}

/** Removes plates from `plates` summing to exactly `amount`, taking off as few as possible. */
function removeExactly(plates: PlateCounts, amount: number): PlateCounts | null {
    const keys = sortedKeys(plates);
    let best: { remove: number[]; count: number } | null = null;
    const remove = keys.map(() => 0);
    const walk = (i: number, left: number, count: number) => {
        if (Math.abs(left) < EPS) {
            if (!best || count < best.count) best = { remove: [...remove], count };
            return;
        }
        if (i >= keys.length || left < 0 || (best && count >= best.count)) return;
        const lbs = Number(keys[i]);
        const have = plates[keys[i]] ?? 0;
        for (let n = Math.min(have, Math.floor((left + EPS) / lbs)); n >= 0; n--) {
            remove[i] = n;
            walk(i + 1, round2(left - n * lbs), count + n);
        }
        remove[i] = 0;
    };
    walk(0, round2(amount), 0);
    if (!best) return null;
    const chosen: { remove: number[] } = best;
    const out = { ...plates };
    keys.forEach((k, i) => (out[k] = (out[k] ?? 0) - chosen.remove[i]));
    return cleanPlates(out);
}

/**
 * Plates for a target weight. With `previous` plates it changes as little as possible: going up
 * adds plates on top (135 → 145 adds a 5 a side), going down takes plates off. Otherwise, or when
 * that can't land exactly, it loads heaviest-first. A target that can't be loaded gets the nearest
 * weight that can (the lower one on a tie), with `exact: false`.
 */
export function platesForWeight(
    target: number,
    config: PlateConfig,
    previous?: PlateCounts | null,
    available: readonly number[] = AVAILABLE_PLATES_LBS,
): PlatesForWeight {
    const result = (plates: PlateCounts, exact: boolean): PlatesForWeight => ({
        plates,
        load: loadFromPlates(config, plates),
        exact,
    });
    if (!Number.isFinite(target) || target <= config.base + EPS) {
        return result({}, Math.abs(target - config.base) < EPS);
    }
    const want = round2((target - config.base) / sideMultiplier(config.mode));

    const prev = cleanPlates(previous);
    if (Object.keys(prev).length > 0) {
        const delta = round2(want - plateSum(prev));
        if (Math.abs(delta) < EPS) return result(prev, true);
        if (delta > 0) {
            const add = greedy(delta, available);
            if (Math.abs(add.rest) < EPS) return result(mergePlates(prev, add.plates), true);
        } else {
            const removed = removeExactly(prev, -delta);
            if (removed) return result(removed, true);
        }
    }

    const fill = greedy(want, available);
    if (Math.abs(fill.rest) < EPS) return result(fill.plates, true);

    // Not loadable: nearest of the weight just below and the one just above.
    const smallest = Math.min(...available);
    const up = greedy(Math.ceil(want / smallest) * smallest, available).plates;
    const below = loadFromPlates(config, fill.plates);
    const above = loadFromPlates(config, up);
    return result(above - target < target - below ? up : fill.plates, false);
}

/** Logged weight minus what the plates add up to, or null when they agree (within 0.25 lb). */
export function plateMismatch(weight: number, load: number): number | null {
    const delta = round2(weight - load);
    return Math.abs(delta) > 0.25 ? delta : null;
}

/* ---------- text (the weight_setup column) ---------- */

/** "45 + 25 + 2×10", heaviest first; "" for no plates. */
export function formatPlates(plates: PlateCounts): string {
    return sortedKeys(plates)
        .map((k) => {
            const n = plates[k] ?? 0;
            return n === 1 ? k : `${n}×${k}`;
        })
        .join(" + ");
}

/**
 * Text for the legacy `weight_setup` column, in the format older sets used (and `parseLegacyPlates`
 * reads): plates, plus ", bar N" when a per-side base isn't the 45 lb default.
 */
export function plateSetupText(config: PlateConfig, plates: PlateCounts): string {
    const text = formatPlates(plates);
    if (config.mode !== "per_side" || config.base === 45) return text;
    return text ? `${text}, bar ${config.base}` : `bar ${config.base}`;
}

function parsePlateToken(token: string): { plate: number; count: number } | null {
    const mul = token.match(/^(\d+)\s*[x×]\s*(\d+(?:\.\d+)?)$/i);
    if (mul) return { count: Number(mul[1]), plate: Number(mul[2]) };
    const single = token.match(/^(\d+(?:\.\d+)?)$/);
    if (single) return { count: 1, plate: Number(single[1]) };
    return null;
}

/**
 * Plates from old `weight_setup` text ("2×45 + 10", "45 + 25, bar 35"). The bar part is ignored:
 * the base comes from the exercise. Null for free text that isn't plate notation; {} for "".
 */
export function parseLegacyPlates(text: string | null | undefined): PlateCounts | null {
    const raw = (text ?? "").trim();
    if (!raw) return {};
    const body = raw.replace(/(?:^|,\s*)bar\s*\d+(?:\.\d+)?\s*$/i, "").replace(/,\s*$/, "").trim();
    const plates: PlateCounts = {};
    for (const token of body.split("+").map((t) => t.trim()).filter(Boolean)) {
        const piece = parsePlateToken(token);
        if (!piece || !(piece.plate > 0)) return null;
        const key = String(piece.plate);
        plates[key] = (plates[key] ?? 0) + piece.count;
    }
    return cleanPlates(plates);
}

/* ---------- sets ---------- */

/** A set's plates: the stored setup, else parsed from its text. Null when neither has plates. */
export function platesOfSet(
    set: Pick<LoggedSet, "plate_setup" | "weight_setup"> | null | undefined,
): PlateCounts | null {
    if (!set) return null;
    if (set.plate_setup) return cleanPlates(set.plate_setup.plates);
    const parsed = parseLegacyPlates(set.weight_setup);
    return parsed && Object.keys(parsed).length > 0 ? parsed : null;
}

/** The snapshot stored on a set. */
export function toPlateSetup(config: PlateConfig, plates: PlateCounts): PlateSetup {
    return { base: config.base, mode: config.mode, plates: cleanPlates(plates) };
}
