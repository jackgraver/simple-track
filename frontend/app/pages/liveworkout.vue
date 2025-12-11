<script setup lang="ts">
import type { Exercise, LoggedExercise, WorkoutLog } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { nextTick } from "vue";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
}>(`workout/previous?offset=0`);

const log = ref<ExerciseGroup[]>(
    data.value?.previous_exercises ?? [],
);

// Watch for data changes
watch(() => data.value, (newData) => {
    if (newData) {
        log.value = newData.previous_exercises ?? [];
    }
}, { immediate: true });

// View state: 'list' or 'logging'
const currentView = ref<"list" | "logging">("list");
const currentExerciseIndex = ref<number | null>(null);

// Set logging state
const currentWeight = ref<number>(0);
const currentReps = ref<number>(0);
const currentSetNumber = ref<number>(1);
const loggedSets = ref<Array<{ weight: number; reps: number }>>([]);

// Edit mode state
const weightEditMode = ref<boolean>(false);
const repsEditMode = ref<boolean>(false);
const weightInputValue = ref<string>("");
const repsInputValue = ref<string>("");

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

// Initialize exercise for logging
const startLoggingExercise = (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;

    currentExerciseIndex.value = index;
    currentView.value = "logging";
    currentSetNumber.value = 1;
    loggedSets.value = [];

    // Initialize weight from previous exercise if available
    if (exerciseGroup.previous && exerciseGroup.previous.sets.length > 0) {
        const firstSet = exerciseGroup.previous.sets[0];
        currentWeight.value = firstSet ? firstSet.weight : 0;
    } else {
        currentWeight.value = 0;
    }
    currentReps.value = 0;
    weightEditMode.value = false;
    repsEditMode.value = false;
};

// Add next set
const addNextSet = () => {
    if (currentWeight.value === 0 && currentReps.value === 0) {
        toast.push("Please enter weight and reps", "error");
        return;
    }

    loggedSets.value.push({
        weight: currentWeight.value,
        reps: currentReps.value,
    });

    currentSetNumber.value++;
    currentReps.value = 0; // Reset reps, keep weight
};

// Finish logging and save
const finishLogging = async () => {
    if (currentWeight.value === 0 && currentReps.value === 0 && loggedSets.value.length === 0) {
        toast.push("Please log at least one set", "error");
        return;
    }

    // Add the current set if it has values
    if (currentWeight.value !== 0 || currentReps.value !== 0) {
        loggedSets.value.push({
            weight: currentWeight.value,
            reps: currentReps.value,
        });
    }

    if (currentExerciseIndex.value === null) return;

    const exerciseGroup = log.value[currentExerciseIndex.value];
    if (!exerciseGroup) return;

    // Get or create the logged exercise
    let exerciseToLog: LoggedExercise;
    if (exerciseGroup.logged) {
        exerciseToLog = { ...exerciseGroup.logged };
    } else if (exerciseGroup.previous) {
        // Copy from previous but reset ID for new log entry
        exerciseToLog = {
            ...exerciseGroup.previous,
            ID: 0,
            workout_log_id: data.value?.day.ID ?? 0,
            sets: [], // Reset sets, we'll add new ones
        };
    } else {
        // Create new logged exercise
        exerciseToLog = {
            ID: 0,
            workout_log_id: data.value?.day.ID ?? 0,
            exercise_id: exerciseGroup.planned.ID,
            exercise: exerciseGroup.planned,
            sets: [],
            weight_setup: "",
            created_at: "",
            updated_at: "",
        };
    }

    // Update sets
    exerciseToLog.sets = loggedSets.value.map((set, index) => ({
        logged_exercise_id: exerciseToLog.ID,
        reps: set.reps,
        weight: set.weight,
        ID: 0,
        created_at: "",
        updated_at: "",
    }));

    exerciseToLog.workout_log_id = data.value?.day.ID ?? exerciseToLog.workout_log_id;

    // Log the exercise
    const success = await logExercise(
        exerciseToLog,
        exerciseGroup.logged ? "logged" : "previous",
    );

    if (success) {
        // Update the log entry
        if (exerciseGroup.logged) {
            exerciseGroup.logged = exerciseToLog;
        } else {
            exerciseGroup.logged = exerciseToLog;
        }
    }

    // Reset and go back to list
    currentView.value = "list";
    currentExerciseIndex.value = null;
    currentWeight.value = 0;
    currentReps.value = 0;
    currentSetNumber.value = 1;
    loggedSets.value = [];
    weightEditMode.value = false;
    repsEditMode.value = false;
};

