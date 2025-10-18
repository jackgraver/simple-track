<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";
import TodayLogEditedDialog from "~/components/today/LogEditedDialog.vue";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";

function formatNum(n: number): string {
    const s = n.toFixed(2); // always 2 decimals
    return s.replace(/\.?0+$/, ""); // drop trailing zeros and optional dot
}

function mealMacros(meal: Meal): string {
    let calories = 0;
    let protein = 0;
    let fiber = 0;

    for (const item of meal.items) {
        calories += (item.food?.calories ?? 0) * item.amount;
        protein += (item.food?.protein ?? 0) * item.amount;
        fiber += (item.food?.fiber ?? 0) * item.amount;
    }

    return `${formatNum(calories)} / ${formatNum(protein)} / ${formatNum(
        fiber,
    )}`;
}

const { data, pending, error } = useAPIGet<{
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
}>(`mealplan/today`);

const logPlannedMeal = async (meal: Meal) => {
    const { response, error } = await useAPIPost<{
        day: Day;
        totalCalories: number;
        totalProtein: number;
        totalFiber: number;
    }>(`mealplan/meal/log-planned`, "POST", {
        meal_id: meal.ID,
    });

    if (error) {
        toast.push("Planned Meal Log Failed!", "error");
    } else if (response) {
        toast.push("Planned Meal Log Successfully!", "success");
        if (data.value) {
            data.value = {
                day: response.day,
                totalCalories: response.totalCalories,
                totalProtein: response.totalProtein,
                totalFiber: response.totalFiber,
            };
        }
    }
};

const logEditedMeal = async (meal: Meal) => {
    dialogManager
        .custom<Meal>({
            title: "Log Edited Meal",
            component: TodayLogEditedDialog,
            props: { meal },
        })
        .then(async (editedMeal) => {
            const { response, error } = await useAPIPost<{
                day: Day;
                totalCalories: number;
                totalProtein: number;
                totalFiber: number;
            }>("mealplan/meal/logedited", "POST", { meal: editedMeal });

            if (error)
                toast.push("Log Edited Failed!" + error.message, "error");
            else if (response) {
                toast.push("Planned Meal Log Successfully!", "success");
                if (data.value) {
                    data.value = {
                        day: response.day,
                        totalCalories: response.totalCalories,
                        totalProtein: response.totalProtein,
                        totalFiber: response.totalFiber,
                    };
                }
            }
        })
        .catch((err) => {
            console.error("Dialog error:", err);
            toast.push("Dialog Error", "error");
        });
};

const deleteLoggedMeal = async (meal: Meal) => {
    console.log(meal);
    const { response, error } = await useAPIPost<{
        day: Day;
        totalCalories: number;
        totalProtein: number;
        totalFiber: number;
    }>(`mealplan/meal/logged`, "DELETE", {
        meal_id: meal.ID,
        day_id: data.value?.day.ID,
    });

    if (error) {
        toast.push("Delete Failed!", "error");
    } else if (response) {
        toast.push("Delete Successfully!", "success");
        if (data.value) {
            data.value = {
                day: response.day,
                totalCalories: response.totalCalories,
                totalProtein: response.totalProtein,
                totalFiber: response.totalFiber,
            };
        }
    }
};

