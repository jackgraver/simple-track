<script lang="ts" setup>
import type { MealPlanDay } from "~/types/models";
import { formatDateShort } from "~/utils/dateUtil";

const props = defineProps<{
    day: MealPlanDay;
}>();

// const { data, pending, error } = useApiFetch<MealPlanDay>(
//     `mealplan/day/${props.meal_plan_day_id}`,
// );
</script>

<template>
    <!-- <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div> -->
    <div class="container">
        <h1>Modify {{ formatDate(day.date) }}</h1>
        <div
            v-for="meal in day.meals.filter((m) => m.status === 'actual')"
            :key="meal.ID"
        >
            <h3>{{ meal.meal.name }}<button>X</button></h3>

            <p v-for="food in meal.meal.items" :key="food.ID">
                {{ food?.food?.name }} {{ food.amount }}
                <button>X</button>
            </p>
        </div>
    </div>
</template>

<style scoped>
.container {
    position: absolute;
    top: 25%;
    left: 25%;
    z-index: 999;
    width: 60%;
    height: 50%;
    background-color: rgb(43, 43, 43);
}
</style>
