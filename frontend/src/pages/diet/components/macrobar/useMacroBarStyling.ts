export type MacroBarNutrientType =
    | "calories"
    | "protein"
    | "fiber"
    | "carbs"
    | "fat"
    | "water";

export function formatInt(n: number): string {
    return String(Math.round(n));
}

export function calcWidth(total: number, planned: number): number {
    if (planned <= 0) {
        return total > 0 ? 100 : 0;
    }
    return Math.min(100, (total / planned) * 100);
}

export function calcPercentage(total: number, planned: number): number | null {
    if (!planned || Number.isNaN(total)) return null;
    return Math.round((total / planned) * 100);
}

export function formatPercentage(total: number, planned: number): string {
    const percent = calcPercentage(total, planned);
    return percent == null ? "" : `(${percent}%)`;
}

export function percentageColorClass(percent: number | null): string {
    if (percent == null) return "";
    if (percent >= 105) return "text-danger";
    if (percent >= 85) return "text-success";
    return "text-fg-muted";
}

export function determineOverflow(
    total: number,
    planned: number,
    indicateOverflow: boolean | undefined,
): string {
    if (planned <= 0 || !indicateOverflow) return "";
    const overflow = (total / planned) * 100 - 100;

    if (overflow > 20) return "text-danger";
    if (overflow > 0) return "text-warn";
    return "";
}

export const typeLabels: Record<MacroBarNutrientType, string> = {
    calories: "Calories",
    protein: "Protein",
    fiber: "Fiber",
    carbs: "Carbs",
    fat: "Fat",
    water: "Water",
};

// Full class strings so Tailwind can see them.
export const macroFillClass: Record<MacroBarNutrientType, string> = {
    calories: "bg-macro-calories",
    protein: "bg-macro-protein",
    fiber: "bg-macro-fiber",
    carbs: "bg-macro-carbs",
    fat: "bg-macro-fat",
    water: "bg-macro-water",
};
