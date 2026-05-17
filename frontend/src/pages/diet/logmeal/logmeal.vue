<script setup lang="ts">
import "~/pages/diet/macro-colors.css";
import { useRoute } from "vue-router";
import type {
    CompositeFood,
    Food,
    Meal,
    MealItem,
    SavedMeal,
} from "~/types/diet";
import SearchList from "~/shared/SearchList.vue";
import LogMealGroupBlock from "./LogMealGroupBlock.vue";
import LogMealItemRow from "./LogMealItemRow.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
import { formatNum } from "./logmealItemFormat";
import { mealItemsListGridClass } from "./mealItemsListGrid";
import Input from "~/shared/input/Input.vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";
import CreateFoodDialog from "./dialog/CreateFoodDialog.vue";
import { computed, nextTick, ref, toRaw, watch } from "vue";
import { mealItemsToDisplayBlocks } from "~/utils/mealItemGroups";
import { useMeal } from "./queries/useMeal";
import { useSavedMeal } from "./queries/useSavedMeal";
import { useDietLogsToday } from "./queries/useDietLogsToday";
import {
    useCreateMeal,
    useCreateSavedMeal,
    useLogEditedMeal,
    useUpdateLoggedMeal,
    useUpdateSavedMeal,
} from "./queries/useMealMutations";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import { savedMealToMeal } from "~/utils/savedMealToMeal";
import {
    EDIT_VARIANT,
    PAGE_MODE,
    parseEditMealVariant,
    parseLogMealPageMode,
    type LogMealPageMode,
} from "./logmealMode";
import SimpleMacros from "~/shared/SimpleMacros.vue";

type MealMacroTotals = {
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
    fat: number;
};

type LogMealDraftSnapshot = {
    name: string;
    items: MealItem[];
};

function macrosForMeal(m: Meal): MealMacroTotals {
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
function cloneMeal(m: Meal): Meal {
    return JSON.parse(JSON.stringify(toRaw(m))) as Meal;
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

function mealItemGroupKey(item: MealItem): string {
    return (item.group_id ?? "").trim();
}

const mealItemBlocks = computed(() =>
    mealItemsToDisplayBlocks(meal.value.items),
);

const collapsedGroups = ref<Record<string, boolean>>({});
function isGroupExpanded(groupId: string): boolean {
    return !collapsedGroups.value[groupId];
}
function toggleGroupCollapse(groupId: string) {
    collapsedGroups.value = {
        ...collapsedGroups.value,
        [groupId]: !collapsedGroups.value[groupId],
    };
}

const selectedForGroup = ref<Record<number, boolean>>({});
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
        (it.group_id ?? "").trim() === gid ? { ...it, group_label: label } : it,
    );
}

const groupSelectedRows = () => {
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
};

const ungroupSelectedRows = () => {
    const idxs = selectedIndices();
    if (idxs.length === 0) {
        toast.push("Select items to ungroup.", "error");
        return;
    }
    meal.value.items = meal.value.items.map((it, i) =>
        idxs.includes(i)
            ? { ...it, group_id: "", group_label: "", composite_food_id: null }
            : it,
    );
    selectedForGroup.value = {};
};

const removeGroupLines = (indices: number[]) => {
    const sorted = [...new Set(indices)].sort((a, b) => b - a);
    const items = [...meal.value.items];
    for (const i of sorted) {
        items.splice(i, 1);
    }
    meal.value.items = items;
    selectedForGroup.value = {};
};

const route = useRoute();

