import type { MealItem } from "~/types/diet";

export type MealItemWithIndex = { item: MealItem; index: number };

export type MealItemDisplayBlock =
    | { kind: "ungrouped"; rows: MealItemWithIndex[] }
    | { kind: "group"; groupId: string; label: string; rows: MealItemWithIndex[] };

/** Build collapsible blocks: each non-empty group_id is one group (all matching indices), in first-seen order. */
export function mealItemsToDisplayBlocks(items: MealItem[]): MealItemDisplayBlock[] {
    const seen = new Set<string>();
    const blocks: MealItemDisplayBlock[] = [];
    for (let i = 0; i < items.length; i++) {
        const it = items[i];
        const gid = (it.group_id ?? "").trim();
        if (!gid) {
            blocks.push({ kind: "ungrouped", rows: [{ item: it, index: i }] });
            continue;
        }
        if (seen.has(gid)) continue;
        seen.add(gid);
        const rows: MealItemWithIndex[] = [];
        items.forEach((item, idx) => {
            if ((item.group_id ?? "").trim() === gid) {
                rows.push({ item, index: idx });
            }
        });
        rows.sort((a, b) => a.index - b.index);
        blocks.push({
            kind: "group",
            groupId: gid,
            label: it.group_label ?? "",
            rows,
        });
    }
    return blocks;
}

export function blockMacros(rows: MealItemWithIndex[]): {
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
} {
    return rows.reduce(
        (acc, { item }) => {
            const a = Number(item.amount);
            return {
                calories: acc.calories + (item.food?.calories ?? 0) * a,
                protein: acc.protein + (item.food?.protein ?? 0) * a,
                fiber: acc.fiber + (item.food?.fiber ?? 0) * a,
                carbs: acc.carbs + (item.food?.carbs ?? 0) * a,
            };
        },
        { calories: 0, protein: 0, fiber: 0, carbs: 0 },
    );
}
