import { computed, ref, toRaw } from "vue";
import type {
    CompositeFood,
    Food,
    Meal,
    MealItem,
    SavedMeal,
} from "~/types/diet";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { mealItemsToDisplayBlocks } from "~/utils/mealItemGroups";
import { savedMealToMeal } from "~/utils/savedMealToMeal";
import CreateFoodDialog from "./dialog/CreateFoodDialog.vue";
import { formatNum } from "./logmealItemFormat";

export type MealMacroTotals = {
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
    fat: number;
};

export function emptyMeal(): Meal {
    return {
        ID: 0,
        created_at: "",
        updated_at: "",
        name: "",
        items: [],
    };
}

export function macrosForMeal(m: Meal): MealMacroTotals {
    return {
        calories: formatNum(
            m.items.reduce(
                (total, item) =>
                    total + Number(item.amount) * (item.food?.calories ?? 0),
                0,
            ),
        ),
        protein: formatNum(
            m.items.reduce(
                (total, item) =>
                    total + Number(item.amount) * (item.food?.protein ?? 0),
                0,
            ),
        ),
        fiber: formatNum(
            m.items.reduce(
                (total, item) =>
                    total + Number(item.amount) * (item.food?.fiber ?? 0),
                0,
            ),
        ),
        carbs: formatNum(
            m.items.reduce(
                (total, item) =>
                    total + Number(item.amount) * (item.food?.carbs ?? 0),
                0,
            ),
        ),
        fat: formatNum(
            m.items.reduce(
                (total, item) =>
                    total + Number(item.amount) * (item.food?.fat ?? 0),
                0,
            ),
        ),
    };
}

/** Vue Query wraps API data in readonly proxies; clone so we can edit amounts. */
export function cloneMeal(m: Meal): Meal {
    return JSON.parse(JSON.stringify(toRaw(m))) as Meal;
}

