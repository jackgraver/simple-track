<script setup lang="ts">
import { useRoute } from "vue-router";
import type { Food, Meal, MealItem, SavedMeal } from "~/types/diet";
import SearchList from "~/shared/SearchList.vue";
import { Plus, Trash2, Minus } from "lucide-vue-next";
import FoodDisplay from "~/shared/FoodDisplay.vue";
import Input from "~/shared/input/Input.vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import CreateFoodDialog from "./dialog/CreateFoodDialog.vue";
import { computed, ref, toRaw, watch } from "vue";
import { useMeal } from "./queries/useMeal";
import { useDietLogsToday } from "./queries/useDietLogsToday";
import {
    useCreateMeal,
    useCreateSavedMeal,
    useLogEditedMeal,
    useUpdateLoggedMeal,
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

type MealMacroTotals = {
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
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
    };
}

function formatNum(n: number): number {
    const s = n.toFixed(2); // always 2 decimals
    return Number(s.replace(/\.?0+$/, "")); // drop trailing zeros and optional dot
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
        return;
    }
    meal.value.items = items.map((it, i) =>
        i === index ? { ...it, amount: next } : it,
    );
}

function itemServingAmount(item: MealItem): number {
    return (item.food?.serving_amount || 1) * item.amount;
}

const route = useRoute();
const backTo = computed(() =>
    route.name === "diet-log"
        ? { name: "diet" as const }
        : { name: "home" as const },
);
const backLabel = computed(() =>
    route.name === "diet-log" ? "← Diet" : "← Home",
);
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
const mealId = computed(() => {
    if (pageMode.value !== PAGE_MODE.edit) return null;
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
            return editVariant.value === EDIT_VARIANT.planned
                ? "Log Meal"
                : "Edit Logged Meal";
        default:
            return "Log Meal";
    }
});
const { data: mealData, error: mealError } = useMeal(mealId);

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

/** Snapshot when opening "edit logged meal"; today totals already include this meal. */
const baselineMealMacros = ref<MealMacroTotals | null>(null);

watch(
    [pageMode, id, mealData, queryType],
    ([mode, newId, newMealData, qType]) => {
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
            return;
        }
        if (newMealData?.meal) {
            meal.value = cloneMeal(newMealData.meal);
            if (
                mode === PAGE_MODE.edit &&
                parseEditMealVariant(qType) === EDIT_VARIANT.logged
            ) {
                baselineMealMacros.value = macrosForMeal(meal.value);
            } else {
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
        };
    }
    return {
        totalCalories: dayCal + tm.calories,
        totalProtein: dayProtein + tm.protein,
        totalFiber: dayFiber + tm.fiber,
        totalCarbs: dayCarbs + tm.carbs,
    };
});

const addFood = async (food: Food): Promise<boolean> => {
    const existingIndex = meal.value.items.findIndex(
        (i) => i?.food?.ID === food.ID,
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
            meal_id: meal.value.ID,
            food_id: food.ID,
            food: food,
            amount: 1,
        } as MealItem,
    ];
    return true;
};

const createFood = async (name: string): Promise<boolean> => {
    try {
        const food = await dialogManager.custom<Food>({
            title: "Create " + name,
            component: CreateFoodDialog,
            props: { foodName: name },
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
        console.error("Dialog error:", err);
        toast.push("Dialog Error", "error");
        return false;
    }
};

const removeFood = async (index: number) => {
    meal.value.items = meal.value.items.filter((_, i) => i !== index);
};

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
            })),
        });
        toast.push("Saved meal created!", "success");
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
    try {
        await logEditedMealMutation.mutateAsync(mealToLog);
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
        });
        toast.push("Meal Updated Successfully!", "success");
    } catch (error: any) {
        toast.push("Update Failed! " + (error.message || ""), "error");
    }
};
</script>

