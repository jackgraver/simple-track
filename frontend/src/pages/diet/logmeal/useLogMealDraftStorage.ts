import type { ComputedRef, Ref } from "vue";
import { computed } from "vue";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";
import type { Food, Meal, MealItem } from "~/types/diet";
import { PAGE_MODE, type LogMealPageMode } from "./logmealMode";

const DRAFT_VERSION = 1;

type DraftFoodSnapshot = Pick<
    Food,
    | "ID"
    | "name"
    | "serving_type"
    | "serving_amount"
    | "calories"
    | "protein"
    | "fiber"
    | "carbs"
    | "fat"
    | "variant_group_id"
> & {
    variants?: DraftFoodSnapshot[];
};

type LogMealDraftItemSnapshot = {
    food_id: number;
    amount: number;
    group_id: string;
    group_label: string;
    composite_food_id: number | null;
    food: DraftFoodSnapshot;
};

export type LogMealDraftSnapshot = {
    version: number;
    updatedAt: number;
    name: string;
    items: LogMealDraftItemSnapshot[];
};

function snapshotFood(food: Food): DraftFoodSnapshot {
    const variants = food.variants?.map((v) => ({
        ID: v.ID,
        name: v.name,
        serving_type: v.serving_type,
        serving_amount: v.serving_amount,
        calories: v.calories,
        protein: v.protein,
        fiber: v.fiber,
        carbs: v.carbs,
        fat: v.fat,
        variant_group_id: v.variant_group_id ?? null,
    }));
    return {
        ID: food.ID,
        name: food.name,
        serving_type: food.serving_type,
        serving_amount: food.serving_amount,
        calories: food.calories,
        protein: food.protein,
        fiber: food.fiber,
        carbs: food.carbs,
        fat: food.fat,
        variant_group_id: food.variant_group_id ?? null,
        ...(variants?.length ? { variants } : {}),
    };
}

function isDraftFoodSnapshot(v: unknown): v is DraftFoodSnapshot {
    if (!v || typeof v !== "object") return false;
    const f = v as DraftFoodSnapshot;
    return (
        typeof f.ID === "number" &&
        typeof f.name === "string" &&
        typeof f.serving_type === "string" &&
        typeof f.serving_amount === "number" &&
        typeof f.calories === "number" &&
        typeof f.protein === "number" &&
        typeof f.fiber === "number" &&
        typeof f.carbs === "number" &&
        typeof f.fat === "number"
    );
}

function isDraftItemSnapshot(v: unknown): v is LogMealDraftItemSnapshot {
    if (!v || typeof v !== "object") return false;
    const it = v as LogMealDraftItemSnapshot;
    return (
        typeof it.food_id === "number" &&
        typeof it.amount === "number" &&
        typeof it.group_id === "string" &&
        typeof it.group_label === "string" &&
        (it.composite_food_id === null ||
            typeof it.composite_food_id === "number") &&
        isDraftFoodSnapshot(it.food)
    );
}

function draftHasContent(snapshot: LogMealDraftSnapshot): boolean {
    return snapshot.name.trim().length > 0 || snapshot.items.length > 0;
}

function snapshotToMealItem(
    row: LogMealDraftItemSnapshot,
    mealId: number,
): MealItem {
    const { variants: draftVariants, ...foodBase } = row.food;
    const variants = draftVariants?.map(
        (v) =>
            ({
                ...v,
                created_at: "",
                updated_at: "",
            }) as Food,
    );
    const food: Food = {
        ...foodBase,
        created_at: "",
        updated_at: "",
        ...(variants?.length ? { variants } : {}),
    };
    return {
        ID: 0,
        created_at: "",
        updated_at: "",
        meal_id: mealId,
        food_id: row.food_id,
        food,
        amount: row.amount,
        group_id: row.group_id,
        group_label: row.group_label,
        composite_food_id: row.composite_food_id,
    };
}

function mealItemToSnapshot(item: MealItem): LogMealDraftItemSnapshot | null {
    if (!item.food || typeof item.food_id !== "number") return null;
    return {
        food_id: item.food_id,
        amount: Number(item.amount),
        group_id: item.group_id ?? "",
        group_label: item.group_label ?? "",
        composite_food_id: item.composite_food_id ?? null,
        food: snapshotFood(item.food),
    };
}

export function useLogMealDraftStorage(options: {
    meal: Ref<Meal>;
    pageMode: ComputedRef<LogMealPageMode>;
    routeMealId: ComputedRef<number>;
}) {
    const storageKey = computed(() => {
        if (
            (options.pageMode.value !== PAGE_MODE.log &&
                options.pageMode.value !== PAGE_MODE.create) ||
            options.routeMealId.value !== 0
        ) {
            return "";
        }
        if (options.pageMode.value === PAGE_MODE.log) {
            return "simpletracker:diet:logmeal-draft:v2:log";
        }
        return "simpletracker:diet:logmeal-draft:v2:create";
    });

    const sync = useWebStorageJsonSync<LogMealDraftSnapshot>({
        storage: window.sessionStorage,
        key: storageKey,
        watchSources: [options.meal],
        deep: true,
        canPersist: () => storageKey.value !== "",
        shouldPersist: draftHasContent,
        getSnapshot: () => ({
            version: DRAFT_VERSION,
            updatedAt: Date.now(),
            name: options.meal.value.name,
            items: options.meal.value.items
                .map(mealItemToSnapshot)
                .filter((row): row is LogMealDraftItemSnapshot => row !== null),
        }),
        tryRestore: (parsed, { remove }) => {
            if (parsed.version !== DRAFT_VERSION) {
                remove();
                return false;
            }
            if (typeof parsed.name !== "string") {
                remove();
                return false;
            }
            if (!Array.isArray(parsed.items)) {
                remove();
                return false;
            }
            if (
                !parsed.items.every(isDraftItemSnapshot) ||
                !draftHasContent({
                    version: DRAFT_VERSION,
                    updatedAt: parsed.updatedAt ?? 0,
                    name: parsed.name,
                    items: parsed.items,
                })
            ) {
                remove();
                return false;
            }
            const items = parsed.items.map((row) =>
                snapshotToMealItem(row, options.meal.value.ID),
            );
            options.meal.value = {
                ...options.meal.value,
                name: parsed.name,
                items,
            };
            return true;
        },
    });

    function clearAndPause() {
        sync.setSaveEnabled(false);
        sync.clear();
    }

    return {
        storageKey,
        restore: sync.restore,
        setSaveEnabled: sync.setSaveEnabled,
        clearAndPause,
    };
}