const queryType = computed(() => {
    const t = route.query.type;
    return Array.isArray(t) ? t[0] : t;
});
const pageMode = computed(
    (): LogMealPageMode => parseLogMealPageMode(queryType.value),
);
const editVariant = computed(() =>
    pageMode.value === PAGE_MODE.edit
        ? parseEditMealVariant(queryType.value)
        : null,
);
const id = computed(() => Number(route.query.id ?? 0));
const mealLogDayId = computed(() => {
    const d = route.query.dayId;
    const v = Array.isArray(d) ? d[0] : d;
    const n = Number(v ?? 0);
    return Number.isFinite(n) && n > 0 ? n : undefined;
});
const mealId = computed(() => {
    if (pageMode.value !== PAGE_MODE.edit) return null;
    if (editVariant.value === EDIT_VARIANT.saved) return null;
    return id.value !== 0 ? id.value : null;
});
const savedMealEditId = computed(() => {
    if (pageMode.value !== PAGE_MODE.edit) return null;
    if (editVariant.value !== EDIT_VARIANT.saved) return null;
    return id.value !== 0 ? id.value : null;
});
const editMissingId = computed(
    () => pageMode.value === PAGE_MODE.edit && id.value === 0,
);
const pageTitle = computed(() => {
    switch (pageMode.value) {
        case PAGE_MODE.create:
            return "Create New Meal";
        case PAGE_MODE.log:
            return "Log Meal";
        case PAGE_MODE.edit:
            if (editVariant.value === EDIT_VARIANT.saved) {
                return "Edit Saved Meal";
            }
            return editVariant.value === EDIT_VARIANT.planned
                ? "Log Meal"
                : "Edit Logged Meal";
        default:
            return "Log Meal";
    }
});
const { data: mealData, error: mealError } = useMeal(mealId);
const { data: savedMealData, error: savedMealError } =
    useSavedMeal(savedMealEditId);

const mealLoadError = computed(() => {
    if (
        pageMode.value === PAGE_MODE.edit &&
        editVariant.value === EDIT_VARIANT.saved
    ) {
        return savedMealError.value;
    }
    return mealError.value;
});

const showDietDayMacroBars = computed(
    () =>
        pageMode.value !== PAGE_MODE.create &&
        !(
            pageMode.value === PAGE_MODE.edit &&
            editVariant.value === EDIT_VARIANT.saved
        ),
);

// Fetch today's diet logs
const { data: today } = useDietLogsToday();

// start with a default empty meal
const meal = ref<Meal>({
    ID: 0,
    created_at: "",
    updated_at: "",
    name: "",
    items: [],
});

const logMealDraftStorageKey = computed(() => {
    if (
        (pageMode.value !== PAGE_MODE.log &&
            pageMode.value !== PAGE_MODE.create) ||
        id.value !== 0
    ) {
        return "";
    }
    if (pageMode.value === PAGE_MODE.log) {
        const d = mealLogDayId.value;
        return d
            ? `simpletracker:diet:logmeal-draft:v1:log:day:${d}`
            : `simpletracker:diet:logmeal-draft:v1:log`;
    }
    return `simpletracker:diet:logmeal-draft:v1:create`;
});

const logMealDraftSync = useWebStorageJsonSync<LogMealDraftSnapshot>({
    key: logMealDraftStorageKey,
    watchSources: [meal],
    deep: true,
    canPersist: () => logMealDraftStorageKey.value !== "",
    getSnapshot: () => ({
        name: meal.value.name,
        items: JSON.parse(JSON.stringify(toRaw(meal.value.items))) as MealItem[],
    }),
    tryRestore: (parsed, { remove }) => {
        if (typeof parsed.name !== "string") {
            remove();
            return false;
        }
        if (!Array.isArray(parsed.items)) {
            remove();
            return false;
        }
        let items: MealItem[];
        try {
            items = JSON.parse(JSON.stringify(parsed.items)) as MealItem[];
        } catch {
            remove();
            return false;
        }
        if (
            !items.every(
                (it) =>
                    typeof it?.food_id === "number" &&
                    it.food &&
                    typeof it.food === "object",
            )
        ) {
            remove();
            return false;
        }
        meal.value = {
            ...meal.value,
            name: parsed.name,
            items,
        };
        return true;
    },
});

