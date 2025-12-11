<script setup lang="ts">
import type { Exercise, LoggedExercise } from "~/types/workout";
import { nextTick } from "vue";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const props = defineProps<{
    exerciseGroup: ExerciseGroup;
    currentSetNumber: number;
    loggedSets: Array<{ weight: number; reps: number }>;
    currentWeight: number;
    currentReps: number;
    weightEditMode: boolean;
    repsEditMode: boolean;
    weightInputValue: string;
    repsInputValue: string;
}>();

const emit = defineEmits<{
    (e: "update:currentWeight", value: number): void;
    (e: "update:currentReps", value: number): void;
    (e: "update:weightEditMode", value: boolean): void;
    (e: "update:repsEditMode", value: boolean): void;
    (e: "update:weightInputValue", value: string): void;
    (e: "update:repsInputValue", value: string): void;
    (e: "increment-weight"): void;
    (e: "decrement-weight"): void;
    (e: "increment-reps"): void;
    (e: "decrement-reps"): void;
    (e: "enter-weight-edit"): void;
    (e: "exit-weight-edit"): void;
    (e: "enter-reps-edit"): void;
    (e: "exit-reps-edit"): void;
    (e: "add-next-set"): void;
    (e: "finish-logging"): void;
}>();

// Weight increment/decrement functions
const incrementWeight = () => {
    emit("increment-weight");
};

const decrementWeight = () => {
    emit("decrement-weight");
};

// Reps increment/decrement functions
const incrementReps = () => {
    emit("increment-reps");
};

const decrementReps = () => {
    emit("decrement-reps");
};

// Enter edit mode for weight
const enterWeightEditMode = () => {
    emit("enter-weight-edit");
    nextTick(() => {
        const input = document.getElementById('weight-input') as HTMLInputElement;
        if (input) {
            input.focus();
            input.select();
        }
    });
};

// Exit edit mode for weight
const exitWeightEditMode = () => {
    emit("exit-weight-edit");
};

// Enter edit mode for reps
const enterRepsEditMode = () => {
    emit("enter-reps-edit");
    nextTick(() => {
        const input = document.getElementById('reps-input') as HTMLInputElement;
        if (input) {
            input.focus();
            input.select();
        }
    });
};

// Exit edit mode for reps
const exitRepsEditMode = () => {
    emit("exit-reps-edit");
};
</script>

<template>
    <div class="logging-view">
        <div class="logging-header">
            <h2>{{ exerciseGroup.planned.name }}</h2>
            <span class="set-indicator">Set {{ currentSetNumber }}</span>
        </div>

        <div class="sets-logged" v-if="loggedSets.length > 0">
            <h3>Logged Sets:</h3>
            <ul class="sets-list">
                <li v-for="(set, index) in loggedSets" :key="index">
                    Set {{ index + 1 }}: {{ set.weight }}lbs × {{ set.reps }} reps
                </li>
            </ul>
        </div>
        <div class="input-group">
            <div class="stepper-container">
                <label>Weight (lbs)</label>
                <div class="stepper">
                    <button @click="decrementWeight" class="stepper-button" type="button">−</button>
                    <div 
                        v-if="!weightEditMode" 
                        @click="enterWeightEditMode" 
                        class="stepper-display"
                    >
                        {{ currentWeight || 0 }}
                    </div>
                    <input
                        v-else
                        id="weight-input"
                        type="number"
                        :value="weightInputValue"
                        @input="emit('update:weightInputValue', ($event.target as HTMLInputElement).value)"
                        @blur="exitWeightEditMode"
                        @keyup.enter="exitWeightEditMode"
                        @keyup.escape="exitWeightEditMode"
                        class="stepper-input"
                        min="0"
                        step="0.5"
                    />
                    <button @click="incrementWeight" class="stepper-button" type="button">+</button>
                </div>
            </div>
            <div class="stepper-container">
                <label>Reps</label>
                <div class="stepper">
                    <button @click="decrementReps" class="stepper-button" type="button">−</button>
                    <div 
                        v-if="!repsEditMode" 
                        @click="enterRepsEditMode" 
                        class="stepper-display"
                    >
                        {{ currentReps || 0 }}
                    </div>
                    <input
                        v-else
                        id="reps-input"
                        type="number"
                        :value="repsInputValue"
                        @input="emit('update:repsInputValue', ($event.target as HTMLInputElement).value)"
                        @blur="exitRepsEditMode"
                        @keyup.enter="exitRepsEditMode"
                        @keyup.escape="exitRepsEditMode"
                        class="stepper-input"
                        min="0"
                    />
                    <button @click="incrementReps" class="stepper-button" type="button">+</button>
                </div>
            </div>
        </div>
        <div class="button-group">
            <button @click="emit('add-next-set')" class="next-set-button">
                <span>Next Set</span>
            </button>
            <button @click="emit('finish-logging')" class="finish-button">
                <span>Finish</span>
            </button>
        </div>
    </div>
</template>

<style scoped>
.logging-view {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.logging-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 1rem;
    border-bottom: 1px solid rgb(56, 56, 56);
}

.logging-header h2 {
    margin: 0;
}

.set-indicator {
    color: rgb(150, 150, 150);
    font-size: 0.9rem;
}

.sets-logged {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.sets-logged h3 {
    margin: 0;
    font-size: 1rem;
    color: rgb(150, 150, 150);
}

.sets-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.sets-list li {
    padding: 0.5rem;
    background: rgb(27, 27, 27);
    border-radius: 3px;
    border: 1px solid rgb(56, 56, 56);
}

.input-group {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.stepper-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.stepper-container label {
    font-weight: 500;
    font-size: 0.9rem;
    color: rgb(150, 150, 150);
}

.stepper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
}

.stepper-button {
    width: 3rem;
    height: 3rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1.5rem;
    font-weight: 300;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s, border-color 0.2s;
    user-select: none;
}

.stepper-button:hover {
    background: rgb(40, 40, 40);
    border-color: rgb(100, 100, 100);
}

.stepper-button:active {
    background: rgb(50, 50, 50);
}

.stepper-display {
    min-width: 6rem;
    height: 3rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
    font-weight: 500;
    cursor: pointer;
    user-select: none;
    padding: 0 1rem;
    border-radius: 5px;
    transition: background-color 0.2s;
}

.stepper-display:hover {
    background: rgb(30, 30, 30);
}

.stepper-input {
    min-width: 6rem;
    height: 3rem;
    padding: 0 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 2rem;
    font-weight: 500;
    text-align: center;
}

.stepper-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}

.button-group {
    display: flex;
    gap: 1rem;
}

.button-group button,
.finish-button {
    flex: 1;
    padding: 0.75rem 1.5rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.button-group button:hover,
.finish-button:hover {
    background: rgb(40, 40, 40);
}

.next-set-button {
    background: rgb(40, 80, 40) !important;
}

.next-set-button:hover {
    background: rgb(50, 100, 50) !important;
}

.finish-button {
    background: rgb(80, 80, 40) !important;
}

.finish-button:hover {
    background: rgb(100, 100, 50) !important;
}

@media (max-width: 767px) {
    .button-group {
        flex-direction: column;
    }
}
</style>

