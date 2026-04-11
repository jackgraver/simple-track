<script setup lang="ts">
import { useRoute } from "vue-router";
import type { Food, Meal, MealItem, SavedMeal } from "~/types/diet";
import SearchList from "~/shared/SearchList.vue";
import { Plus, Trash2, Minus } from "lucide-vue-next";
import FoodDisplay from "~/shared/FoodDisplay.vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import CreateFoodDialog from "./dialog/CreateFoodDialog.vue";
import { computed, ref, watch } from "vue";
import { useMeal } from "./queries/useMeal";
import { useDietLogsToday } from "./queries/useDietLogsToday";
import {
    useCreateMeal,
    useCreateSavedMeal,
    useLogEditedMeal,
    useUpdateLoggedMeal,
} from "./queries/useMealMutations";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { savedMealToMeal } from "~/utils/savedMealToMeal";
import {
    EDIT_VARIANT,
    PAGE_MODE,
    parseEditMealVariant,
    parseLogMealPageMode,
    type LogMealPageMode,
} from "./logmealMode";

function formatNum(n: number): number {
    const s = n.toFixed(2); // always 2 decimals
    return Number(s.replace(/\.?0+$/, "")); // drop trailing zeros and optional dot
}

function amountPlusMinus(item: MealItem, direction: "plus" | "minus") {
    if (direction === "plus") {
        item.amount += 1 / (item.food?.serving_amount || 1);
    } else {
        item.amount = Math.max(
            item.amount - 1 / (item.food?.serving_amount || 1),
            0,
        );
    }
}

function formatFoodLabel(item: any): string {
    const servingAmount = (item.food?.serving_amount || 1) * item.amount;
    const type = item.food?.serving_type === "g" ? "g" : "";
    const name = item.food?.name ?? "";

    return `(${servingAmount}${type}) ${name}`;
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

watch(
    [pageMode, id, mealData],
    ([mode, newId, newMealData]) => {
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
            return;
        }
        if (newMealData?.meal) {
            meal.value = newMealData.meal;
        }
    },
    { immediate: true },
);

const totalMacros = computed(() => {
    return {
        calories: formatNum(
            meal.value.items.reduce(
                (total, item) => total + item.amount * item.food!.calories,
                0,
            ),
        ),
        protein: formatNum(
            meal.value.items.reduce(
                (total, item) => total + item.amount * item.food!.protein,
                0,
            ),
        ),
        fiber: formatNum(
            meal.value.items.reduce(
                (total, item) => total + item.amount * item.food!.fiber,
                0,
            ),
        ),
        carbs: formatNum(
            meal.value.items.reduce(
                (total, item) => total + item.amount * item.food!.carbs,
                0,
            ),
        ),
    };
});

const addFood = async (food: Food): Promise<boolean> => {
    const existing = meal.value.items.find((i) => i?.food?.ID === food.ID);
    if (existing) {
        existing.amount++;
        return true;
    }

    meal.value.items.push({
        meal_id: meal.value.ID,
        food_id: food.ID,
        food: food,
        amount: 1,
    } as MealItem);
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
    meal.value.items.splice(index, 1);
};

const setMeal = async (item: Meal | SavedMeal): Promise<boolean> => {
    const first = item.items[0];
    meal.value =
        first && "saved_meal_id" in first
            ? savedMealToMeal(item as SavedMeal)
            : (item as Meal);
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
                    <div class="flex items-end justify-between gap-6">
                        <div class="flex min-w-0 flex-1 flex-col gap-2">
                            <label class="text-sm text-textSecondary" for="name"
                                >Meal Name</label
                            >
                            <input
                                class="w-full rounded bg-secondBg px-2 py-2 text-base text-textPrimary focus:outline-none focus:ring-1 focus:ring-thirdBg"
                                type="text"
                                id="name"
                                v-model="meal.name"
                            />
                        </div>
                        <SimpleMacros
                            :calories="totalMacros.calories"
                            :protein="totalMacros.protein"
                            :fiber="totalMacros.fiber"
                            font-size="1.3rem"
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
                        class="flex min-h-0 flex-1 flex-col gap-2.5 overflow-y-auto px-4"
                    >
                        <div
                            v-for="(item, i) in meal.items"
                            :key="item.ID"
                            class="flex w-full items-center justify-between gap-2"
                        >
                            <button
                                class="shrink-0 rounded bg-secondBg p-1.5 text-textPrimary hover:bg-thirdBg"
                                type="button"
                                @click="amountPlusMinus(item, 'minus')"
                            >
                                <Minus />
                            </button>
                            <span
                                class="min-w-0 flex-1 text-lg text-textPrimary"
                                >{{ formatFoodLabel(item) }}</span
                            >
                            <span
                                v-if="item.food!.serving_type === 'g'"
                                class="shrink-0 text-textSecondary"
                                >g</span
                            >
                            <button
                                class="shrink-0 rounded bg-secondBg p-1.5 text-textPrimary hover:bg-thirdBg"
                                type="button"
                                @click="amountPlusMinus(item, 'plus')"
                            >
                                <Plus />
                            </button>
                            <span class="shrink-0 text-sm text-textSecondary">
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
                                class="shrink-0 rounded p-1 text-textSecondary hover:bg-secondBg hover:text-cfRed"
                                type="button"
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
                        :totalCalories="
                            totalMacros.calories + (today?.totalCalories ?? 0)
                        "
                        :totalProtein="
                            totalMacros.protein + (today?.totalProtein ?? 0)
                        "
                        :totalFiber="
                            totalMacros.fiber + (today?.totalFiber ?? 0)
                        "
                        :totalCarbs="
                            totalMacros.carbs + (today?.totalCarbs ?? 0)
                        "
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
