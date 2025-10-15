<script setup lang="ts">
import { ref } from "vue";
import type { LoggedExercise, WorkoutLog } from "~/types/workout";

const log = ref(false);

const { data, pending, error } = useApiFetch<WorkoutLog>(`workout/previous`);

const workoutLog = data && data.value ? data?.value : null;

const todayLog: WorkoutLog = {
    date: "",
    exercises: workoutLog?.exercises ?? [],
    ID: 0,
    created_at: "",
    updated_at: "",
};
console.log("today", todayLog);

const logWorkout = () => {
    console.log(todayLog);
    log.value = false;
};
</script>

<template>
    <div class="page-container">
        <div class="title-container">
            <h1>Previous {{ workoutLog?.workout_plan?.name }} day</h1>
            <button @click="log = !log">Log</button>
            <button @click="logWorkout" v-if="log">Confirm</button>
        </div>
        <div class="exercise-list-container">
            <div
                v-for="e in workoutLog?.exercises"
                :key="e.ID"
                class="exercise-card"
            >
                <ExcerciseLog :exercise="e" />
                <!-- <h2>{{ e.name }}</h2>
                <div class="set-row">
                    <div class="previous-sets">
                        <h3>Previous</h3>
                        <p v-for="s in e.sets" :key="s.ID">
                            {{ s.reps }} x {{ s.weight }}
                        </p>
                    </div>
                </div> -->
            </div>
            <div class="current-sets" v-if="log">
                <h3>Today</h3>
                <div
                    v-for="e in todayLog.exercises"
                    :key="e.ID"
                    class="set-inputs"
                >
                    <div v-for="s in e.sets" :key="s.ID"></div>
                    <input type="number" :value="todayLog" placeholder="Reps" />
                    <span>x</span>
                    <input
                        type="number"
                        v-model="todayLog"
                        placeholder="Weight"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.page-container {
    padding: 1rem;
    color: #fff;
}

.title-container {
    display: flex;
    align-items: center;
    margin-bottom: 1.5rem;
}

.title-container h1 {
    margin: 0;
    font-size: 1.5rem;
}

.title-container button {
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    cursor: pointer;
}

.exercise-list-container {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.exercise-card {
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 0.5rem;
    padding: 1rem;
    width: 20%;
    width: fit-content;
}

.exercise-card h2 {
    margin: 0 0 0.75rem;
    font-size: 1.2rem;
}

.set-row {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
}

.previous-sets h3,
.current-sets h3 {
    margin-bottom: 0.5rem;
    font-size: 1rem;
    font-weight: 500;
}

.set-inputs {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
}

.set-inputs input {
    width: 70px;
    background: #2a2a2a;
    border: 1px solid #444;
    color: #fff;
    border-radius: 0.25rem;
    padding: 0.25rem;
    text-align: center;
}
</style>
