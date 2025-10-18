<script setup lang="ts">
import { ref } from "vue";
import type { LoggedExercise, LoggedSet } from "~/types/workout";

const props = defineProps<{ exercise: LoggedExercise }>();

// create a separate reactive copy so editing inputs doesn't mutate props
const today = ref<LoggedExercise>(structuredClone(props.exercise));

const setReps = (i: number, event: Event) => {
    if (!today.value.sets[i]) return;
    const input = event.target as HTMLInputElement;
    today.value.sets[i].reps = Number(input.value);
};

const setWeight = (i: number, event: Event) => {
    if (!today.value.sets[i]) return;
    const input = event.target as HTMLInputElement;
    today.value.sets[i].weight = Number(input.value);
};

const logToday = () => {
    console.log("today", today.value);
    console.log("exercise", props.exercise);
};
</script>

<template>
    <h1>{{ exercise.name }}</h1>

    <div v-for="(s, i) in exercise.sets" :key="s.ID" class="exercise-container">
        <p>{{ s.reps }} x {{ s.weight }}</p>
        <div class="log-container">
            <input
                type="number"
                :value="s.reps"
                @change="setReps(i, $event)"
                placeholder="Reps"
            />
            <input
                type="number"
                :value="s.weight"
                @change="setWeight(i, $event)"
                placeholder="Weight"
            />
        </div>
    </div>

    <button @click="logToday">Submit</button>
</template>

<style scoped>
.exercise-container {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
}

.exercise-container p {
    margin: 0;
}

.log-container {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
}

.log-container input {
    width: 70px;
    background: #2a2a2a;
    border: 1px solid #444;
    color: #fff;
    border-radius: 0.25rem;
    padding: 0.25rem;
    text-align: center;
}
</style>
