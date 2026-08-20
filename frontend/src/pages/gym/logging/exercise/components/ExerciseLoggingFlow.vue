<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
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
    parseStoredExerciseLoggingFlow,
    transitionExerciseLoggingFlow,
    type ExerciseLoggingFlowAction,
    type ExerciseLoggingFlowStep,
} from "../domain/exerciseLoggingFlow";
import { previousRepsForSetAtWeight } from "../domain/loggedSetDefaults";
import {
    formatExerciseWeightSetup,
    isPlateLoadedExercise,
    parseWeightSetup,
    type WeightSetupDialogResult,
} from "../domain/weightSetup";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";
import WeightSetupDialog from "../dialog/WeightSetupDialog.vue";
import ExerciseProgressionChartPanel from "~/pages/gym/progression/ExerciseProgressionChartPanel.vue";
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
const restSetNumber = ref<number | null>(null);
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
const previousRepForCurrentSet = computed(() =>
    previousRepsForSetAtWeight(
        previousSets.value,
        session.currentSetNumber - 1,
        session.currentWeight,
    ),
);
const displayedSetNumber = computed(() =>
    step.value === "rest"
        ? (restSetNumber.value ?? Math.max(1, session.currentSetNumber - 1))
        : session.currentSetNumber,
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
const weightSetupPlaceholder = computed(() =>
    exerciseLoadType.value === "plate_loaded_total"
        ? "Set total plates"
        : "Set plates per side",
);
const weightSetupSummary = computed(() => {
    const setup = session.currentWeightSetup.trim();
    if (!setup) return weightSetupPlaceholder.value;
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
const navigationSteps: {
    id: ExerciseLoggingFlowStep;
    label: string;
}[] = [
    { id: "setup", label: "Setup" },
    { id: "reps", label: "Log" },
    { id: "rest", label: "Rest" },
];
const flowStorageKey = computed(
    () => `gym-flow:v1:day:${dayId.value}:exercise:${exerciseId.value ?? 0}`,
);

const transition = (
    action: ExerciseLoggingFlowAction,
    setNumber: number | null = null,
) => {
    if (action === "setLogged") restSetNumber.value = setNumber;
    if (action === "nextSet") restSetNumber.value = null;
    step.value = transitionExerciseLoggingFlow(step.value, action);
};

const clearFlowState = () => {
    window.sessionStorage.removeItem(flowStorageKey.value);
};

const goBackToList = () => {
    session.goBackToList();
};

const selectStep = (nextStep: ExerciseLoggingFlowStep) => {
    if (nextStep === step.value) return;
    if (step.value === "reps") session.cancelEditingSet();
    if (nextStep === "rest") {
        restSetNumber.value = Math.max(1, session.currentSetNumber - 1);
    } else {
        restSetNumber.value = null;
    }
    step.value = nextStep;
};

const finish = async () => {
    if (await session.finishLogging()) {
        clearFlowState();
    }
};

const editSet = (setIndex: number) => {
    session.editSet(setIndex);
    if (session.editingSetIndex != null) {
        restSetNumber.value = null;
        step.value = "reps";
    }
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

const openProgressionDialog = () => {
    const id = exerciseId.value;
    if (id == null) return;
    void dialogManager.custom<void>({
        title: `${exerciseName.value} Progression`,
        component: ExerciseProgressionChartPanel,
        componentProps: {
            exerciseId: id,
            lastTimeSets: previousSets.value,
            initialRange: "6m",
        },
        wide: true,
    });
};

watch(
    flowStorageKey,
    (key) => {
        const stored = parseStoredExerciseLoggingFlow(
            window.sessionStorage.getItem(key),
        );
        step.value = stored.step;
        restSetNumber.value = stored.setNumber;
    },
    { immediate: true },
);

watch([step, restSetNumber], ([currentStep, setNumber]) => {
    if (exerciseId.value == null || dayId.value <= 0) return;
    window.sessionStorage.setItem(
        flowStorageKey.value,
        JSON.stringify({ step: currentStep, setNumber }),
    );
});
</script>

<template>
    <div class="pb-[calc(5.5rem_+_env(safe-area-inset-bottom))]">
        <ExerciseSetupScreen
            v-if="step === 'setup'"
            :exercise-name="exerciseName"
            :set-number="displayedSetNumber"
            :weight="session.currentWeight"
            :weight-step="session.weightIncrement"
            :previous-sets="previousSets"
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
            @back="goBackToList"
            @logged="(setNumber) => transition('setLogged', setNumber)"
        />
        <ExerciseRestScreen
            v-else
            :exercise-name="exerciseName"
            :set-number="displayedSetNumber"
            :current-weight="session.currentWeight"
            :session="session"
            @back="goBackToList"
            @next="transition('nextSet')"
            @finish="finish"
            @progression="openProgressionDialog"
            @edit="editSet"
            @edit-setup="transition('backToSetup')"
        />
    </div>
    <nav
        aria-label="Exercise logging steps"
        class="fixed inset-x-0 bottom-0 z-auto border-t border-zinc-800 bg-mainBg pb-[env(safe-area-inset-bottom)]"
    >
        <div class="flex w-full border-x border-zinc-800 md:mx-auto md:max-w-md">
            <button
                v-for="navigationStep in navigationSteps"
                :key="navigationStep.id"
                type="button"
                class="m-0 flex-1 border-b-2 border-r border-zinc-800 border-b-transparent px-3 py-3 text-xs font-medium transition-colors last:border-r-0"
                :class="
                    navigationStep.id === step
                        ? 'border-b-zinc-100 text-zinc-100'
                        : 'text-zinc-500 hover:text-zinc-300'
                "
                :aria-current="
                    navigationStep.id === step ? 'step' : undefined
                "
                @click="selectStep(navigationStep.id)"
            >
                {{ navigationStep.label }}
            </button>
        </div>
    </nav>
</template>
