<script setup lang="ts">
import { isSameMonth } from "~/utils/dateUtil";

const props = defineProps<{
    // days: any[];
    fetchURL: string;
    displayComponent: any;
}>();

const monthOffset = ref(0);

const data = ref<{ today: string; days: any[]; month: number } | null>(null);
const pending = ref(false);
const error = ref<Error | null>(null);

async function fetchMealPlan() {
    pending.value = true;
    error.value = null;
    try {
        const res = await useAPIGet<{
            today: string;
            days: any[];
            month: number;
        }>(props.fetchURL, {
            query: {
                monthoffset: monthOffset.value,
            },
        });
        data.value = res.data.value ?? null;
        console.log(data.value);
    } catch (err) {
        error.value = err as Error;
    } finally {
        pending.value = false;
    }
}

function setCurrentMonth() {
    monthOffset.value = 0;
    fetchMealPlan();
}

function nextMonth() {
    monthOffset.value += 1;
    fetchMealPlan();
}

function prevMonth() {
    monthOffset.value -= 1;
    fetchMealPlan();
}

onMounted(() => {
    setCurrentMonth();
});

const today = ref(new Date());

const weekdays = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
];
</script>

<template>
    <h1>
        {{ monthName(data?.days[Math.floor(data?.days.length / 2)]?.date) }}
    </h1>
    <div class="flex gap-2">
        <button @click="prevMonth"><-</button>
        <button @click="setCurrentMonth" :disabled="monthOffset === 0">
            Today
        </button>
        <button @click="nextMonth">-></button>
    </div>
    <div class="grid-container">
        <div v-for="weekday in weekdays" :key="weekday" class="weekday-header">
            {{ weekday }}
        </div>
        <div
            v-for="day in data?.days"
            :key="day.ID"
            :class="[
                'grid-item',
                today && isSameDay(today, day.date) ? 'today' : '',
                today && !isSameMonth(data?.month ?? 0, day.date)
                    ? 'faded'
                    : '',
            ]"
            @click="() => {}"
        >
            <span>{{ formatDateShort(day.date) }}</span>
            <props.displayComponent :day="day" />
        </div>
    </div>
</template>

<style scoped>
.grid-container {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 10px;
    width: 100%;
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

.faded {
    opacity: 0.3;
    filter: grayscale(1);
}

.grid-item > span:first-child {
    text-align: left;
    font-weight: bold;
}

.grid-item.empty {
    background: transparent;
    border: none;
}
</style>
