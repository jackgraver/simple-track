<script setup lang="ts">
import type { ExerciseGroup } from "../store/useWorkoutStore";
import type { MobilityLogged } from "~/types/workout";
import { useWorkoutStore } from "../store/useWorkoutStore";
import { computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import ExerciseLoggingView from "./components/ExerciseLoggingView.vue";
import CardioLoggingView from "./components/CardioLoggingView.vue";
import MobilityLoggingView from "./components/MobilityLoggingView.vue";
import LoggingPageShell from "./components/LoggingPageShell.vue";
import { useExerciseLoggingSession } from "./composables/useExerciseLoggingSession";
import { useLoggingRouteContext } from "../composables/useLoggingRouteContext";

const router = useRouter();
const route = useRoute();
const { goBackToLogging, offset } = useLoggingRouteContext();

const {
    log,
    data,
    logExercise,
    deleteLoggedSet,
    pending,
    error,
    plannedCardio,
    loggedCardio,
    loggedPreMobility,
    loggedPostMobility,
} = useWorkoutStore(offset);

const mobilitySlot = computed<"pre" | "post" | null>(() => {
    const raw = route.params.slot;
    const value = typeof raw === "string" ? raw : "";
    if (value === "pre" || value === "post") return value;
    return null;
});

const loggingKind = computed<"exercise" | "cardio" | "mobility" | "unknown">(
    () => {
        if (route.name === "logging-exercise") return "exercise";
        if (route.name === "logging-cardio") return "cardio";
        if (route.name === "logging-mobility" && mobilitySlot.value)
            return "mobility";
        return "unknown";
    },
);

watch(
    () => loggingKind.value,
    (kind) => {
        if (kind === "unknown") {
            goBackToLogging();
        }
    },
    { immediate: true },
);

const exerciseId = computed(() => {
    if (loggingKind.value !== "exercise") return null;
    const id = route.params.id;
    const parsed =
        typeof id === "string"
            ? Number.parseInt(id, 10)
            : Array.isArray(id)
              ? Number.parseInt(id[0], 10)
              : id;
    return Number.isFinite(parsed) ? parsed : null;
});

const exerciseGroup = computed<ExerciseGroup | null>(() => {
    if (
        loggingKind.value !== "exercise" ||
        !exerciseId.value ||
        pending.value
    ) {
        return null;
    }
    const index = log.value.findIndex(
        (eg) =>
            eg.planned?.ID === exerciseId.value ||
            eg.logged?.exercise_id === exerciseId.value,
    );
    return index >= 0 ? log.value[index] : null;
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
    enabled: computed(() => loggingKind.value === "exercise"),
});

const loggedMobility = computed<MobilityLogged | null>(() => {
    if (loggingKind.value !== "mobility") return null;
    return mobilitySlot.value === "post"
        ? loggedPostMobility.value
        : loggedPreMobility.value;
});

const emptyMessage = computed(() => {
    if (loggingKind.value === "exercise") return "Exercise not found";
    if (loggingKind.value === "mobility") {
        return mobilitySlot.value === "post"
            ? "No post-workout mobility for this day."
            : "No pre-workout mobility for this day.";
    }
    return "Logging item not found";
});

const isEmpty = computed(() => {
    if (pending.value) return false;
    if (loggingKind.value === "exercise") return !exerciseGroup.value;
    if (loggingKind.value === "mobility") return !loggedMobility.value;
    return false;
});
</script>

<template>
    <LoggingPageShell
        :pending="pending"
        :error="error"
        :empty="isEmpty"
        :empty-message="emptyMessage"
        @back="goBackToLogging"
    >
        <div
            v-if="loggingKind === 'exercise' && exerciseGroup"
            class="container"
        >
            <ExerciseLoggingView :session="session" />
        </div>
        <div v-else-if="loggingKind === 'cardio'" class="container">
            <CardioLoggingView
                :planned-cardio="plannedCardio"
                :logged-cardio="loggedCardio"
            />
        </div>
        <div
            v-else-if="loggingKind === 'mobility' && loggedMobility"
            class="container"
        >
            <MobilityLoggingView
                :logged-mobility="loggedMobility"
                :slot="mobilitySlot ?? 'pre'"
            />
        </div>
    </LoggingPageShell>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    align-self: stretch;
}

@media (max-width: 767px) {
    .container {
        padding: 0.5rem 0;
    }
}
</style>
