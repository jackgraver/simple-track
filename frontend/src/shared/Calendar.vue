<script setup lang="ts">
import { computed, ref } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/utils/axios";
import { formatDateShort, isSameDay, isSameMonth, monthName } from "~/utils/dateUtil";

const props = defineProps<{
    fetchURL: string;
    displayComponent: any;
}>();

const monthOffset = ref(0);
const today = ref(new Date());

const { data, isPending, error, refetch } = useQuery({
    queryKey: computed(() => ["calendar", props.fetchURL, monthOffset.value]),
    queryFn: async () => {
        const path = props.fetchURL.startsWith("/") ? props.fetchURL : `/${props.fetchURL}`;
        const res = await apiClient.get<{ today: string; days: any[]; month: number }>(path, {
            params: { monthoffset: monthOffset.value },
        });
        return res.data;
    },
    enabled: computed(() => !!props.fetchURL),
});

function setCurrentMonth() {
    monthOffset.value = 0;
}

function nextMonth() {
    monthOffset.value += 1;
}

function prevMonth() {
    monthOffset.value -= 1;
}

const weekdays = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
];

const headerMonthLabel = computed(() => {
    const days = data.value?.days;
    if (!days?.length) return "";
    const mid = days[Math.floor(days.length / 2)];
    return mid?.date ? monthName(mid.date) : "";
});
</script>

<template>
    <div class="page">
        <h1>{{ headerMonthLabel }}</h1>
        <div v-if="error" class="cal-error">Failed to load calendar</div>
        <div v-else-if="isPending" class="cal-loading">Loading...</div>
        <template v-else>
            <div class="flex gap-2">
                <button type="button" @click="prevMonth">&lt;-</button>
                <button type="button" @click="setCurrentMonth" :disabled="monthOffset === 0">
                    Today
                </button>
                <button type="button" @click="nextMonth">-&gt;</button>
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
                        today && !isSameMonth(data?.month ?? 0, day.date) ? 'faded' : '',
                    ]"
                    @click="() => {}"
                >
                    <span>{{ formatDateShort(day.date) }}</span>
                    <component :is="displayComponent" :day="day" />
                </div>
            </div>
        </template>
    </div>
</template>

<style scoped>
.page {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 1rem;
    width: 75%;
    max-width: 1200px;
    margin: 0 auto;
    min-height: 100vh;
}

.flex {
    display: flex;
}
.gap-2 {
    gap: 0.5rem;
}

.cal-error,
.cal-loading {
    padding: 1rem;
}

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
