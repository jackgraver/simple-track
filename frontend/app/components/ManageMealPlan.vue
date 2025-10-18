<script lang="ts" setup>
import type { Day } from "~/types/diet";

const {
    data: mealPlan,
    pending,
    error,
} = useAPIGet<{
    today: string;
    days: Day[];
}>("mealplan/week");
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
                <p>{{ formatDate(day.date) }}</p>
                <p>{{ dayOfWeek(day.date) }}</p>
                <span>
                    <p class="calories">
                        C {{ day.totalCalories ?? 0 }}/{{ day.plan.calories }}
                    </p>
                    <p class="protein">
                        P {{ day.totalProtein ?? 0 }}/{{ day.plan.protein }}g
                    </p>
                    <p class="fiber">
                        F {{ day.totalFiber ?? 0 }}/{{ day.plan.fiber }}g
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
