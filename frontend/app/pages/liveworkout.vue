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
}>(`workout/previous?offset=-1`);

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

// Save state for debouncing
const saveTimeout = ref<number | null>(null);
const isSaving = ref<boolean>(false);


// Initialize exercise for logging
const startLoggingExercise = (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;

    currentExerciseIndex.value = index;
    currentView.value = "logging";
    
    // Load existing sets if exercise was partially logged
    if (exerciseGroup.logged && exerciseGroup.logged.sets && exerciseGroup.logged.sets.length > 0) {
        loggedSets.value = exerciseGroup.logged.sets.map(set => ({
            weight: set.weight,
            reps: set.reps,
        }));
        currentSetNumber.value = loggedSets.value.length + 1;
        
        // Initialize weight/reps from last logged set
        const lastSet = exerciseGroup.logged.sets[exerciseGroup.logged.sets.length - 1];
        if (lastSet) {
            currentWeight.value = lastSet.weight;
            currentReps.value = 0; // Reset reps for new set
        } else {
            currentWeight.value = 0;
            currentReps.value = 0;
        }
    } else {
        loggedSets.value = [];
        currentSetNumber.value = 1;
        
        // Initialize weight from previous exercise if available
        if (exerciseGroup.previous && exerciseGroup.previous.sets.length > 0) {
            const firstSet = exerciseGroup.previous.sets[0];
            currentWeight.value = firstSet ? firstSet.weight : 0;
        } else {
            currentWeight.value = 0;
        }
        currentReps.value = 0;
    }
    
    weightEditMode.value = false;
    repsEditMode.value = false;
};