/** Snapshot when opening "edit logged meal"; today totals already include this meal. */
const baselineMealMacros = ref<MealMacroTotals | null>(null);

let logMealDraftHydrateGeneration = 0;

watch(
    [pageMode, id, mealData, savedMealData, queryType, mealLogDayId],
    ([mode, newId, newMealData, newSavedData, qType]) => {
        const variant = parseEditMealVariant(qType);
        logMealDraftSync.setSaveEnabled(false);
        logMealDraftHydrateGeneration++;
        if (
            (mode === PAGE_MODE.log || mode === PAGE_MODE.create) &&
            newId === 0
        ) {
            meal.value = {
                ID: 0,
                created_at: "",
                updated_at: "",
                name: "",
                items: [],
            };
            baselineMealMacros.value = null;
            const hydrateGen = logMealDraftHydrateGeneration;
            void nextTick(() => {
                if (hydrateGen !== logMealDraftHydrateGeneration) return;
                logMealDraftSync.restore();
                logMealDraftSync.setSaveEnabled(true);
            });
            return;
        }
        if (mode === PAGE_MODE.edit && newId !== 0) {
            if (variant === EDIT_VARIANT.saved) {
                if (newSavedData) {
                    meal.value = cloneMeal(savedMealToMeal(newSavedData));
                } else {
                    meal.value = {
                        ID: 0,
                        created_at: "",
                        updated_at: "",
                        name: "",
                        items: [],
                    };
                }
                baselineMealMacros.value = null;
                return;
            }
            if (newMealData?.meal) {
                meal.value = cloneMeal(newMealData.meal);
                baselineMealMacros.value =
                    variant === EDIT_VARIANT.logged
                        ? macrosForMeal(meal.value)
                        : null;
            } else {
                meal.value = {
                    ID: 0,
                    created_at: "",
                    updated_at: "",
                    name: "",
                    items: [],
                };
                baselineMealMacros.value = null;
            }
        }
    },
    { immediate: true },
);

const totalMacros = computed(() => macrosForMeal(meal.value));

/** Day preview for MacroBars: add draft meal to today, except edit-logged replace baseline. */
const macroBarsDayTotals = computed(() => {
    const t = today.value;
    const tm = totalMacros.value;
    const dayCal = t?.totalCalories ?? 0;
    const dayProtein = t?.totalProtein ?? 0;
    const dayFiber = t?.totalFiber ?? 0;
    const dayCarbs = t?.totalCarbs ?? 0;
    const dayFat = t?.totalFat ?? 0;

    if (
        pageMode.value === PAGE_MODE.edit &&
        editVariant.value === EDIT_VARIANT.logged &&
        baselineMealMacros.value
    ) {
        const b = baselineMealMacros.value;
        return {
            totalCalories: dayCal - b.calories + tm.calories,
            totalProtein: dayProtein - b.protein + tm.protein,
            totalFiber: dayFiber - b.fiber + tm.fiber,
            totalCarbs: dayCarbs - b.carbs + tm.carbs,
            totalFat: dayFat - b.fat + tm.fat,
        };
    }
    return {
        totalCalories: dayCal + tm.calories,
        totalProtein: dayProtein + tm.protein,
        totalFiber: dayFiber + tm.fiber,
        totalCarbs: dayCarbs + tm.carbs,
        totalFat: dayFat + tm.fat,
    };
});

const addFood = async (food: Food): Promise<boolean> => {
    const g = "";
    const existingIndex = meal.value.items.findIndex(
        (i) => i?.food?.ID === food.ID && mealItemGroupKey(i) === g,
    );
    if (existingIndex !== -1) {
        meal.value.items = meal.value.items.map((it, i) =>
            i === existingIndex ? { ...it, amount: Number(it.amount) + 1 } : it,
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
};

const addComposite = async (cf: CompositeFood): Promise<boolean> => {
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
};

const pickFoodOrComposite = async (
    row: Food & { entry_kind?: string } & Partial<CompositeFood>,
): Promise<boolean> => {
    if (row.entry_kind === "composite") {
        return addComposite(row as CompositeFood);
    }
    return addFood(row as Food);
};

const createFood = async (name: string): Promise<boolean> => {
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
        } else {
            toast.push("Food Creation Failed!", "error");
            return false;
        }
    } catch (err) {
        toast.push("Dialog Error", "error");
        return false;
    }
};

