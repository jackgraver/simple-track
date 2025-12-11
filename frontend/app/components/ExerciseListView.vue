<script setup lang="ts">
import type { Exercise, LoggedExercise } from "~/types/workout";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const props = defineProps<{
    exercises: ExerciseGroup[];
}>();

const emit = defineEmits<{
    (e: "select-exercise", index: number): void;
    (e: "finish-workout"): void;
}>();

// Get maximum weight from previous exercise
const getMaxWeight = (exerciseGroup: ExerciseGroup): number | null => {
    if (!exerciseGroup.previous || !exerciseGroup.previous.sets || exerciseGroup.previous.sets.length === 0) {
        return null;
    }
    return Math.max(...exerciseGroup.previous.sets.map(set => set.weight));
};

// Get last set from logged exercise
const getLastSet = (exerciseGroup: ExerciseGroup): { weight: number; reps: number } | null => {
    if (!exerciseGroup.logged || !exerciseGroup.logged.sets || exerciseGroup.logged.sets.length === 0) {
        return null;
    }
    const lastSet = exerciseGroup.logged.sets[exerciseGroup.logged.sets.length - 1];
    if (!lastSet) {
        return null;
    }
    return {
        weight: lastSet.weight,
        reps: lastSet.reps,
    };
};

// Check if exercise is logged
const isLogged = (exerciseGroup: ExerciseGroup): boolean => {
    return !!exerciseGroup.logged && exerciseGroup.logged.sets && exerciseGroup.logged.sets.length > 0;
};
</script>

<template>
    <div class="list-view">
        <h2>Exercises</h2>
        <ul class="exercise-list">
            <li
                v-for="(exerciseGroup, index) in exercises"
                :key="index"
                @click="emit('select-exercise', index)"
                :class="['exercise-item', { 'logged': isLogged(exerciseGroup) }]"
            >
                <span class="exercise-name">{{ exerciseGroup.planned.name }}</span>
                <div class="exercise-info">
                    <span v-if="isLogged(exerciseGroup) && getLastSet(exerciseGroup)" class="last-set">
                        {{ getLastSet(exerciseGroup)!.weight }}lbs × {{ getLastSet(exerciseGroup)!.reps }}
                    </span>
                    <span v-else-if="!isLogged(exerciseGroup) && getMaxWeight(exerciseGroup) !== null" class="previous-weight">
                        Prev {{ getMaxWeight(exerciseGroup) }}lbs
                    </span>
                </div>
            </li>
        </ul>
        <button @click="emit('finish-workout')" class="finish-button">
            <span>Finish Workout</span>
        </button>
    </div>
</template>

<style scoped>
.list-view {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.list-view h2 {
    margin: 0 0 1rem 0;
}

.exercise-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
}

.exercise-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    cursor: pointer;
    transition: background-color 0.2s, opacity 0.2s;
    width: 100%;
    box-sizing: border-box;
}

.exercise-item:hover {
    background: rgb(40, 40, 40);
}

.exercise-item.logged {
    opacity: 0.6;
    background: rgb(20, 20, 20);
}

.exercise-item.logged:hover {
    background: rgb(30, 30, 30);
    opacity: 0.8;
}

.exercise-name {
    font-weight: 500;
    font-size: 1.1rem;
}

.exercise-info {
    display: flex;
    align-items: center;
}

.previous-weight {
    color: rgb(150, 150, 150);
    font-size: 0.9rem;
}

.last-set {
    color: rgb(200, 200, 200);
    font-size: 0.9rem;
    font-weight: 500;
}

.finish-button {
    margin-top: 1rem;
    padding: 0.75rem 1.5rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(80, 80, 40) !important;
    color: inherit;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.finish-button:hover {
    background: rgb(100, 100, 50) !important;
}
</style>