// Weight increment/decrement functions
const incrementWeight = () => {
    currentWeight.value = (currentWeight.value || 0) + 2.5;
};

const decrementWeight = () => {
    currentWeight.value = Math.max(0, (currentWeight.value || 0) - 2.5);
};

// Reps increment/decrement functions
const incrementReps = () => {
    currentReps.value = (currentReps.value || 0) + 1;
};

const decrementReps = () => {
    currentReps.value = Math.max(0, (currentReps.value || 0) - 1);
};

// Enter edit mode for weight
const enterWeightEditMode = () => {
    weightEditMode.value = true;
    weightInputValue.value = currentWeight.value.toString();
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
    const numValue = parseFloat(weightInputValue.value);
    if (!isNaN(numValue) && numValue >= 0) {
        currentWeight.value = numValue;
    }
    weightEditMode.value = false;
};

// Enter edit mode for reps
const enterRepsEditMode = () => {
    repsEditMode.value = true;
    repsInputValue.value = currentReps.value.toString();
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
    const numValue = parseInt(repsInputValue.value);
    if (!isNaN(numValue) && numValue >= 0) {
        currentReps.value = numValue;
    }
    repsEditMode.value = false;
};

const logExercise = async (
    exercise: LoggedExercise,
    type: "logged" | "previous",
): Promise<boolean> => {
    const rawExercise = toRaw(exercise);
    rawExercise.sets = toRaw(rawExercise.sets).filter(
        (set) => !(set.reps === 0 && set.weight === 0),
    );
    rawExercise.workout_log_id =
        data.value?.day.ID ?? rawExercise.workout_log_id;

    const { response, error } = await useAPIPost<{
        exercise: LoggedExercise;
    }>(
        `workout/exercise/log`,
        "POST",
        {
            exercise: rawExercise,
            type: type,
        },
        {},
        false,
    );

    if (error) {
        console.error(error);
        toast.push(error.message, "error");
        return false;
    }

    return true;
};

const confirmLogs = async () => {
    const { response, error } = await useAPIPost<{
        all_logged: boolean;
    }>(`workout/exercise/all-logged`, "POST", {});

    if (error) {
        console.error(error);
        toast.push(error.message, "error");
        return;
    }
    if (response) {
        if (response.all_logged) {
            toast.push("All logged!", "success");
        } else {
            toast.push("Not all logged!", "error");
        }
    }
};
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div v-if="currentView === 'list'" class="list-view">
            <h2>Exercises</h2>
            <ul class="exercise-list">
                <li
                    v-for="(exerciseGroup, index) in log"
                    :key="index"
                    @click="startLoggingExercise(index)"
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
            <button @click="confirmLogs" class="finish-button">
                <span>Finish Workout</span>
            </button>
        </div>
        <div v-else-if="currentView === 'logging' && currentExerciseIndex !== null && log[currentExerciseIndex]" class="logging-view">
            <div class="logging-header">
                <h2>{{ log[currentExerciseIndex]?.planned.name }}</h2>
                <span class="set-indicator">Set {{ currentSetNumber }}</span>
            </div>

            <div class="sets-logged" v-if="loggedSets.length > 0">
                <h3>Logged Sets:</h3>
                <ul class="sets-list">
                    <li v-for="(set, index) in loggedSets" :key="index">
                        Set {{ index + 1 }}: {{ set.weight }}kg × {{ set.reps }} reps
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
                            v-model="weightInputValue"
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
                            v-model="repsInputValue"
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
                <button @click="addNextSet" class="next-set-button">
                    <span>Next Set</span>
                </button>
                <button @click="finishLogging" class="finish-button">
                    <span>Finish</span>
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

/* List View Styles */
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
}

.exercise-list {
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

/* Logging View Styles */
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
    margin-top: 1rem;
    background: rgb(80, 80, 40) !important;
}

.finish-button:hover {
    background: rgb(100, 100, 50) !important;
}

@media (max-width: 767px) {
    .container {
        padding: 0.5rem;
    }

    .button-group {
        flex-direction: column;
    }
}
</style>
