export type MacroBarNutrientType =
    | "calories"
    | "protein"
    | "fiber"
    | "carbs"
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
    water: "Water",
};

export const macroFillClass: Record<MacroBarNutrientType, string> = {
    calories: "bg-[orange]",
    protein: "bg-[#60a5fa]",
    fiber: "bg-[green]",
    carbs: "bg-[red]",
    water: "bg-[blue]",
};
