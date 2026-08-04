import type { ExerciseLoadType } from "~/types/workout";

export const STANDARD_PLATE_LBS = [45, 35, 25, 10, 5, 2.5] as const;
export const DEFAULT_BAR_LBS = 45;

export type PlateCountsPerSide = Partial<Record<number, number>>;

export type ParsedWeightSetup = {
    barLbs: number;
    platesPerSide: PlateCountsPerSide;
    parsed: boolean;
};

export function roundWeightHalfLb(weight: number): number {
    return Math.round(weight * 2) / 2;
}

export function perSideLoadLbs(platesPerSide: PlateCountsPerSide): number {
    let load = 0;
    for (const [plate, count] of Object.entries(platesPerSide)) {
        const n = Number(plate);
        const c = count ?? 0;
        if (c > 0 && n > 0) load += n * c;
    }
    return load;
}

export function computeBarbellTotalLbs(
    barLbs: number,
    platesPerSide: PlateCountsPerSide,
): number {
    return roundWeightHalfLb(barLbs + 2 * perSideLoadLbs(platesPerSide));
}

export function barLbsForExercise(loadType: ExerciseLoadType): number {
    return loadType === "plate_loaded_with_bar" ? DEFAULT_BAR_LBS : 0;
}

export function computeExerciseLoadLbs(
    loadType: ExerciseLoadType,
    platesPerSide: PlateCountsPerSide,
): number {
    if (loadType === "weight_stack") return 0;
    return computeBarbellTotalLbs(
        barLbsForExercise(loadType),
        platesPerSide,
    );
}

export function formatWeightSetup(
    barLbs: number,
    platesPerSide: PlateCountsPerSide,
): string {
    const parts: string[] = [];
    for (const plate of STANDARD_PLATE_LBS) {
        const count = platesPerSide[plate] ?? 0;
        if (count <= 0) continue;
        parts.push(
            count === 1 ? String(plate) : `${count}×${plate}`,
        );
    }
    const plates = parts.join(" + ");
    if (plates.length === 0) {
        if (barLbs === 0) return "";
        return barLbs !== DEFAULT_BAR_LBS ? `bar ${barLbs}` : "";
    }
    if (barLbs === 0) {
        return `${plates}, bar 0`;
    }
    if (barLbs !== DEFAULT_BAR_LBS) {
        return `${plates}, bar ${barLbs}`;
    }
    return plates;
}

export function formatExerciseWeightSetup(
    loadType: ExerciseLoadType,
    platesPerSide: PlateCountsPerSide,
): string {
    if (loadType === "weight_stack") return "";
    return formatWeightSetup(barLbsForExercise(loadType), platesPerSide);
}

export function weightSetupForExercise(
    loadType: ExerciseLoadType,
    setup: string,
): string {
    return loadType === "weight_stack" ? "" : setup;
}

export function weightSetupMismatchLbs(
    computedLbs: number,
    targetWeightLbs: number,
): number | null {
    const delta = roundWeightHalfLb(computedLbs) - roundWeightHalfLb(targetWeightLbs);
    if (Math.abs(delta) <= 0.5) return null;
    return delta;
}

function parsePlateToken(token: string): { plate: number; count: number } | null {
    const trimmed = token.trim();
    if (!trimmed) return null;
    const mul = trimmed.match(/^(\d+)\s*[x×]\s*(\d+(?:\.\d+)?)$/i);
    if (mul) {
        return { count: Number(mul[1]), plate: Number(mul[2]) };
    }
    const single = trimmed.match(/^(\d+(?:\.\d+)?)$/);
    if (single) {
        return { count: 1, plate: Number(single[1]) };
    }
    return null;
}

export function parseWeightSetup(text: string): ParsedWeightSetup {
    const empty: ParsedWeightSetup = {
        barLbs: DEFAULT_BAR_LBS,
        platesPerSide: {},
        parsed: false,
    };
    const raw = text.trim();
    if (!raw) return { ...empty, parsed: true };

    let barLbs = DEFAULT_BAR_LBS;
    let body = raw;
    const barMatch = raw.match(/(?:^|,\s*)bar\s*(\d+(?:\.\d+)?)\s*$/i);
    if (barMatch) {
        barLbs = Number(barMatch[1]);
        body = raw.slice(0, barMatch.index).replace(/,\s*$/, "").trim();
    }

    const platesPerSide: PlateCountsPerSide = {};
    const tokens = body.split("+").map((t) => t.trim()).filter(Boolean);
    if (tokens.length === 0) {
        return { barLbs, platesPerSide, parsed: barMatch != null };
    }

    for (const token of tokens) {
        const piece = parsePlateToken(token);
        if (!piece) return empty;
        platesPerSide[piece.plate] = (platesPerSide[piece.plate] ?? 0) + piece.count;
    }

    return { barLbs, platesPerSide, parsed: true };
}

export type WeightSetupDialogResult = {
    weightSetup: string;
    totalLbs: number;
    syncWeight: boolean;
};
