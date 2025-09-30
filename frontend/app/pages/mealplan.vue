<script lang="ts" setup>
import type { MealPlanDay } from "~/types/models";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";

const {
    data: mealPlan,
    pending,
    error,
} = useApiFetch<{
    today: string;
    days: MealPlanDay[];
}>("mealplan/month");

const weekdays = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
];

const allDays = computed(() => {
    if (!mealPlan?.value?.days) return [];
    return [...mealPlan.value.days].sort(
        (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
    );
});

const firstDayIndex = computed(() => {
    if (!allDays.value.length) return 0;
    const firstDate = new Date(allDays.value[0]!.date);
    return firstDate.getDay(); // Sunday = 0
});

const displayDayDialog = async (day: MealPlanDay) => {
    dialogManager.confirm({
        title: "Modify " + formatDate(day.date),
        message: "Are you sure you want to modify this meal?",
        confirmText: "Modify",
        cancelText: "Cancel",
    });
};
</script>

<template>
    <!-- <MealPlanDay v-if="selectedDay" :day="selectedDay" /> -->
    <h1>Meal Plan</h1>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="grid-container">
        <div v-for="weekday in weekdays" :key="weekday" class="weekday-header">
            {{ weekday }}
        </div>
        <div
            v-for="n in firstDayIndex"
            :key="'empty-' + n"
            class="grid-item empty"
        ></div>
        <div
            v-for="day in allDays"
            :key="day.ID"
            :class="[
                'grid-item',
                mealPlan?.today && isSameDay(mealPlan.today, day.date)
                    ? 'today'
                    : '',
            ]"
            @click="
                () => {
                    displayDayDialog(day);
                }
            "
        >
            <span>{{ formatDateShort(day.date) }}</span>
            <div class="macros">
                <div class="macro">
                    <div class="macro-bar">
                        <div
                            class="macro-fill calories"
                            :style="{
                                width: `${Math.min(100, (totalCaloriesEaten(day) / day.goals.calories) * 100)}%`,
                            }"
                        >
                            {{ totalCaloriesEaten(day) }}
                        </div>
                    </div>
                    <div class="macro-goal">{{ day.goals.calories }}</div>
                </div>
                <div class="macro">
                    <div class="macro-bar">
                        <div
                            class="macro-fill protein"
                            :style="{
                                width: `${Math.min(100, (totalProteinEaten(day) / day.goals.protein) * 100)}%`,
                            }"
                        >
                            {{ totalProteinEaten(day) }}g
                        </div>
                    </div>
                    <div class="macro-goal">{{ day.goals.protein }}g</div>
                </div>
                <div class="macro">
                    <div class="macro-bar">
                        <div
                            class="macro-fill fiber"
                            :style="{
                                width: `${Math.min(100, (totalFiberEaten(day) / day.goals.fiber) * 100)}%`,
                            }"
                        >
                            {{ totalFiberEaten(day) }}g
                        </div>
                    </div>
                    <div class="macro-goal">{{ day.goals.fiber }}g</div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.grid-container {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 10px;
}

.weekday-header {
    font-weight: bold;
    margin-bottom: 4px;
    text-align: center;
}

.grid-item {
    background: #252525;
    color: #fff;
    padding: 8px;
    border-radius: 4px;
    margin-bottom: 6px;
    height: 100px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
}

.today {
    background-color: #808080;
}

.grid-item > span:first-child {
    text-align: left;
    font-weight: bold;
}

.grid-item.empty {
    background: transparent;
    border: none;
}

.macros {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.macro {
    display: flex;
    align-items: center;
    gap: 4px;
}

.macro-bar {
    flex: 1;
    background: #444;
    border-radius: 4px;
    height: 20px;
    overflow: hidden;
    position: relative;
}

.macro-fill {
    height: 100%;
    color: #ffffff;
    font-weight: bold;
    text-align: right;
    padding: 0 8px;
    white-space: nowrap;
    line-height: 20px;
    border-radius: 4px 0 0 4px;
}

.macro-goal {
    width: 40px;
    text-align: right;
    font-weight: bold;
}

/* Colors */
.calories {
    background-color: orange;
}
.protein {
    background-color: blue;
}
.fiber {
    background-color: green;
}
</style>
