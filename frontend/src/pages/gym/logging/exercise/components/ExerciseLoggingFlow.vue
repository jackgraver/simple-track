<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { onBeforeRouteLeave, useRoute, useRouter } from "vue-router";
import { dialogManager } from "~/composables/dialog/useDialog";
import {
    DEFAULT_EXERCISE_LOAD_TYPE,
    type ExerciseLoadType,
} from "~/types/workout";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import {
    findExerciseGroupByExerciseId,
    parseExerciseIdParam,
} from "../domain/exerciseRouteGroup";
import {
    transitionExerciseLoggingFlow,
    type ExerciseLoggingFlowAction,
    type ExerciseLoggingFlowStep,
} from "../domain/exerciseLoggingFlow";
import {
    formatExerciseWeightSetup,
    isPlateLoadedExercise,
    parseWeightSetup,
    type WeightSetupDialogResult,
} from "../domain/weightSetup";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";
import WeightSetupDialog from "../dialog/WeightSetupDialog.vue";
import ExerciseSetupScreen from "./ExerciseSetupScreen.vue";
import ExerciseRepsScreen from "./ExerciseRepsScreen.vue";
import ExerciseRestScreen from "./ExerciseRestScreen.vue";

const route = useRoute();
const router = useRouter();
const { offset } = useLoggingRouteContext();
const { log, data, pending, logExercise, deleteLoggedSet } =
    useWorkoutStore(offset);

const exerciseId = computed(() => parseExerciseIdParam(route));
const exerciseGroup = computed(() => {
    if (exerciseId.value == null || pending.value) return null;
    return findExerciseGroupByExerciseId(log.value, exerciseId.value);
});
const dayId = computed(() => data.value?.day.ID ?? 0);
const session = useExerciseLoggingSession({
    exerciseGroup,
    pending,
    dayId,
    offset,
    logExercise,
    deleteLoggedSet,
    router,
});

const step = ref<ExerciseLoggingFlowStep>("setup");
const exerciseName = computed(
    () =>
        session.exerciseGroup?.planned?.name ??
        session.exerciseGroup?.logged?.exercise?.name ??
        "",
);
const previousSets = computed(() =>
    (session.exerciseGroup?.previous?.sets ?? []).map((set) => ({
        weight: set.weight,
        reps: set.reps,
    })),
);
const previousReps = computed(() => previousSets.value.map((set) => set.reps));
const previousRepForCurrentSet = computed(
    () => previousReps.value[session.currentSetNumber - 1] ?? null,
);
const cues = computed(
    () =>
        session.exerciseGroup?.planned?.cues ??
        session.exerciseGroup?.logged?.exercise?.cues ??
        "",
);
const exerciseLoadType = computed<ExerciseLoadType>(
    () =>
        (session.exerciseGroup?.planned?.load_type ??
            session.exerciseGroup?.logged?.exercise?.load_type ??
            DEFAULT_EXERCISE_LOAD_TYPE) as ExerciseLoadType,
);
const tracksWeightSetup = computed(() =>
    isPlateLoadedExercise(exerciseLoadType.value),
);
const weightSetupSummary = computed(() => {
    const setup = session.currentWeightSetup.trim();
    if (!setup) return "Set plates per side";
    const parsed = parseWeightSetup(setup);
    if (!parsed.parsed) return setup;
    const formatted = formatExerciseWeightSetup(
        exerciseLoadType.value,
        parsed.platesPerSide,
    );
    return formatted || "No plates";
});
const repRollover = computed(() => {
    const g = session.exerciseGroup;
    const rollover =
        g?.planned?.rep_rollover ??
        g?.logged?.exercise?.rep_rollover ??
        g?.previous?.exercise?.rep_rollover;
    return typeof rollover === "number" && Number.isFinite(rollover)
        ? rollover
        : undefined;
});
const continueLabel = computed(() =>
    session.loggedSets.length > 0 ? "Continue logging" : "Start logging",
);
const flowStorageKey = computed(
    () => `gym-flow:v1:day:${dayId.value}:exercise:${exerciseId.value ?? 0}`,
);

const transition = (action: ExerciseLoggingFlowAction) => {
    step.value = transitionExerciseLoggingFlow(step.value, action);
};

const clearFlowState = () => {
    window.sessionStorage.removeItem(flowStorageKey.value);
};

const goBackToList = () => {
    clearFlowState();
    session.goBackToList();
};

const goBackToSetup = () => {
    if (step.value === "reps") session.cancelEditingSet();
    transition("backToSetup");
};

const finish = async () => {
    if (await session.finishLogging()) {
        clearFlowState();
    }
};

const editSet = (setIndex: number) => {
    session.editSet(setIndex);
    if (session.editingSetIndex != null) step.value = "reps";
};

const openWeightSetupDialog = async () => {
    if (!tracksWeightSetup.value) return;
    const result = await dialogManager.custom<WeightSetupDialogResult>({
        title: "Weight setup",
        component: WeightSetupDialog,
        componentProps: {
            currentSetup: session.currentWeightSetup,
            targetWeightLbs: session.currentWeight,
            loadType: exerciseLoadType.value,
        },
        wide: true,
    });
    if (result == null) return;
    session.updateWeightSetup(result.weightSetup);
    if (result.syncWeight) session.commitWeightFromInput(result.totalLbs);
};

watch(
    flowStorageKey,
    (key, previousKey) => {
        if (previousKey) window.sessionStorage.removeItem(previousKey);
        step.value = "setup";
        const stored = window.sessionStorage.getItem(key);
        if (stored === "setup" || stored === "reps" || stored === "rest") {
            step.value = stored;
        }
    },
    { immediate: true },
);

watch(step, (value) => {
    if (exerciseId.value != null && dayId.value > 0) {
        window.sessionStorage.setItem(flowStorageKey.value, value);
    }
});

onBeforeRouteLeave(clearFlowState);
</script>

<template>
    <ExerciseSetupScreen
        v-if="step === 'setup'"
        :exercise-name="exerciseName"
        :weight="session.currentWeight"
        :weight-step="session.weightIncrement"
        :previous-reps="previousReps"
        :notes="session.notes"
        :cues="cues"
        :weight-increase="session.weightProgression?.increase ?? null"
        :tracks-weight-setup="tracksWeightSetup"
        :weight-setup-summary="weightSetupSummary"
        :continue-label="continueLabel"
        @back="goBackToList"
        @continue="transition('startLogging')"
        @undo-weight-increase="session.undoWeightProgression()"
        @edit-weight-setup="openWeightSetupDialog()"
        @update:weight="session.commitWeightFromInput"
    />
    <ExerciseRepsScreen
        v-else-if="step === 'reps'"
        :exercise-name="exerciseName"
        :session="session"
        :previous-rep="previousRepForCurrentSet"
        :rep-rollover="repRollover"
        @back="goBackToSetup"
        @logged="transition('setLogged')"
    />
    <ExerciseRestScreen
        v-else
        :exercise-name="exerciseName"
        :set-number="session.currentSetNumber"
        :current-weight="session.currentWeight"
        :session="session"
        @back="goBackToSetup"
        @next="transition('nextSet')"
        @finish="finish"
        @edit="editSet"
        @edit-setup="transition('backToSetup')"
    />
</template>