const removeFood = async (index: number) => {
    meal.value.items = meal.value.items.filter((_, i) => i !== index);
    selectedForGroup.value = {};
};

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

const setMeal = async (item: Meal | SavedMeal): Promise<boolean> => {
    const first = item.items[0];
    meal.value =
        first && "saved_meal_id" in first
            ? savedMealToMeal(item as SavedMeal)
            : cloneMeal(item as Meal);
    return true;
};

// Mutations
const createMealMutation = useCreateMeal();
const createSavedMealMutation = useCreateSavedMeal();
const logEditedMealMutation = useLogEditedMeal();
const updateLoggedMealMutation = useUpdateLoggedMeal();
const updateSavedMealMutation = useUpdateSavedMeal();

const logMealToDay = async (saveToLibrary: boolean) => {
    const mealToCreate = { ...meal.value, ID: 0 };
    try {
        await createMealMutation.mutateAsync({
            meal: mealToCreate,
            log: true,
            saveToLibrary,
        });
        toast.push(
            saveToLibrary ? "Meal logged and saved for later!" : "Meal logged!",
            "success",
        );
        logMealDraftSync.clear();
        meal.value = {
            ID: 0,
            created_at: "",
            updated_at: "",
            name: "",
            items: [],
        };
    } catch (error: any) {
        toast.push("Log meal failed! " + (error.message || ""), "error");
    }
};

const saveSavedMealTemplate = async () => {
    const name = meal.value.name.trim();
    if (!name || meal.value.items.length === 0) {
        toast.push("Add a name and at least one food.", "error");
        return;
    }
    try {
        await createSavedMealMutation.mutateAsync({
            name,
            items: meal.value.items.map((i) => ({
                food_id: i.food_id,
                amount: i.amount,
                group_id: i.group_id ?? "",
                group_label: i.group_label ?? "",
                composite_food_id: i.composite_food_id ?? null,
            })),
        });
        toast.push("Saved meal created!", "success");
        logMealDraftSync.clear();
        meal.value = {
            ID: 0,
            created_at: "",
            updated_at: "",
            name: "",
            items: [],
        };
    } catch (error: any) {
        toast.push("Could not save meal. " + (error?.message || ""), "error");
    }
};

const logEditedMeal = async () => {
    const mealToLog = { ...meal.value, ID: 0 };
    const plannedSourceMealId =
        editVariant.value === EDIT_VARIANT.planned && id.value !== 0
            ? id.value
            : undefined;
    try {
        await logEditedMealMutation.mutateAsync({
            meal: mealToLog,
            plannedSourceMealId,
            dayId: mealLogDayId.value,
        });
        toast.push("Meal Logged Successfully!", "success");
    } catch (error: any) {
        toast.push("Log Edited Failed! " + (error.message || ""), "error");
    }
};

const updateLoggedMeal = async () => {
    const oldMealID = meal.value.ID;
    const mealToUpdate = { ...meal.value, ID: 0 };
    try {
        await updateLoggedMealMutation.mutateAsync({
            meal: mealToUpdate,
            oldMealId: oldMealID,
            dayId: mealLogDayId.value,
        });
        toast.push("Meal Updated Successfully!", "success");
    } catch (error: any) {
        toast.push("Update Failed! " + (error.message || ""), "error");
    }
};

