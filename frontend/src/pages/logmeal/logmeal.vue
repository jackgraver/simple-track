<script setup lang="ts">
import { useRoute } from "vue-router";
import type { Food, Meal, MealItem } from "~/types/diet";
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
    useLogEditedMeal, 
    useUpdateLoggedMeal 
} from "./queries/useMealMutations";
import MacroBars from "~/shared/MacroBars.vue";
import SimpleMacros from "~/shared/SimpleMacros.vue";

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
const type = route.query.type as string | undefined;
const id = computed(() => Number(route.query.id ?? 0));

// Fetch meal if ID is provided - pass computed ref for reactivity
const mealId = computed(() => id.value !== 0 ? id.value : null);
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

// Watch for meal data and update meal ref
watch(mealData, (newMealData) => {
    if (newMealData?.meal) {
        meal.value = newMealData.meal;
    }
}, { immediate: true });

// Reset meal when ID becomes 0 (creating new meal) or when no meal data
watch([id, mealData], ([newId, newMealData]) => {
    if (newId === 0 && !newMealData?.meal) {
        meal.value = {
            ID: 0,
            created_at: "",
            updated_at: "",
            name: "",
            items: [],
        };
    }
}, { immediate: true });

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

const setMeal = async (newMeal: Meal): Promise<boolean> => {
    meal.value = newMeal;
    return true;
};

// Mutations
const createMealMutation = useCreateMeal();
const logEditedMealMutation = useLogEditedMeal();
const updateLoggedMealMutation = useUpdateLoggedMeal();

const createMeal = async (log: boolean) => {
    const mealToCreate = { ...meal.value, ID: 0 };
    try {
        await createMealMutation.mutateAsync({ meal: mealToCreate, log });
        toast.push("Meal Created Successfully!", "success");
        if (!log) {
            // Reset meal if not logging
            meal.value = {
                ID: 0,
                created_at: "",
                updated_at: "",
                name: "",
                items: [],
            };
        }
    } catch (error: any) {
        toast.push("Create Meal Failed! " + (error.message || ""), "error");
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
            oldMealId: oldMealID 
        });
        toast.push("Meal Updated Successfully!", "success");
    } catch (error: any) {
        toast.push("Update Failed! " + (error.message || ""), "error");
    }
};
</script>