// Save current exercise state to backend
const saveCurrentExercise = async (): Promise<boolean> => {
    if (currentExerciseIndex.value === null) return false;

    const exerciseGroup = log.value[currentExerciseIndex.value];
    if (!exerciseGroup) return false;

    // Build all sets including current input if valid
    const allSets = [...loggedSets.value];
    if (currentWeight.value !== 0 || currentReps.value !== 0) {
        allSets.push({
            weight: currentWeight.value,
            reps: currentReps.value,
        });
    }

    // Skip save if no sets
    if (allSets.length === 0) return false;

    // Get or create the logged exercise
    let exerciseToLog: LoggedExercise;
    if (exerciseGroup.logged && exerciseGroup.logged.ID > 0) {
        exerciseToLog = { ...exerciseGroup.logged };
    } else if (exerciseGroup.previous) {
        // Copy from previous but reset ID for new log entry
        exerciseToLog = {
            ...exerciseGroup.previous,
            ID: 0,
            workout_log_id: data.value?.day.ID ?? 0,
            sets: [],
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
    exerciseToLog.sets = allSets.map((set) => ({
        logged_exercise_id: exerciseToLog.ID,
        reps: set.reps,
        weight: set.weight,
        ID: 0,
        created_at: "",
        updated_at: "",
    }));

    exerciseToLog.workout_log_id = data.value?.day.ID ?? exerciseToLog.workout_log_id;

    // Determine type
    const type: "logged" | "previous" = (exerciseGroup.logged && exerciseGroup.logged.ID > 0) ? "logged" : "previous";

    // Save to backend
    const success = await logExercise(exerciseToLog, type);

    if (success) {
        // Update the log entry with saved exercise
        exerciseGroup.logged = exerciseToLog;
        return true;
    }

    return false;
};

// Debounced save function
const debouncedSave = () => {
    // Clear existing timeout
    if (saveTimeout.value) {
        clearTimeout(saveTimeout.value);
    }

    // Set new timeout
    saveTimeout.value = setTimeout(async () => {
        if (isSaving.value) return; // Prevent concurrent saves
        
        isSaving.value = true;
        try {
            await saveCurrentExercise();
        } catch (error) {
            console.error("Error saving exercise:", error);
            toast.push("Failed to save. Will retry on next action.", "error");
        } finally {
            isSaving.value = false;
        }
    }, 400); // 400ms debounce
};

// Add next set
const addNextSet = async () => {
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

    // Save immediately after adding set
    debouncedSave();
};

// Finish logging and save
const finishLogging = async () => {
    if (currentWeight.value === 0 && currentReps.value === 0 && loggedSets.value.length === 0) {
        toast.push("Please log at least one set", "error");
        return;
    }

    // Clear any pending debounced save
    if (saveTimeout.value) {
        clearTimeout(saveTimeout.value);
        saveTimeout.value = null;
    }

    // Add the current set if it has values (but don't add to loggedSets yet)
    // saveCurrentExercise will include it in the save
    const hasCurrentSet = currentWeight.value !== 0 || currentReps.value !== 0;

    // Ensure final state is saved (may already be saved, but ensure current set is included)
    const success = await saveCurrentExercise();

    if (!success && hasCurrentSet) {
        // If save failed and we have a current set, add it to loggedSets and try again
        loggedSets.value.push({
            weight: currentWeight.value,
            reps: currentReps.value,
        });
        const retrySuccess = await saveCurrentExercise();
        if (!retrySuccess) {
            toast.push("Failed to save exercise. Please try again.", "error");
            return;
        }
    } else if (!success) {
        toast.push("Failed to save exercise. Please try again.", "error");
        return;
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
    isSaving.value = false;
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

// Add exercise to workout
const addExerciseToWorkout = async (exerciseId: number) => {
    const { response, error } = await useAPIPost<{
        exercise: LoggedExercise;
    }>(`workout/exercise/add`, "POST", {
        exercise_id: exerciseId,
    });

    if (error) {
        console.error(error);
        toast.push(error.message, "error");
        return;
    }

    if (response?.exercise) {
        // Add the new exercise to the log
        const newExerciseGroup: ExerciseGroup = {
            planned: response.exercise.exercise!,
            logged: response.exercise,
            previous: {} as LoggedExercise,
        };
        log.value.push(newExerciseGroup);
        toast.push(`Added ${response.exercise.exercise?.name}`, "success");
    }
};

// Remove exercise from workout
const removeExerciseFromWorkout = async (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;

    // If exercise is logged, delete it from backend
    if (exerciseGroup.logged && exerciseGroup.logged.ID > 0) {
        const exerciseId = exerciseGroup.logged.exercise_id;
        if (!exerciseId) {
            toast.push("Cannot remove exercise: ID not found", "error");
            return;
        }

        const { error } = await useAPIPost(`workout/exercise/remove`, "DELETE", {
            exercise_id: exerciseId,
        });

        if (error) {
            console.error(error);
            toast.push(error.message, "error");
            return;
        }
    }

    // Remove from local list (works for both planned-only and logged exercises)
    log.value.splice(index, 1);
    toast.push("Exercise removed", "success");
};
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <ExerciseListView
            v-if="currentView === 'list'"
            :exercises="log"
            @select-exercise="startLoggingExercise"
            @finish-workout="confirmLogs"
            @add-exercise="addExerciseToWorkout"
            @remove-exercise="removeExerciseFromWorkout"
        />
        <ExerciseLoggingView
            v-else-if="currentView === 'logging' && currentExerciseIndex !== null && log[currentExerciseIndex]"
            :exercise-group="log[currentExerciseIndex]!"
            :current-set-number="currentSetNumber"
            :logged-sets="loggedSets"
            :current-weight="currentWeight"
            :current-reps="currentReps"
            :weight-edit-mode="weightEditMode"
            :reps-edit-mode="repsEditMode"
            :weight-input-value="weightInputValue"
            :reps-input-value="repsInputValue"
            @update:current-weight="currentWeight = $event"
            @update:current-reps="currentReps = $event"
            @update:weight-edit-mode="weightEditMode = $event"
            @update:reps-edit-mode="repsEditMode = $event"
            @update:weight-input-value="weightInputValue = $event"
            @update:reps-input-value="repsInputValue = $event"
            @increment-weight="incrementWeight"
            @decrement-weight="decrementWeight"
            @increment-reps="incrementReps"
            @decrement-reps="decrementReps"
            @enter-weight-edit="enterWeightEditMode"
            @exit-weight-edit="exitWeightEditMode"
            @enter-reps-edit="enterRepsEditMode"
            @exit-reps-edit="exitRepsEditMode"
            @add-next-set="addNextSet"
            @finish-logging="finishLogging"
        />
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

@media (max-width: 767px) {
    .container {
        padding: 0.5rem;
    }
}
</style>
