import type { MealItem, SavedMealItem } from "~/types/diet";
import {
    dateAtDietDayOffset,
    dietDayOffsetFromLocalDate,
    formatDateLong,
} from "~/utils/dateUtil";

export type MacroSummary = {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
};
const ZERO_MACROS: MacroSummary = {
    calories: 0,
    protein: 0,
    fat: 0,
    carbs: 0,
    fiber: 0,
};

export function macroTotals(
    items: (MealItem | SavedMealItem)[] | undefined,
): MacroSummary {
    if (!items?.length) return ZERO_MACROS;
    return items.reduce(
        (acc, i) => ({
            calories: acc.calories + (i.food?.calories ?? 0) * i.amount,
            protein: acc.protein + (i.food?.protein ?? 0) * i.amount,
            carbs: acc.carbs + (i.food?.carbs ?? 0) * i.amount,
            fat: acc.fat + (i.food?.fat ?? 0) * i.amount,
            fiber: acc.fiber + (i.food?.fiber ?? 0) * i.amount,
        }),
        { ...ZERO_MACROS },
    );
}

export function daysInMonth(year: number, month: number): number {
    return new Date(year, month + 1, 0).getDate();
}

function parseIsoLocal(
    iso: string,
): { y: number; m: number; d: number } | null {
    const [y, mo, day] = iso.split("-").map(Number);
    if (!y || !mo || !day) return null;
    return { y, m: mo - 1, d: day };
}

export function enumerateOffsetsInclusive(
    fromIso: string,
    toIso: string,
): number[] | null {
    const a = parseIsoLocal(fromIso);
    const b = parseIsoLocal(toIso);
    if (!a || !b) return null;
    let t1 = new Date(a.y, a.m, a.d);
    let t2 = new Date(b.y, b.m, b.d);
    if (t1.getTime() > t2.getTime()) [t1, t2] = [t2, t1];
    const out: number[] = [];
    const cur = new Date(t1);
    while (cur.getTime() <= t2.getTime()) {
        out.push(
            dietDayOffsetFromLocalDate(
                cur.getFullYear(),
                cur.getMonth(),
                cur.getDate(),
            ),
        );
        cur.setDate(cur.getDate() + 1);
    }
    return out;
}

export function offsetToHeadingLabel(offset: number): string {
    const d = dateAtDietDayOffset(offset);
    const ymd = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T12:00:00`;
    return formatDateLong(ymd);
}