<template>
    <div class="flex flex-col items-center py-4">
        <router-link
            :to="backTo"
            class="mb-2 w-[80%] max-w-[80%] text-left text-sm text-textSecondary hover:text-textPrimary"
            >{{ backLabel }}</router-link
        >
        <div
            v-if="editMissingId"
            class="flex h-[60dvh] items-center justify-center p-8 text-cfRed"
        >
            <span>Missing meal id for this edit.</span>
        </div>
        <div
            v-else-if="mealError && id !== 0"
            class="flex h-[60dvh] items-center justify-center p-8 text-cfRed"
        >
            <span
                >Error loading meal:
                {{ mealError?.message || "Unknown error" }}</span
            >
        </div>
        <div
            v-else-if="meal"
            class="grid h-[calc(100dvh-7rem)] w-full grid-cols-[2fr_1fr] grid-rows-2 gap-4 pb-8"
        >
            <article
                class="row-span-2 flex min-h-0 flex-col overflow-hidden rounded-lg bg-firstBg"
            >
                <header
                    class="shrink-0 border-b border-secondBg p-4 text-textPrimary"
                >
                    <div class="mb-4">
                        <h1 class="m-0 text-xl font-semibold">
                            {{ pageTitle }}
                        </h1>
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
                    class="flex min-h-0 flex-1 flex-col overflow-hidden text-textPrimary"
                >
                    <h3
                        class="m-0 shrink-0 px-4 pb-3 pt-4 text-base font-medium"
                    >
                        Meal Items
                    </h3>
                    <div
                        class="flex min-h-0 flex-1 flex-col overflow-y-auto px-4 pb-1"
                    >
                        <div
                            v-if="meal.items.length"
                            class="mb-1 hidden grid-cols-[minmax(0,1fr)_9rem_11rem_2.5rem] gap-x-3 border-b border-secondBg pb-2 text-xs font-medium text-textSecondary sm:grid"
                        >
                            <span>Item</span>
                            <span class="text-center">Qty</span>
                            <span class="text-right">Macros</span>
                            <span
                                class="w-9 shrink-0"
                                aria-hidden="true"
                            ></span>
                        </div>
                        <div
                            v-for="(item, i) in meal.items"
                            :key="`${item.food_id}-${i}`"
                            class="grid grid-cols-[minmax(0,1fr)_9rem_11rem_2.5rem] items-center gap-x-3 gap-y-1 border-b border-secondBg py-2.5 last:border-b-0"
                        >
                            <span
                                class="min-w-0 truncate text-base font-medium text-textPrimary"
                                :title="item.food?.name"
                                >{{ item.food?.name ?? "" }}</span
                            >
                            <div
                                class="flex items-center justify-center gap-1 tabular-nums"
                            >
                                <button
                                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                                    type="button"
                                    @click="amountPlusMinus(i, 'minus')"
                                >
                                    <Minus :size="18" />
                                </button>
                                <span
                                    class="min-w-11 shrink-0 text-center text-sm text-textPrimary"
                                    >{{ formatNum(itemServingAmount(item))
                                    }}<span
                                        v-if="item.food?.serving_type === 'g'"
                                        class="text-textSecondary"
                                        >g</span
                                    ></span
                                >
                                <button
                                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                                    type="button"
                                    @click="amountPlusMinus(i, 'plus')"
                                >
                                    <Plus :size="18" />
                                </button>
                            </div>
                            <span
                                class="min-w-0 text-right text-sm tabular-nums text-textSecondary"
                            >
                                {{
                                    formatNum(
                                        item.amount * item.food!.calories,
                                    )
                                }}C /
                                {{
                                    formatNum(item.amount * item.food!.protein)
                                }}P /
                                {{ formatNum(item.amount * item.food!.fiber) }}F
                            </span>
                            <button
                                class="flex h-9 w-9 shrink-0 items-center justify-center justify-self-end rounded text-textSecondary transition-colors hover:bg-secondBg hover:text-cfRed"
                                type="button"
                                aria-label="Remove item"
                                @click="removeFood(i)"
                            >
                                <Trash2 :size="20" />
                            </button>
                        </div>
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
                        v-if="pageMode !== PAGE_MODE.create"
                        :totalCalories="macroBarsDayTotals.totalCalories"
                        :totalProtein="macroBarsDayTotals.totalProtein"
                        :totalFiber="macroBarsDayTotals.totalFiber"
                        :totalCarbs="macroBarsDayTotals.totalCarbs"
                        :planned-calories="today?.day.plan.calories ?? 0"
                        :planned-protein="today?.day.plan.protein ?? 0"
                        :planned-fiber="today?.day.plan.fiber ?? 0"
                        :planned-carbs="today?.day.plan.carbs ?? 0"
                    />
                </footer>
            </article>
            <aside
                class="flex min-h-0 flex-col overflow-hidden rounded-lg bg-firstBg p-4 text-textPrimary"
            >
                <h2 class="mt-0 text-lg font-semibold">Add Foods</h2>
                <SearchList
                    :route="'diet/meals/food/all'"
                    :onSelect="addFood"
                    :onCreate="createFood"
                    :displayComponent="FoodDisplay"
                />
            </aside>
            <aside
                class="flex min-h-0 flex-col overflow-hidden rounded-lg bg-firstBg p-4 text-textPrimary"
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
