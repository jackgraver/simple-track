import type { MealItem } from "~/types/diet";

export function formatNum(n: number): number {
    const s = n.toFixed(2);
    return Number(s.replace(/\.?0+$/, ""));
}

export function itemServingAmount(item: MealItem): number {
    return (item.food?.serving_amount || 1) * item.amount;
}
