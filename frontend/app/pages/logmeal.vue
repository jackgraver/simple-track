<script setup lang="ts">
import { useRoute } from "vue-router";
import type { Day, Food, Meal, MealItem } from "~/types/diet";
import SearchList from "~/components/SearchList.vue";
import { Plus, Trash2, Minus } from "lucide-vue-next";
import FoodDisplay from "~/components/FoodDisplay.vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import CreateFoodDialog from "~/components/CreateFoodDialog.vue";

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

const router = useRouter();

const route = useRoute();
const type = route.query.type as string | undefined;
const id = Number(route.query.id ?? 0);

// start with a default empty meal
const meal = ref<Meal>({
    ID: 0,
    created_at: "",
    updated_at: "",
    name: "",
    items: [],
});

if (id !== 0) {
    const { data, pending, error } = await useAPIGet<{ meal: Meal }>(
        `mealplan/meal/${id}`,
    );
    if (data.value) {
        meal.value = data.value.meal; // replace default with fetched
    } else {
        console.warn("No meal found for ID:", id);
    }
}

const {
    data: today,
    pending,
    error,
} = useAPIGet<{
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
}>(`mealplan/today`);

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

//TODO: potential bug this gets called every character input to meal name
const createMeal = async (log: boolean) => {
    console.log(meal.value);
    meal.value.ID = 0;
    const { response, error } = await useAPIPost<{
        meal_id: number;
    }>(`mealplan/meal/new`, "POST", {
        meal: meal.value,
        log: log,
    });
    if (error) {
        toast.push("Create Meal Failed!" + error.message, "error");
    } else if (response) {
        toast.push("Meal Created Successfully!", "success");
        if (log) {
            router.push("/");
        } else {
            meal.value = {
                ID: 0,
                created_at: "",
                updated_at: "",
                name: "",
                items: [],
            };
        }
    }
};

const logEditedMeal = async () => {
    meal.value.ID = 0;
    const { response, error } = await useAPIPost<{
        day: Day;
        totalCalories: number;
        totalProtein: number;
        totalFiber: number;
    }>(`mealplan/meal/logedited`, "POST", {
        meal: meal.value,
    });

    if (error) {
        toast.push("Log Edited Failed!" + error.message, "error");
    } else if (response) {
        console.log(response);
        if (response.day) {
            toast.push("Planned Meal Log Successfully!", "success");
            router.push("/");
        }
    }
};

const updateLoggedMeal = async () => {
    const oldMealID = meal.value.ID;
    meal.value.ID = 0;
    const { response, error } = await useAPIPost<{
        day: Day;
        totalCalories: number;
        totalProtein: number;
        totalFiber: number;
    }>(`mealplan/meal/editlogged`, "POST", {
        meal: meal.value,
        oldMealID: oldMealID,
    });
    if (error) {
        toast.push("Log Edited Failed!" + error.message, "error");
    } else if (response) {
        if (response.day) {
            toast.push("Planned Meal Log Successfully!", "success");
            router.push("/");
        }
    }
};
</script>

<template>
    <div class="page-wrapper">
        <div v-if="meal" class="container">
            <div class="cell left">
                <div class="title-row">
                    <h1 v-if="type === 'edit'">Log Meal</h1>
                    <h1 v-if="type === 'editlogged'">Log Meal</h1>
                    <h1 v-if="type === 'create'">Create Meal</h1>
                </div>
                <div class="meal-info-row">
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
                <!-- TODO: Removed 0'd items on submit  -->
                <TodayBars
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
            </div>
            <div class="cell right-top">
                <h2>Add Foods</h2>
                <SearchList
                    :route="'mealplan/food/all'"
                    :onSelect="addFood"
                    :onCreate="createFood"
                    :displayComponent="FoodDisplay"
                />
            </div>
            <div class="cell right-bottom">
                <h2>Select Saved Meal</h2>
                <SearchList
                    :key="meal.ID"
                    :route="'mealplan/meal/all'"
                    :on-select="setMeal"
                />
            </div>
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

.title-row h1 {
    margin: 0 0 1.5rem 0;
}

.meal-info-row {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    align-items: center;
    padding-bottom: 0.5rem;
}

.meal-name {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.meal-macros {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
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
    padding: 1rem;
}

.right-top,
.right-bottom {
    padding: 1rem;
}

.meal-items {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
}

.items-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.small-input {
    width: 16px;
    text-align: center;
}
</style>