const saveEditedSavedMeal = async () => {
    const name = meal.value.name.trim();
    if (!name || meal.value.items.length === 0) {
        toast.push("Add a name and at least one food.", "error");
        return;
    }
    if (id.value === 0) return;
    try {
        await updateSavedMealMutation.mutateAsync({
            savedMealId: id.value,
            name,
            items: meal.value.items.map((i) => ({
                food_id: i.food_id,
                amount: i.amount,
                group_id: i.group_id ?? "",
                group_label: i.group_label ?? "",
                composite_food_id: i.composite_food_id ?? null,
            })),
        });
        toast.push("Saved meal updated!", "success");
    } catch (error: any) {
        toast.push(
            "Could not update saved meal. " + (error?.message || ""),
            "error",
        );
    }
};
</script>

<template>
    <div class="diet-module flex flex-col items-center pt-4">
        <div
            v-if="editMissingId"
            class="flex h-[60dvh] items-center justify-center p-8 text-cfRed"
        >
            <span>Missing meal id for this edit.</span>
        </div>
        <div
            v-else-if="mealLoadError && id !== 0"
            class="flex h-[60dvh] items-center justify-center p-8 text-cfRed"
        >
            <span
                >Error loading meal:
                {{ mealLoadError?.message || "Unknown error" }}</span
            >
        </div>
        <div
            v-else-if="meal"
            class="flex w-full flex-col gap-4 md:grid md:h-[calc(100dvh-7rem)] md:grid-cols-[2fr_1fr] md:grid-rows-2 pb-16 md:pb-0"
        >
            <article
                class="flex min-h-[min(28rem,55dvh)] flex-col overflow-hidden rounded-lg bg-firstBg md:row-span-2 md:min-h-0"
            >
                <header
                    class="shrink-0 border-b border-secondBg p-4 text-textPrimary"
                >
                    <div
                        class="mb-4 flex flex-row items-center justify-between"
                    >
                        <h1 class="m-0 text-xl font-semibold">
                            {{ pageTitle }}
                        </h1>
                        <SimpleMacros
                            :calories="totalMacros.calories"
                            :protein="totalMacros.protein"
                            :fat="totalMacros.fat"
                            :carbs="totalMacros.carbs"
                            :fiber="totalMacros.fiber"
                        />
                    </div>
                    <div class="flex min-w-0 flex-1 flex-col gap-2">
                        <Input
                            label="Meal Name"
                            type="text"
                            v-model="meal.name"
                            required
                        />
                    </div>
                </header>
                <section
                    class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden text-textPrimary"
                >
                    <div
                        class="flex shrink-0 items-center justify-between gap-3 px-4 py-2.5"
                    >
                        <h3
                            class="m-0 min-w-0 text-base font-medium leading-tight"
                        >
                            Meal Items
                        </h3>
                        <div
                            class="flex shrink-0 flex-wrap items-center justify-end gap-2"
                        >
                            <button
                                type="button"
                                class="rounded border border-secondBg bg-secondBg px-3 py-1.5 text-sm hover:bg-thirdBg"
                                @click="groupSelectedRows"
                            >
                                Group selected
                            </button>
                            <button
                                type="button"
                                class="rounded border border-secondBg bg-secondBg px-3 py-1.5 text-sm hover:bg-thirdBg"
                                @click="ungroupSelectedRows"
                            >
                                Ungroup
                            </button>
                        </div>
                    </div>
                    <div
                        class="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto px-4 pb-1"
                    >
                        <div
                            :class="[
                                mealItemsListGridClass,
                                'mb-1 border-b border-secondBg pb-2 text-xs font-medium text-textSecondary sm:grid',
                            ]"
                        >
                            <span><span class="sr-only">Select</span></span>
                            <span class="min-w-0">Item</span>
                            <span class="min-w-0 text-center">Qty</span>
                            <span class="hidden md:block min-w-0 text-right"
                                >Macros</span
                            >
                            <span
                                class="flex h-9 w-9 shrink-0 justify-self-end"
                                aria-hidden="true"
                            ></span>
                        </div>
                        <template
                            v-for="(block, bi) in mealItemBlocks"
                            :key="'b-' + bi"
                        >
                            <template v-if="block.kind === 'ungrouped'">
                                <LogMealItemRow
                                    v-for="{ item, index: i } in block.rows"
                                    :key="`u-${i}`"
                                    :item="item"
                                    :row-index="i"
                                    :selected="!!selectedForGroup[i]"
                                    @toggle-select="toggleSelectRow"
                                    @amount-plus-minus="amountPlusMinus"
                                    @set-item-amount="setMealItemAmount"
                                    @swap-variant="swapVariant"
                                    @remove="removeFood"
                                />
                            </template>
                            <LogMealGroupBlock
                                v-else
                                :block="block"
                                :expanded="isGroupExpanded(block.groupId)"
                                :selected-for-group="selectedForGroup"
                                @toggle-collapse="toggleGroupCollapse"
                                @set-group-label="setGroupLabel"
                                @remove-group="removeGroupLines"
                                @toggle-select="toggleSelectRow"
                                @amount-plus-minus="amountPlusMinus"
                                @set-item-amount="setMealItemAmount"
                                @swap-variant="swapVariant"
                                @remove-item="removeFood"
                            />
                        </template>
                    </div>
                </section>
                <footer
                    class="flex shrink-0 flex-col gap-4 border-t border-secondBg p-4"
                >
                    <div class="flex w-full flex-row gap-3">
                        <button
                            v-if="pageMode === PAGE_MODE.create"
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="saveSavedMealTemplate"
                        >
                            Save Meal
                        </button>
                        <button
                            v-if="pageMode === PAGE_MODE.log"
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="logMealToDay(false)"
                        >
                            Log Meal
                        </button>
                        <button
                            v-if="pageMode === PAGE_MODE.log"
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="logMealToDay(true)"
                        >
                            Log and Save Meal
                        </button>
                        <button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.saved
                            "
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="saveEditedSavedMeal"
                        >
                            Save changes
                        </button>
                        <button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.logged
                            "
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="updateLoggedMeal"
                        >
                            Update
                        </button>
                        <button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.planned
                            "
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="logEditedMeal"
                        >
                            Log
                        </button>
                    </div>
                    <MacroBars
                        v-if="showDietDayMacroBars"
                        :totalCalories="macroBarsDayTotals.totalCalories"
                        :totalProtein="macroBarsDayTotals.totalProtein"
                        :totalFat="macroBarsDayTotals.totalFat"
                        :totalCarbs="macroBarsDayTotals.totalCarbs"
                        :planned-calories="today?.day.plan.calories ?? 0"
                        :planned-protein="today?.day.plan.protein ?? 0"
                        :planned-fat="today?.day.plan.fat ?? 0"
                        :planned-carbs="today?.day.plan.carbs ?? 0"
                    />
                </footer>
            </article>
            <aside
                class="flex min-h-[min(20rem,45dvh)] flex-col overflow-hidden rounded-lg bg-firstBg p-4 text-textPrimary md:min-h-0"
            >
                <h2 class="mt-0 text-lg font-semibold">Add Foods</h2>
                <SearchList
                    :route="'diet/meals/food/all'"
                    :onSelect="pickFoodOrComposite"
                    :onCreate="createFood"
                    :displayComponent="FoodDisplay"
                />
            </aside>
            <aside
                class="flex min-h-[min(20rem,45dvh)] flex-col overflow-hidden rounded-lg bg-firstBg p-4 text-textPrimary md:min-h-0"
            >
                <h2 class="mt-0 text-lg font-semibold">Select Saved Meal</h2>
                <SearchList
                    :key="meal.ID"
                    :route="'diet/meals/saved-meal/all'"
                    :on-select="setMeal"
                />
            </aside>
        </div>
    </div>
</template>
