<script setup lang="ts">
import type { LoggedSetWithStatus, ExerciseGroup } from "../store/useWorkoutStore";
import { useWorkoutStore } from "../store/useWorkoutStore";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { ref, computed, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import ExerciseLoggingView from "../components/ExerciseLoggingView.vue";

const router = useRouter();
const route = useRoute();
const { log, data, logExercise, pending } = useWorkoutStore();

const exerciseId = computed(() => {
    const id = route.params.id;
    return typeof id === 'string' ? parseInt(id, 10) : Array.isArray(id) ? parseInt(id[0], 10) : id;
});

const exerciseGroup = computed<ExerciseGroup | null>(() => {
    if (!exerciseId.value || pending.value) return null;
    const index = log.value.findIndex(
        (eg) => eg.planned?.ID === exerciseId.value || eg.logged?.exercise_id === exerciseId.value
    );
    return index >= 0 ? log.value[index] : null;
});

const currentWeight = ref<number>(0);
const currentReps = ref<number>(0);
const currentWeightSetup = ref<string>("");
const currentSetNumber = ref<number>(1);

const loggedSets = ref<LoggedSetWithStatus[]>([]);
let tempIdCounter = 0;

const weightEditMode = ref<boolean>(false);
const repsEditMode = ref<boolean>(false);
const weightInputValue = ref<string>("");
const repsInputValue = ref<string>("");
const notes = ref<string>("");

const saveTimeout = ref<number | null>(null);
const isSaving = ref<boolean>(false);

const initializeExercise = () => {
    if (pending.value) return;
    
    const group = exerciseGroup.value;
    if (!group) {
        router.push('/liveworkout');
        return;
    }

    if (group.logged && group.logged.sets && Array.isArray(group.logged.sets) && group.logged.sets.length > 0) {
        loggedSets.value = group.logged.sets.map(set => ({
            weight: set.weight,
            reps: set.reps,
            weight_setup: set.weight_setup || "",
            status: 'success' as const,
            id: set.ID,
            error: null,
            tempId: `existing-${set.ID}`,
        }));
        currentSetNumber.value = loggedSets.value.length + 1;
        
        const lastSet = group.logged.sets[group.logged.sets.length - 1];
        if (lastSet) {
            currentWeight.value = lastSet.weight;
            currentReps.value = 0;
            currentWeightSetup.value = lastSet.weight_setup || "";
        } else {
            currentWeight.value = 0;
            currentReps.value = 0;
            currentWeightSetup.value = "";
        }
        
        notes.value = group.logged.notes || "";
    } else {
        loggedSets.value = [];
        currentSetNumber.value = 1;
        
        if (group.previous && group.previous.sets && Array.isArray(group.previous.sets) && group.previous.sets.length > 0) {
            const firstSet = group.previous.sets[0];
            currentWeight.value = firstSet ? firstSet.weight : 0;
            currentWeightSetup.value = firstSet ? (firstSet.weight_setup || "") : "";
        } else {
            currentWeight.value = 0;
            currentWeightSetup.value = "";
        }
        currentReps.value = 0;
        
        notes.value = group.previous?.notes || "";
    }
    
    weightEditMode.value = false;
    repsEditMode.value = false;
};

watch([() => exerciseGroup.value, () => pending.value], ([group, isPending]) => {
    if (!isPending && group) {
        initializeExercise();
    } else if (!isPending && !group) {
        router.push('/liveworkout');
    }
}, { immediate: true });

const saveCurrentExercise = async (): Promise<boolean> => {
    const group = exerciseGroup.value;
    if (!group) return false;

    const allSets = [...loggedSets.value];
    if (currentWeight.value !== 0 && currentReps.value !== 0) {
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

    if (allSets.length === 0) return false;

    let exerciseToLog: any;
    if (group.logged && group.logged.ID > 0) {
        exerciseToLog = { ...group.logged };
    } else if (group.previous) {
        exerciseToLog = {
            ...group.previous,
            ID: 0,
            workout_log_id: data.value?.day.ID ?? 0,
            sets: [],
        };
    } else {
        exerciseToLog = {
            ID: 0,
            workout_log_id: data.value?.day.ID ?? 0,
            exercise_id: group.planned.ID,
            exercise: group.planned,
            sets: [],
            notes: "",
            created_at: "",
            updated_at: "",
        };
    }

    exerciseToLog.sets = allSets.map((set) => ({
        logged_exercise_id: exerciseToLog.ID,
        reps: set.reps,
        weight: set.weight,
        weight_setup: set.weight_setup || "",
        ID: set.id || 0,
        created_at: "",
        updated_at: "",
    }));

    exerciseToLog.notes = notes.value;
    exerciseToLog.workout_log_id = data.value?.day.ID ?? exerciseToLog.workout_log_id;

    const type: "logged" | "previous" = (group.logged && group.logged.ID > 0) ? "logged" : "previous";
    const savedExercise = await logExercise(exerciseToLog, type);

    if (savedExercise && savedExercise.sets && Array.isArray(savedExercise.sets)) {
        const pendingSets = loggedSets.value.filter(s => s.status === 'pending');
        
        savedExercise.sets.forEach((savedSet) => {
            if (!savedSet) return;
            
            const pendingSet = pendingSets.find(ps => 
                Math.abs(ps.weight - savedSet.weight) < 0.01 &&
                ps.reps === savedSet.reps &&
                ps.weight_setup === (savedSet.weight_setup || "")
            );
            
            if (pendingSet) {
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
                const existingSet = loggedSets.value.find(s => s.id === savedSet.ID);
                if (existingSet) {
                    existingSet.status = 'success';
                    existingSet.error = null;
                }
            }
        });

        pendingSets.forEach(ps => {
            const wasMatched = savedExercise.sets.some(savedSet => 
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

        group.logged = savedExercise;
        return true;
    } else {
        loggedSets.value.forEach(set => {
            if (set.status === 'pending') {
                set.status = 'error';
                set.error = 'Failed to save exercise';
            }
        });
    }

    return false;
};

const debouncedSave = () => {
    if (saveTimeout.value) {
        clearTimeout(saveTimeout.value);
    }

    saveTimeout.value = setTimeout(async () => {
        if (isSaving.value) return;
        
        isSaving.value = true;
        try {
            await saveCurrentExercise();
        } catch (error) {
            console.error("Error saving exercise:", error);
            toast.push("Failed to save. Will retry on next action.", "error");
        } finally {
            isSaving.value = false;
        }
    }, 400);
};

const addNextSet = async () => {
    if (currentWeight.value === 0 && currentReps.value === 0) {
        toast.push("Please enter weight and reps", "error");
        return;
    }

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
    currentReps.value = 0;

    await saveCurrentExercise();
};

const finishLogging = async () => {
    if (currentReps.value === 0) {
        const confirmed = await dialogManager.confirm({
            title: "No Rep Set",
            message: "You're about to log a set with 0 reps. Go back to list view without logging this set?",
            confirmText: "Go back",
            cancelText: "Stay here",
        });
        if (confirmed) {
            if (saveTimeout.value) {
                clearTimeout(saveTimeout.value);
                saveTimeout.value = null;
            }
            currentWeight.value = 0;
            currentReps.value = 0;
            currentWeightSetup.value = "";
            router.push('/liveworkout');
            return;
        }
        return;
    }

    if (loggedSets.value.length === 0) {
        toast.push("Please log at least one set", "error");
        return;
    }

    if (saveTimeout.value) {
        clearTimeout(saveTimeout.value);
        saveTimeout.value = null;
    }

    const hasCurrentSet = currentWeight.value !== 0 && currentReps.value !== 0;
    const success = await saveCurrentExercise();

    if (!success && hasCurrentSet) {
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

    router.push('/liveworkout');
};

const incrementWeight = () => {
    currentWeight.value = (currentWeight.value || 0) + 2.5;
};

const decrementWeight = () => {
    currentWeight.value = Math.max(0, (currentWeight.value || 0) - 2.5);
};

const incrementReps = () => {
    currentReps.value = (currentReps.value || 0) + 1;
};

const decrementReps = () => {
    currentReps.value = Math.max(0, (currentReps.value || 0) - 1);
};

const enterWeightEditMode = () => {
    weightEditMode.value = true;
    weightInputValue.value = currentWeight.value.toString();
};

const exitWeightEditMode = () => {
    const numValue = parseFloat(weightInputValue.value);
    if (!isNaN(numValue) && numValue >= 0) {
        currentWeight.value = numValue;
    }
    weightEditMode.value = false;
};

const enterRepsEditMode = () => {
    repsEditMode.value = true;
    repsInputValue.value = currentReps.value.toString();
};

const exitRepsEditMode = () => {
    const numValue = parseInt(repsInputValue.value);
    if (!isNaN(numValue) && numValue >= 0) {
        currentReps.value = numValue;
    }
    repsEditMode.value = false;
};

const retrySet = async (setIndex: number) => {
    const set = loggedSets.value[setIndex];
    if (!set || set.status !== 'error') return;

    set.status = 'pending';
    set.error = null;
    await saveCurrentExercise();
};

const deleteSet = async (setIndex: number) => {
    const set = loggedSets.value[setIndex];
    if (!set) return;

    if (set.status !== 'success') {
        toast.push("Can only delete sets that have been saved", "error");
        return;
    }

    loggedSets.value.splice(setIndex, 1);
    currentSetNumber.value = loggedSets.value.length + 1;
    
    await saveCurrentExercise();
};

const editSet = (setIndex: number) => {
    const set = loggedSets.value[setIndex];
    if (!set) return;

    if (set.status !== 'success') {
        toast.push("Can only edit sets that have been saved", "error");
        return;
    }

    currentWeight.value = set.weight;
    currentReps.value = set.reps;
    currentWeightSetup.value = set.weight_setup;

    loggedSets.value.splice(setIndex, 1);
    currentSetNumber.value = loggedSets.value.length + 1;
};

const goBackToList = () => {
    router.push('/liveworkout');
};
</script>

<template>
    <div v-if="pending" class="container">
        <div>Loading...</div>
    </div>
    <div v-else-if="!exerciseGroup" class="container">
        <div>Exercise not found</div>
    </div>
    <div v-else class="container">
        <ExerciseLoggingView
            :exercise-group="exerciseGroup"
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

