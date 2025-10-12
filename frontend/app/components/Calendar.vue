<script setup lang="ts">
const props = defineProps<{
    days: any[];
    displayComponent: any;
}>();

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

const allDays = computed(() => {
    if (!props.days) return [];
    return [...props.days].sort(
        (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
    );
});

const firstDayIndex = computed(() => {
    if (!allDays.value.length) return 0;
    const firstDate = new Date(allDays.value[0]!.date);
    return firstDate.getDay(); // Sunday = 0
});
</script>

<template>
    <div class="grid-container">
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
                today && isSameDay(today, day.date) ? 'today' : '',
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

.grid-item > span:first-child {
    text-align: left;
    font-weight: bold;
}

.grid-item.empty {
    background: transparent;
    border: none;
}
</style>
