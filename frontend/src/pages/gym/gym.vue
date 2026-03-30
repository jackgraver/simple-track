<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkoutLogToday } from "~/api/workout/queries";
import { formatDateLong } from "~/utils/dateUtil";

const route = useRoute();
const router = useRouter();
const isGymHome = computed(() => route.name === "gym");
const dayOffset = computed(() => {
    const raw = route.query.offset;
    const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
    return Number.isNaN(value) ? 0 : value;
});
const { data, isPending, isError, error } = useWorkoutLogToday(dayOffset);

const updateOffset = (nextOffset: number) => {
    const nextQuery = { ...route.query };
    if (nextOffset === 0) {
        delete nextQuery.offset;
    } else {
        nextQuery.offset = String(nextOffset);
    }

    router.push({
        name: "gym",
        query: nextQuery,
    });
};

const goToPreviousDay = () => {
    updateOffset(dayOffset.value + 1);
};

const goToNextDay = () => {
    updateOffset(dayOffset.value - 1);
};

const dateLabel = computed(() => {
    const d = data.value?.date;
    return d ? formatDateLong(d) : "";
});

const splitLabel = computed(
    () => data.value?.workout_plan?.name ?? "No split assigned",
);

const loggingRoute = computed(() => ({
    name: "logging",
    query: dayOffset.value === 0 ? {} : { offset: String(dayOffset.value) },
}));
</script>

<template>
    <div class="gym">
        <template v-if="isGymHome">
            <div v-if="isPending" class="state">Loading...</div>
            <div v-else-if="isError" class="state">
                Error: {{ error?.message ?? "Failed to load" }}
            </div>
            <template v-else>
                <div class="date-nav">
                    <button
                        type="button"
                        class="date-nav-button"
                        @click="goToPreviousDay"
                    >
                        Prev
                    </button>
                    <p class="date">{{ dateLabel }}</p>
                    <button
                        type="button"
                        class="date-nav-button"
                        @click="goToNextDay"
                    >
                        Next
                    </button>
                </div>
                <h1 class="split">{{ splitLabel }}</h1>
                <router-link :to="loggingRoute" class="gym-cta"
                    >Log workout</router-link
                >
            </template>
        </template>
        <router-view v-else />
    </div>
</template>

<style scoped>
.gym {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 28rem;
}
.state {
    color: #aaa;
}
.date-nav {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}
.date {
    margin: 0;
    font-size: 0.95rem;
    color: #b0b0b0;
}
.date-nav-button {
    padding: 0.35rem 0.7rem;
    border: 1px solid #666;
    border-radius: 6px;
    background: #2d2d2d;
    color: #fff;
    font: inherit;
    cursor: pointer;
}
.date-nav-button:hover {
    background: #3a3a3a;
}
.split {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
}
.plan-block {
    margin: 0;
}
.plan-heading {
    margin: 0 0 0.5rem;
    font-size: 0.85rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #909090;
}
.plan-list {
    margin: 0;
    padding-left: 1.15rem;
    color: #dcdcdc;
    line-height: 1.5;
}
.plan-item {
    margin: 0.15rem 0;
}
.plan-empty {
    margin: 0;
    font-size: 0.9rem;
    color: #888;
}
.gym-cta {
    display: inline-block;
    align-self: flex-start;
    margin-top: 0.5rem;
    padding: 10px 18px;
    border-radius: 6px;
    background: #3a3a3a;
    color: #fff;
    text-decoration: none;
    font-weight: 500;
    border: 1px solid #666;
}
.gym-cta:hover {
    background: #4a4a4a;
}
</style>
