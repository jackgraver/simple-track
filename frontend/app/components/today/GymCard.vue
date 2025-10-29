<script setup lang="ts">
import type { LoggedExercise, LoggedSet } from "~/types/workout";
import { Check, Plus } from "lucide-vue-next";
import { dialogManager } from "~/composables/dialog/useDialog";
import TodayLogExerciseDialog from "./LogExerciseDialog.vue";

const props = defineProps<{
    exercise: LoggedExercise;
}>();

const logExercise = async () => {
    dialogManager.custom({
        title: "Log Exercise",
        component: TodayLogExerciseDialog,
        props: { exercise: props.exercise },
    });
};
</script>

<template>
    <article v-if="exercise.sets && exercise.sets.length" class="workout-card">
        <header class="card-header">
            <h3>{{ exercise.exercise?.name }}</h3>
            {{ exercise.sets[exercise.sets.length - 1]?.weight ?? "X" }}
            x
            {{ exercise.sets[exercise.sets.length - 1]?.reps ?? "X" }}
            <h4 v-if="exercise.weight_setup !== ''">
                {{ exercise.weight_setup }}
            </h4>
        </header>
        <span v-if="exercise.weight_setup !== ''">
            {{ exercise.weight_setup }}
        </span>
        <span
            v-if="
                (exercise.sets[exercise.sets.length - 1]?.reps ?? 0) >
                exercise.exercise?.rep_rollover
            "
            class="info"
            >Up weight next session</span
        >
        <span v-if="exercise.percent_change" class="fact"
            >{{ exercise.percent_change < 0 ? "Down" : "Up" }}
            {{ exercise.percent_change.toFixed(2) }}% since last workout</span
        >
        <button @click="logExercise" class="check">
            <Check />
        </button>
    </article>
</template>

<style scoped>
.workout-card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid #333;
    border-radius: 0.5rem;
    background: #1a1a1a;
    color: #fff;
    min-height: 120px;
}

.workout-card span {
    font-size: small;
}

.card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
}

.card-header h3 {
    margin: 0;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.workout-card .info {
    color: rgb(206, 206, 48);
}

.workout-card .fact {
    color: rgb(63, 197, 46);
}

footer {
    display: flex;
    justify-content: flex-end;
    padding-bottom: 0;
}

.check {
    background-color: rgb(63, 197, 46);
    color: #fff;
    border-radius: 0.25rem;
    padding: 0.25rem 0.5rem;
}
</style>
