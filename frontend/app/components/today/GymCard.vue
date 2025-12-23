<script setup lang="ts">
import type { Exercise, LoggedExercise } from "~/types/workout";

const weightString = (log: LoggedExercise): string => {
    if (!log.sets || log.sets.length === 0) {
        return "";
    }
    const maxWeight = Math.max(...log.sets.map(set => set.weight || 0));
    if (maxWeight > 0) {
        return maxWeight + " lbs";
    }
    return "";
};

const props = defineProps<{
    exercise: LoggedExercise;
    previous: LoggedExercise | null;
    planned: boolean;
}>();

const localExercise = ref<LoggedExercise>(props.exercise);
const localPlanned = ref<boolean>(props.planned);

const actualExercise = computed<Exercise>(() => localExercise.value.exercise);
const isUsingLocalExercise = computed(() => props.previous === null);
</script>

<template>
    <article class="workout-card">
        <header class="card-header">
            <h2>{{ actualExercise?.name }}</h2>
            <h3 :class="{ 'local-exercise': isUsingLocalExercise }">{{ weightString(previous ?? localExercise) }}</h3>
        </header>
        <span>{{ localExercise.notes }}</span>
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

.workout-card header h2 {
    margin: 0;
}

.workout-card footer {
    display: flex;
    justify-content: flex-end;
    margin-top: auto; /* pushes footer to bottom */
}

.sets {
    display: flex;
    flex-direction: row;
    gap: 0.3rem;
}

.sets span:not(:last-child)::after {
    content: ",";
    margin-left: 0.05rem;
}

.card-header h3 {
    margin: 0;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.card-header h3.local-exercise {
    color: rgb(63, 197, 46);
}

.workout-card .info {
    color: rgb(206, 206, 48);
}

.workout-card .fact {
    color: rgb(63, 197, 46);
}

.check {
    background-color: rgb(63, 197, 46);
    color: #fff;
    border-radius: 0.25rem;
    padding: 0.25rem 0.5rem;
}
</style>