export function useEditableMeal() {
    const meal = ref<Meal>(emptyMeal());
    const baselineMealMacros = ref<MealMacroTotals | null>(null);
    const collapsedGroups = ref<Record<string, boolean>>({});
    const selectedForGroup = ref<Record<number, boolean>>({});

    const mealItemBlocks = computed(() =>
        mealItemsToDisplayBlocks(meal.value.items),
    );

    const totalMacros = computed(() => macrosForMeal(meal.value));

    function resetMeal() {
        meal.value = emptyMeal();
    }

    function mealItemGroupKey(item: MealItem): string {
        return (item.group_id ?? "").trim();
    }

    function amountPlusMinus(index: number, direction: "plus" | "minus") {
        const items = meal.value.items;
        const item = items[index];
        if (!item?.food) return;
        const step = 1 / (item.food.serving_amount || 1);
        const amount = Number(item.amount);
        if (direction === "plus") {
            const newAmount = amount + step;
            meal.value.items = items.map((it, i) =>
                i === index ? { ...it, amount: newAmount } : it,
            );
            return;
        }
        const next = amount - step;
        if (next <= 0) {
            meal.value.items = items.filter((_, i) => i !== index);
            selectedForGroup.value = {};
            return;
        }
        meal.value.items = items.map((it, i) =>
            i === index ? { ...it, amount: next } : it,
        );
    }

    function setMealItemAmount(index: number, amount: number) {
        meal.value.items = meal.value.items.map((it, i) =>
            i === index ? { ...it, amount } : it,
        );
    }

    function isGroupExpanded(groupId: string): boolean {
        return !collapsedGroups.value[groupId];
    }

    function toggleGroupCollapse(groupId: string) {
        collapsedGroups.value = {
            ...collapsedGroups.value,
            [groupId]: !collapsedGroups.value[groupId],
        };
    }

    function toggleSelectRow(index: number) {
        const next = { ...selectedForGroup.value };
        if (next[index]) delete next[index];
        else next[index] = true;
        selectedForGroup.value = next;
    }

    function selectedIndices(): number[] {
        return Object.keys(selectedForGroup.value)
            .map(Number)
            .filter((k) => selectedForGroup.value[k]);
    }

    function setGroupLabel(groupId: string, label: string) {
        const gid = groupId.trim();
        if (!gid) return;
        meal.value.items = meal.value.items.map((it) =>
            (it.group_id ?? "").trim() === gid
                ? { ...it, group_label: label }
                : it,
        );
    }

    function groupSelectedRows() {
        const idxs = selectedIndices().sort((a, b) => a - b);
        if (idxs.length < 2) {
            toast.push("Select at least two items to group.", "error");
            return;
        }
        const gid = crypto.randomUUID();
        meal.value.items = meal.value.items.map((it, i) =>
            idxs.includes(i)
                ? {
                      ...it,
                      group_id: gid,
                      group_label: "",
                      composite_food_id: null,
                  }
                : it,
        );
        selectedForGroup.value = {};
    }

    function ungroupSelectedRows() {
        const idxs = selectedIndices();
        if (idxs.length === 0) {
            toast.push("Select items to ungroup.", "error");
            return;
        }
        meal.value.items = meal.value.items.map((it, i) =>
            idxs.includes(i)
                ? {
                      ...it,
                      group_id: "",
                      group_label: "",
                      composite_food_id: null,
                  }
                : it,
        );
        selectedForGroup.value = {};
    }

    function removeGroupLines(indices: number[]) {
        const sorted = [...new Set(indices)].sort((a, b) => b - a);
        const items = [...meal.value.items];
        for (const i of sorted) {
            items.splice(i, 1);
        }
        meal.value.items = items;
        selectedForGroup.value = {};
    }

    async function addFood(food: Food): Promise<boolean> {
        const g = "";
        const existingIndex = meal.value.items.findIndex(
            (i) => i?.food?.ID === food.ID && mealItemGroupKey(i) === g,
        );
        if (existingIndex !== -1) {
            meal.value.items = meal.value.items.map((it, i) =>
                i === existingIndex
                    ? { ...it, amount: Number(it.amount) + 1 }
                    : it,
            );
            return true;
        }

        meal.value.items = [
            ...meal.value.items,
            {
                ID: 0,
                created_at: "",
                updated_at: "",
                meal_id: meal.value.ID,
                food_id: food.ID,
                food: food,
                amount: 1,
                group_id: "",
                group_label: "",
                composite_food_id: null,
            } as MealItem,
        ];
        return true;
    }

    async function addComposite(cf: CompositeFood): Promise<boolean> {
        const gid = crypto.randomUUID();
        const label = cf.name;
        const cfid = cf.ID;
        const newItems: MealItem[] = [];
        for (const line of cf.items) {
            const food = line.food;
            if (!food) continue;
            newItems.push({
                ID: 0,
                created_at: "",
                updated_at: "",
                meal_id: meal.value.ID,
                food_id: line.food_id,
                food,
                amount: line.amount,
                group_id: gid,
                group_label: label,
                composite_food_id: cfid,
            } as MealItem);
        }
        if (newItems.length === 0) return false;
        meal.value.items = [...meal.value.items, ...newItems];
        return true;
    }

    async function pickFoodOrComposite(
        row: Food & { entry_kind?: string } & Partial<CompositeFood>,
    ): Promise<boolean> {
        if (row.entry_kind === "composite") {
            return addComposite(row as CompositeFood);
        }
        return addFood(row as Food);
    }

    async function createFood(name: string): Promise<boolean> {
        try {
            const food = await dialogManager.custom<Food>({
                title: "Create " + name,
                component: CreateFoodDialog,
                componentProps: { foodName: name },
            });

            if (food === null) return false;

            if (food) {
                await addFood(food);
                toast.push("Food Created Successfully and Added!", "success");
                return true;
            }
            toast.push("Food Creation Failed!", "error");
            return false;
        } catch {
            toast.push("Dialog Error", "error");
            return false;
        }
    }

    async function removeFood(index: number) {
        meal.value.items = meal.value.items.filter((_, i) => i !== index);
        selectedForGroup.value = {};
    }

    function variantSiblings(f: Food): Food[] {
        const base = [f, ...(f.variants ?? [])];
        const byId = new Map<number, Food>();
        for (const x of base) {
            if (x?.ID) byId.set(x.ID, x);
        }
        return [...byId.values()].sort((a, b) => a.name.localeCompare(b.name));
    }

    function swapVariant(index: number, v: Food) {
        const it = meal.value.items[index];
        const cur = it?.food;
        if (!cur) return;
        const others = variantSiblings(cur).filter((x) => x.ID !== v.ID);
        const enriched: Food = {
            ...v,
            variant_group_id: v.variant_group_id ?? cur.variant_group_id ?? null,
            variants: others,
        };
        meal.value.items = meal.value.items.map((row, i) =>
            i === index ? { ...row, food_id: v.ID, food: enriched } : row,
        );
    }

    async function setMeal(item: Meal | SavedMeal): Promise<boolean> {
        const first = item.items[0];
        const loaded =
            first && "saved_meal_id" in first
                ? savedMealToMeal(item as SavedMeal)
                : cloneMeal(item as Meal);
        const keepName = meal.value.name.trim();
        meal.value = {
            ...loaded,
            name: keepName || loaded.name,
        };
        return true;
    }

    return {
        meal,
        baselineMealMacros,
        collapsedGroups,
        selectedForGroup,
        mealItemBlocks,
        totalMacros,
        resetMeal,
        amountPlusMinus,
        setMealItemAmount,
        isGroupExpanded,
        toggleGroupCollapse,
        toggleSelectRow,
        setGroupLabel,
        groupSelectedRows,
        ungroupSelectedRows,
        removeGroupLines,
        pickFoodOrComposite,
        createFood,
        removeFood,
        swapVariant,
        setMeal,
    };
}
