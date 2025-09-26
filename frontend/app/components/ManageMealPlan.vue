<script lang="ts" setup>
import type { MealPlanDay } from "~/types/models";

const {
    data: mealPlan,
    pending,
    error,
} = useApiFetch<{
    today: string;
    days: MealPlanDay[];
}>("mealplan/week");

function totalCaloriesEaten(day: MealPlanDay): string {
    return day.meals
        .filter((dm) => dm.status === "actual") // only actual meals
        .reduce((sum, dm) => {
            // sum calories for each MealItem in the meal
            const mealCalories = dm.meal.items.reduce((mealSum, item) => {
                if (!item.food) return mealSum;
                return mealSum + item.food.calories * item.amount;
            }, 0);
            return sum + mealCalories;
        }, 0)
        .toFixed(2);
}
function totalProteinEaten(day: MealPlanDay): string {
    return day.meals
        .filter((dm) => dm.status === "actual") // only actual meals
        .reduce((sum, dm) => {
            // sum calories for each MealItem in the meal
            const mealProtein = dm.meal.items.reduce((mealSum, item) => {
                if (!item.food) return mealSum;
                return mealSum + item.food.protein * item.amount;
            }, 0);
            return sum + mealProtein;
        }, 0)
        .toFixed(2);
}
function totalFiberEaten(day: MealPlanDay): string {
    return day.meals
        .filter((dm) => dm.status === "actual") // only actual meals
        .reduce((sum, dm) => {
            // sum calories for each MealItem in the meal
            const mealFiber = dm.meal.items.reduce((mealSum, item) => {
                if (!item.food) return mealSum;
                return mealSum + item.food.fiber * item.amount;
            }, 0);
            return sum + mealFiber;
        }, 0)
        .toFixed(2);
}
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
        <div class="days">
            <div
                v-for="day in mealPlan?.days"
                :key="day.ID"
                :class="[
                    'day',
                    mealPlan?.today && isSameDay(mealPlan.today, day.date)
                        ? 'today'
                        : '',
                ]"
            >
                <p>{{ formatDateShort(day.date) }}</p>
                <p>{{ dayOfWeek(day.date) }}</p>
                <span>
                    <p class="calories">
                        C {{ totalCaloriesEaten(day) }}/{{ day.goals.calories }}
                    </p>
                    <p class="protein">
                        P {{ totalProteinEaten(day) }}/{{ day.goals.protein }}g
                    </p>
                    <p class="fiber">
                        F {{ totalFiberEaten(day) }}/{{ day.goals.fiber }}g
                    </p>
                </span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.days {
    display: flex;
    flex-direction: row;
}

.day {
    padding: 16px;
}

.today {
    background-color: rgba(129, 129, 129, 0.377);
}

.days > span {
    display: flex;
    flex-direction: row;
    gap: 4px;
}
.calories {
    color: orange;
}
.protein {
    color: blue;
}
.fiber {
    color: green;
}
</style>