const editLogMeal = (meal: Meal) => {
    const oldMealID = meal.ID;
    dialogManager
        .custom<Meal>({
            title: "Log Edited Meal",
            component: TodayLogEditedDialog,
            props: { meal },
        })
        .then(async (editedMeal) => {
            console.log(editedMeal);
            const { response, error } = await useAPIPost<{
                day: Day;
                totalCalories: number;
                totalProtein: number;
                totalFiber: number;
            }>("mealplan/meal/editlogged", "POST", {
                meal: editedMeal,
                oldMealID: oldMealID,
            });

            if (error)
                toast.push("Log Edited Failed!" + error.message, "error");
            else if (response) {
                toast.push("Planned Meal Log Successfully!", "success");
                if (data.value) {
                    data.value = {
                        day: response.day,
                        totalCalories: response.totalCalories,
                        totalProtein: response.totalProtein,
                        totalFiber: response.totalFiber,
                    };
                }
            }
        })
        .catch((err) => {
            console.error("Dialog error:", err);
            toast.push("Dialog Error", "error");
        });
};
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div v-if="data">
            <div class="title-row">
                <h1 style="flex: 1">
                    {{
                        formatDate(data.day.date) +
                        ", " +
                        dayOfWeek(data.day.date)
                    }}
                </h1>
                <NuxtLink class="link" to="/mealplan"
                    >Manage Meal Plan</NuxtLink
                >
            </div>
            <div class="bars-container">
                <div v-if="data" class="fill-container">
                    <div
                        class="fill calories"
                        :style="{
                            width: `${Math.min(100, (data?.totalCalories / data?.day.plan.calories) * 100)}%`,
                        }"
                    >
                        <span>{{
                            formatNum(data?.totalCalories ?? 0) +
                            " / " +
                            formatNum(data?.day.plan.calories ?? 0)
                        }}</span>
                    </div>
                </div>
                <div v-if="data" class="fill-container">
                    <div
                        class="fill protein"
                        :style="{
                            width: `${Math.min(100, (data?.totalProtein / data?.day.plan.protein) * 100)}%`,
                        }"
                    >
                        <span>{{
                            formatNum(data?.totalProtein ?? 0) +
                            " / " +
                            formatNum(data?.day.plan.protein ?? 0)
                        }}</span>
                    </div>
                </div>
                <div v-if="data" class="fill-container">
                    <div
                        class="fill fiber"
                        :style="{
                            width: `${Math.min(100, (data?.totalFiber / data?.day.plan.fiber) * 100)}%`,
                        }"
                    >
                        <span>{{
                            formatNum(data?.totalFiber ?? 0) +
                            " / " +
                            formatNum(data?.day.plan.fiber ?? 0)
                        }}</span>
                    </div>
                </div>
            </div>
            <div class="meals-section">
                <div class="meals-container">
                    <div class="title-row">
                        <h2>Logged</h2>
                        <button>Log Other</button>
                    </div>
                    <div
                        v-for="meal in data.day.loggedMeals"
                        :key="meal.ID"
                        class="meal"
                    >
                        <div class="expected-header">
                            <h3>
                                {{ meal.meal.name }} {{ mealMacros(meal.meal) }}
                            </h3>
                            <button
                                class="delete-button"
                                @click="deleteLoggedMeal(meal.meal)"
                            >
                                Delete
                            </button>
                            <button @click="editLogMeal(meal.meal)">
                                Edit
                            </button>
                        </div>
                        <span
                            v-for="food in meal.meal.items"
                            :key="food.ID"
                            class="food"
                        >
                            {{ food.amount }} - {{ food.food?.name }}
                            <span class="details">
                                <span class="cal"
                                    >{{
                                        food.food?.calories ?? 0 * food.amount
                                    }}C</span
                                >
                                /
                                <span class="pro"
                                    >{{
                                        food.food?.protein ?? 0 * food.amount
                                    }}P</span
                                >
                                /
                                <span class="fib"
                                    >{{
                                        food.food?.fiber ?? 0 * food.amount
                                    }}F</span
                                >
                            </span>
                        </span>
                    </div>
                </div>
                <div class="meals-container">
                    <h2>Planned</h2>
                    <div
                        v-for="meal in data.day.plannedMeals"
                        :key="meal.ID"
                        class="meal"
                    >
                        <div class="expected-header">
                            <h3>
                                {{ meal.meal.name }} {{ mealMacros(meal.meal) }}
                            </h3>
                            <button @click="logEditedMeal(meal.meal)">
                                Log Edited
                            </button>
                            <button
                                class="confirm-button"
                                @click="logPlannedMeal(meal.meal)"
                            >
                                Log
                            </button>
                        </div>
                        <span
                            v-for="food in meal.meal.items"
                            :key="food.ID"
                            class="food"
                        >
                            {{ food.amount }} - {{ food.food?.name }}
                            <span class="details">
                                <span class="cal"
                                    >{{
                                        food.food?.calories ?? 0 * food.amount
                                    }}C</span
                                >
                                /
                                <span class="pro"
                                    >{{
                                        food.food?.protein ?? 0 * food.amount
                                    }}P</span
                                >
                                /
                                <span class="fib"
                                    >{{
                                        food.food?.fiber ?? 0 * food.amount
                                    }}F</span
                                >
                            </span>
                        </span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    width: 75%;
}

.title-row {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    align-items: center;
}

.link {
    color: rgb(177, 177, 177);
    text-decoration: none;
    background-color: rgb(56, 56, 56);
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-weight: bold;
}

.link:hover {
    color: rgb(199, 199, 199);
    transition: color 0.3s;
    background-color: rgb(82, 82, 82);
}

.bars-container {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    width: 100%;
}

.fill-container {
    flex: 1;
    height: 20px;
    width: 250px;
    border: #ffffff;
    border-radius: 4px;
    background-color: #252525;
    border-color: #8d8d8d;
}

.fill {
    height: 100%;
    color: #ffffff;
    font-weight: bold;
    text-align: right;
    padding: 0 8px;
    white-space: nowrap;
    line-height: 20px;
    border-radius: 4px 0 0 4px;
}

.calories {
    background-color: orange;
}
.protein {
    background-color: blue;
}
.fiber {
    background-color: green;
}

.meals-section {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    padding-top: 1rem;
}

.meals-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding-top: 1rem;
}

.meal {
    display: flex;
    flex-direction: column;
    padding: 1rem;
    border: 1px solid #333;
    border-radius: 0.5rem;
    background: #1a1a1a;
    color: #fff;
    align-items: left;
    width: 500px;
}

.meal h3 {
    margin-top: 0;
    margin-bottom: 0.5rem;
}

.food .details {
    opacity: 0;
    visibility: hidden;
    margin-left: 0.25rem;
    transition: visibility 0.3s ease;
    transition-delay: 0.5s;
}

.meal h3:hover ~ .food .details {
    opacity: 1;
    visibility: visible;
    transition-delay: 0s;
}

.food:hover .details {
    opacity: 1;
    visibility: visible;
    transition-delay: 0s;
}

.meal span {
    color: gray;
}

.meal span:hover {
    transition-delay: 0s;
}

.details .cal {
    color: #f87171;
}
.details .pro {
    color: #60a5fa;
}
.details .fib {
    color: #34d399;
}

.expected-header {
    display: flex;
    flex-direction: row;
}
</style>
