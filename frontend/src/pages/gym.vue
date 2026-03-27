<script setup lang="ts">
import { computed } from "vue";
import { useWorkoutLogToday } from "~/pages/liveworkout/queries/useWorkoutLogToday";
import { formatDateLong } from "~/utils/dateUtil";

const { data, isPending, isError, error } = useWorkoutLogToday();

const dateLabel = computed(() => {
    const d = data.value?.date;
    return d ? formatDateLong(d) : "";
});

const splitLabel = computed(
    () => data.value?.workout_plan?.name ?? "No split assigned",
);

const planExercises = computed(() => data.value?.workout_plan?.exercises ?? []);
</script>

<template>
    <div class="gym">
        <div v-if="isPending" class="state">Loading...</div>
        <div v-else-if="isError" class="state">
            Error: {{ error?.message ?? "Failed to load" }}
        </div>
        <template v-else>
            <p class="date">{{ dateLabel }}</p>
            <h1 class="split">{{ splitLabel }}</h1>
            <section v-if="planExercises.length" class="plan-block">
                <h2 class="plan-heading">Exercises</h2>
                <ul class="plan-list">
                    <li v-for="ex in planExercises" :key="ex.ID" class="plan-item">
                        {{ ex.name }}
                    </li>
                </ul>
            </section>
            <p v-else-if="data?.workout_plan" class="plan-empty">
                No exercises in this plan yet.
            </p>
            <router-link :to="{ name: 'liveworkout' }" class="gym-cta"
                >Log workout</router-link
            >
        </template>
    </div>
</template>

<style scoped>
.gym {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    max-width: 28rem;
}
.state {
    color: #aaa;
}
.date {
    margin: 0;
    font-size: 0.95rem;
    color: #b0b0b0;
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
