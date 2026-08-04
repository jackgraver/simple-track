<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import {
    findExerciseGroupByExerciseId,
    parseExerciseIdParam,
} from "../domain/exerciseRouteGroup";
import {
    transitionExerciseLoggingFlow,
    type ExerciseLoggingFlowStep,
} from "../domain/exerciseLoggingFlow";
import { pickBestLoggedSet } from "../domain/loggedSetDefaults";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";
import ExerciseSetupScreen from "./ExerciseSetupScreen.vue";
import ExerciseRepsScreen from "./ExerciseRepsScreen.vue";
import ExerciseRestScreen from "./ExerciseRestScreen.vue";

const route = useRoute();
const router = useRouter();
const { offset } = useLoggingRouteContext();
const { log, data, pending, logExercise, deleteLoggedSet } = useWorkoutStore(offset);

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
const previousBestSet = computed(() =>
    pickBestLoggedSet(session.exerciseGroup?.previous?.sets ?? []),
);
const previousReps = computed(() => previousBestSet.value?.reps ?? null);
const cues = computed(
    () =>
        session.exerciseGroup?.planned?.cues ??
        session.exerciseGroup?.logged?.exercise?.cues ??
        "",
);

const transition = (
    action: "startLogging" | "setLogged" | "nextSet",
) => {
    step.value = transitionExerciseLoggingFlow(step.value, action);
};

const finish = async () => {
    await session.finishLogging();
};

watch(exerciseId, () => {
    step.value = "setup";
});
</script>

<template>
    <ExerciseSetupScreen
        v-if="step === 'setup'"
        :exercise-name="exerciseName"
        :expected-weight="session.currentWeight"
        :previous-reps="previousReps"
        :notes="session.notes"
        :cues="cues"
        :weight-increase="session.weightProgression?.increase ?? null"
        machine-setup-note="Machine setup notes are not saved yet. Add pin, seat, or handle settings here for now."
        @back="session.goBackToList()"
        @continue="transition('startLogging')"
        @undo-weight-increase="session.undoWeightProgression()"
    />
    <ExerciseRepsScreen
        v-else-if="step === 'reps'"
        :exercise-name="exerciseName"
        :session="session"
        :previous-reps="previousReps"
        @back="session.goBackToList()"
        @logged="transition('setLogged')"
    />
    <ExerciseRestScreen
        v-else
        :exercise-name="exerciseName"
        :set-number="session.currentSetNumber"
        :session="session"
        @back="session.goBackToList()"
        @next="transition('nextSet')"
        @finish="finish"
    />
</template>
