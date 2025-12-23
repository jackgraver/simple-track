<script setup lang="ts">
import type { Exercise, LoggedExercise, WorkoutLog } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { nextTick } from "vue";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
}>(`workout/logs/previous?offset=-1`);

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
const currentWeightSetup = ref<string>("");
const currentSetNumber = ref<number>(1);

type LoggedSetWithStatus = {
    weight: number;
    reps: number;
    weight_setup: string;
    status: 'pending' | 'success' | 'error';
    id: number | null;
    error: string | null;
    tempId: string; // Temporary ID for tracking during save
};

const loggedSets = ref<LoggedSetWithStatus[]>([]);
let tempIdCounter = 0;

// Edit mode state
const weightEditMode = ref<boolean>(false);
const repsEditMode = ref<boolean>(false);
const weightInputValue = ref<string>("");
const repsInputValue = ref<string>("");
const notes = ref<string>("");

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
            weight_setup: set.weight_setup || "",
            status: 'success' as const,
            id: set.ID,
            error: null,
            tempId: `existing-${set.ID}`,
        }));
        currentSetNumber.value = loggedSets.value.length + 1;
        
        // Initialize weight/reps/weight_setup from last logged set
        const lastSet = exerciseGroup.logged.sets[exerciseGroup.logged.sets.length - 1];
        if (lastSet) {
            currentWeight.value = lastSet.weight;
            currentReps.value = 0; // Reset reps for new set
            currentWeightSetup.value = lastSet.weight_setup || "";
        } else {
            currentWeight.value = 0;
            currentReps.value = 0;
            currentWeightSetup.value = "";
        }
        
        // Initialize notes from logged exercise
        notes.value = exerciseGroup.logged.notes || "";
    } else {
        loggedSets.value = [];
        currentSetNumber.value = 1;
        
        // Initialize weight/weight_setup from previous exercise if available
        if (exerciseGroup.previous && exerciseGroup.previous.sets.length > 0) {
            const firstSet = exerciseGroup.previous.sets[0];
            currentWeight.value = firstSet ? firstSet.weight : 0;
            currentWeightSetup.value = firstSet ? (firstSet.weight_setup || "") : "";
        } else {
            currentWeight.value = 0;
            currentWeightSetup.value = "";
        }
        currentReps.value = 0;
        
        // Initialize notes from previous exercise if available
        notes.value = exerciseGroup.previous?.notes || "";
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
        const tempId = `temp-current-${Date.now()}`;
        allSets.push({
            weight: currentWeight.value,
            reps: currentReps.value,
            weight_setup: currentWeightSetup.value,
            status: 'pending' as const,
            id: null,
            error: null,
            tempId,
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
            notes: "",
            created_at: "",
            updated_at: "",
        };
    }

    // Update sets - map from LoggedSetWithStatus
    exerciseToLog.sets = allSets.map((set) => ({
        logged_exercise_id: exerciseToLog.ID,
        reps: set.reps,
        weight: set.weight,
        weight_setup: set.weight_setup || "",
        ID: set.id || 0, // Use existing ID if available
        created_at: "",
        updated_at: "",
    }));

    // Update notes
    exerciseToLog.notes = notes.value;

    exerciseToLog.workout_log_id = data.value?.day.ID ?? exerciseToLog.workout_log_id;

    // Determine type
    const type: "logged" | "previous" = (exerciseGroup.logged && exerciseGroup.logged.ID > 0) ? "logged" : "previous";

    // Save to backend
    const savedExercise = await logExercise(exerciseToLog, type);

    if (savedExercise && savedExercise.sets) {
        // Match returned sets with loggedSets by comparing weight/reps/weight_setup
        // and update their status and IDs
        const pendingSets = loggedSets.value.filter(s => s.status === 'pending');
        
        savedExercise.sets.forEach((savedSet) => {
            if (!savedSet) return;
            
            // Try to match pending sets by weight/reps/weight_setup
            const pendingSet = pendingSets.find(ps => 
                Math.abs(ps.weight - savedSet.weight) < 0.01 &&
                ps.reps === savedSet.reps &&
                ps.weight_setup === (savedSet.weight_setup || "")
            );
            
            if (pendingSet) {
                // Update the set in loggedSets
                const setIndex = loggedSets.value.findIndex(s => s.tempId === pendingSet.tempId);
                if (setIndex !== -1 && savedSet.ID) {
                    const setToUpdate = loggedSets.value[setIndex];
                    if (setToUpdate) {
                        setToUpdate.status = 'success';
                        setToUpdate.id = savedSet.ID;
                        setToUpdate.error = null;
                    }
                }
            } else if (savedSet.ID) {
                // Match existing sets by ID (for sets that were already saved)
                const existingSet = loggedSets.value.find(s => s.id === savedSet.ID);
                if (existingSet) {
                    existingSet.status = 'success';
                    existingSet.error = null;
                }
            }
        });

        // Mark any remaining pending sets as error if they weren't matched
        pendingSets.forEach(ps => {
            const wasMatched = savedExercise.sets!.some(savedSet => 
                savedSet &&
                Math.abs(ps.weight - savedSet.weight) < 0.01 &&
                ps.reps === savedSet.reps &&
                ps.weight_setup === (savedSet.weight_setup || "")
            );
            if (!wasMatched && ps.status === 'pending') {
                ps.status = 'error';
                ps.error = 'Failed to save set';
            }
        });

        // Update the log entry with saved exercise
        exerciseGroup.logged = savedExercise;
        return true;
    } else {
        // Mark all pending sets as error
        loggedSets.value.forEach(set => {
            if (set.status === 'pending') {
                set.status = 'error';
                set.error = 'Failed to save exercise';
            }
        });
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

    // Confirm if reps = 0
    if (currentReps.value === 0) {
        const confirmed = await dialogManager.confirm({
            title: "Confirm Zero Reps",
            message: "You're about to log a set with 0 reps. Continue?",
            confirmText: "Yes",
            cancelText: "Cancel",
        });
        if (!confirmed) {
            return;
        }
    }

    // Create new set with pending status
    const tempId = `temp-${Date.now()}-${tempIdCounter++}`;
    const newSet: LoggedSetWithStatus = {
        weight: currentWeight.value,
        reps: currentReps.value,
        weight_setup: currentWeightSetup.value,
        status: 'pending',
        id: null,
        error: null,
        tempId,
    };

    loggedSets.value.push(newSet);
    currentSetNumber.value++;
    currentReps.value = 0; // Reset reps, keep weight and weight setup

    // Save immediately after adding set
    await saveCurrentExercise();
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
        const tempId = `temp-finish-${Date.now()}`;
        loggedSets.value.push({
            weight: currentWeight.value,
            reps: currentReps.value,
            weight_setup: currentWeightSetup.value,
            status: 'pending' as const,
            id: null,
            error: null,
            tempId,
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
    currentWeightSetup.value = "";
    currentSetNumber.value = 1;
    loggedSets.value = [];
    notes.value = "";
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
): Promise<LoggedExercise | null> => {
    const rawExercise = toRaw(exercise);
    rawExercise.sets = toRaw(rawExercise.sets).filter(
        (set) => !(set.reps === 0 && set.weight === 0),
    );
    rawExercise.workout_log_id =
        data.value?.day.ID ?? rawExercise.workout_log_id;

    const { response, error } = await useAPIPost<{
        exercise: LoggedExercise;
    }>(
        `workout/exercises/log`,
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
        return null;
    }

    return response?.exercise || null;
};

const confirmLogs = async () => {
    const { response, error } = await useAPIPost<{
        all_logged: boolean;
    }>(`workout/exercises/all-logged`, "POST", {});

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
    console.log("addExerciseToWorkout", exerciseId);
    const { response, error } = await useAPIPost<{
        exercise: LoggedExercise;
    }>(`workout/exercises/add`, "POST", {
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

// Retry saving a failed set
const retrySet = async (setIndex: number) => {
    const set = loggedSets.value[setIndex];
    if (!set || set.status !== 'error') return;

    set.status = 'pending';
    set.error = null;
    await saveCurrentExercise();
};

// Delete a set
const deleteSet = async (setIndex: number) => {
    const set = loggedSets.value[setIndex];
    if (!set) return;

    // Only allow deletion of sets that have been confirmed by backend
    if (set.status !== 'success') {
        toast.push("Can only delete sets that have been saved", "error");
        return;
    }

    loggedSets.value.splice(setIndex, 1);
    currentSetNumber.value = loggedSets.value.length + 1;
    
    // Save to update backend
    await saveCurrentExercise();
};

// Edit a set (load it into current inputs and remove from logged sets)
const editSet = (setIndex: number) => {
    const set = loggedSets.value[setIndex];
    if (!set) return;

    // Only allow editing of sets that have been confirmed by backend
    if (set.status !== 'success') {
        toast.push("Can only edit sets that have been saved", "error");
        return;
    }

    // Load set values into current inputs
    currentWeight.value = set.weight;
    currentReps.value = set.reps;
    currentWeightSetup.value = set.weight_setup;

    // Remove from logged sets
    loggedSets.value.splice(setIndex, 1);
    currentSetNumber.value = loggedSets.value.length + 1;
};

// Go back to list view
const goBackToList = () => {
    currentView.value = "list";
    currentExerciseIndex.value = null;
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

        const { error } = await useAPIPost(`workout/exercises/remove`, "DELETE", {
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
            @add-exercise="addExerciseToWorkout($event)"
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
            :current-weight-setup="currentWeightSetup"
            :notes="notes"
            @update:current-weight="currentWeight = $event"
            @update:current-reps="currentReps = $event"
            @update:weight-edit-mode="weightEditMode = $event"
            @update:reps-edit-mode="repsEditMode = $event"
            @update:weight-input-value="weightInputValue = $event"
            @update:reps-input-value="repsInputValue = $event"
            @update:current-weight-setup="currentWeightSetup = $event"
            @update:notes="notes = $event"
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
            @retry-set="retrySet"
            @delete-set="deleteSet"
            @edit-set="editSet"
            @go-back="goBackToList"
        />
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    max-width: 800px;
    margin: 0 auto;
}

@media (max-width: 767px) {
    .container {
        padding: 0.5rem;
        max-width: 100%;
    }
}
</style>