<template>
    <div class="page-wrapper">
        <div v-if="mealError && id !== 0" class="error-container">
            <span>Error loading meal: {{ mealError?.message || 'Unknown error' }}</span>
        </div>
        <div v-else-if="meal" class="container">
            <article class="cell left">
                <header class="title-bar">
                    <div class="title-row">
                        <h1 v-if="type === 'edit'">Log Meal</h1>
                        <h1 v-if="type === 'editlogged'">Log Meal</h1>
                        <h1 v-if="type === 'create'">Create Meal</h1>
                    </div>
                    <div class="meal-header">
                        <div class="meal-name">
                            <label for="name">Meal Name</label>
                            <input type="text" id="name" v-model="meal.name" />
                        </div>
                        <SimpleMacros
                            :calories="totalMacros.calories"
                            :protein="totalMacros.protein"
                            :fiber="totalMacros.fiber"
                            font-size="1.3rem"
                        />
                    </div>
                </header>
                <section class="meal-items-section">
                    <h3>Meal Items</h3>
                    <div class="meal-items">
                        <div
                            v-for="(item, i) in meal.items"
                            :key="item.ID"
                            class="items-rows"
                        >
                            <button @click="amountPlusMinus(item, 'minus')">
                                <Minus />
                            </button>
                            <span style="font-size: large"
                                >{{ formatFoodLabel(item) }}
                            </span>
                            <span v-if="item.food!.serving_type === 'g'">g</span
                            ><button @click="amountPlusMinus(item, 'plus')">
                                <Plus />
                            </button>
                            <span>
                                {{ formatNum(item.amount * item.food!.calories) }}C
                                / {{ formatNum(item.amount * item.food!.protein) }}P
                                / {{ formatNum(item.amount * item.food!.fiber) }}F
                            </span>
                            <button class="delete-button" @click="removeFood(i)">
                                <Trash2 :size="20" />
                            </button>
                        </div>
                    </div>
                </section>
                <footer class="footer">
                    <div class="action-buttons">
                        <button v-if="type === 'edit'" @click="updateLoggedMeal">
                            Update
                        </button>
                        <button v-if="type === 'editlogged'" @click="logEditedMeal">
                            Log
                        </button>
                        <button v-if="type === 'create'" @click="createMeal(false)">
                            Create
                        </button>
                        <button v-if="type === 'create'" @click="createMeal(true)">
                            Create and Log
                        </button>
                    </div>
                    <MacroBars
                        :totalCalories="
                            totalMacros.calories + (today?.totalCalories ?? 0)
                        "
                        :totalProtein="
                            totalMacros.protein + (today?.totalProtein ?? 0)
                        "
                        :totalFiber="totalMacros.fiber + (today?.totalFiber ?? 0)"
                        :totalCarbs="totalMacros.carbs + (today?.totalCarbs ?? 0)"
                        :planned-calories="today?.day.plan.calories ?? 0"
                        :planned-protein="today?.day.plan.protein ?? 0"
                        :planned-fiber="today?.day.plan.fiber ?? 0"
                        :planned-carbs="today?.day.plan.carbs ?? 0"
                    />
                </footer>
            </article>
            <aside class="cell right-top">
                <h2>Add Foods</h2>
                <SearchList
                    :route="'diet/meals/food/all'"
                    :onSelect="addFood"
                    :onCreate="createFood"
                    :displayComponent="FoodDisplay"
                />
            </aside>
            <aside class="cell right-bottom">
                <h2>Select Saved Meal</h2>
                <SearchList
                    :key="meal.ID"
                    :route="'diet/meals/meal/all'"
                    :on-select="setMeal"
                />
            </aside>
        </div>
    </div>
</template>

<style scoped>
.page-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 98vh;
    padding-block: 2rem;
    box-sizing: border-box;
    overflow: hidden;
}

.container {
    display: grid;
    grid-template-columns: 2fr 1fr;
    grid-template-rows: 1fr 1fr;
    gap: 1rem;
    width: 80%;
    height: 100%;
    max-height: 100%;
}


.cell {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    min-height: 0;
    background: rgb(27, 27, 27);
}

.cell > h1,
h2 {
    margin-top: 0;
}

.left {
    grid-row: 1 / span 2;
    padding: 0;
    display: flex;
    flex-direction: column;
    min-height: 0;
}

.right-top,
.right-bottom {
    padding: 1rem;
}

header.title-bar {
    padding: 1rem;
    border-bottom: 1px solid rgb(56, 56, 56);
    flex-shrink: 0;
}

.title-row h1 {
    margin: 0 0 1rem 0;
}

.meal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 1.5rem;
}

.meal-name {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex: 1;
}

.meal-name label {
    font-size: 0.9rem;
    color: #ccc;
}

.meal-name input {
    background-color: rgb(50, 50, 50);
    color: white;
    border: 1px solid #3d3d3d;
    border-radius: 4px;
    padding: 0.5rem;
    font-size: 1rem;
}

section.meal-items-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
}

.meal-items-section h3 {
    margin: 0;
    padding: 1rem 1rem 0.75rem 1rem;
    flex-shrink: 0;
}

.meal-items {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 0 1rem;
    overflow-y: auto;
    flex: 1;
}

.items-rows {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    gap: 0.5rem;
}

footer.footer {
    padding: 1rem;
    border-top: 1px solid rgb(56, 56, 56);
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.action-buttons {
    display: flex;
    flex-direction: row;
    gap: 0.75rem;
    width: 100%;
}

.action-buttons button {
    flex: 1;
    padding: 0.625rem 1.25rem;
    cursor: pointer;
    font-size: 0.9rem;
}

.small-input {
    width: 16px;
    text-align: center;
}

.loading-container,
.error-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
    padding: 2rem;
    color: #ccc;
}

.error-container {
    color: #ff6b6b;
}
</style>
