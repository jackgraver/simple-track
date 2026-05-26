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
    if (percent >= 105) return "text-(--color-cf-red)";
    if (percent >= 95) return "text-green-600";
    if (percent >= 85) return "text-green-400";
    return "";
}

export function determineOverflow(
    total: number,
    planned: number,
    indicateOverflow: boolean | undefined,
): string {
    if (planned <= 0 || !indicateOverflow) return "";
    const overflow = (total / planned) * 100 - 100;

    if (overflow > 20) return "text-[red]";
    if (overflow > 10) return "text-[orange]";
    if (overflow > 0) return "text-[yellow]";
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

export const macroFillClass: Record<MacroBarNutrientType, string> = {
    calories: "calories",
    protein: "protein",
    fiber: "fiber",
    carbs: "carbs",
    fat: "fat",
    water: "water",
};
