<script setup lang="ts">
import { useWorkoutStore } from "../store/useWorkoutStore";
import { computed, watch } from "vue";
import { useRoute } from "vue-router";
import ExerciseLoggingFlow from "./components/ExerciseLoggingFlow.vue";
import CardioLoggingView from "./components/CardioLoggingView.vue";
import MobilityLoggingView from "./components/MobilityLoggingView.vue";
import LoggingPageShell from "./components/LoggingPageShell.vue";
import { useLoggingRouteContext } from "../composables/useLoggingRouteContext";
import {
    findExerciseGroupByExerciseId,
    parseExerciseIdParam,
} from "./domain/exerciseRouteGroup";

const route = useRoute();
const { goBackToLogging, offset, router } = useLoggingRouteContext();

const {
    log,
    pending,
    error,
    plannedCardio,
    loggedCardio,
    loggedPreMobility,
    loggedPostMobility,
    savePreMobility,
    savePostMobility,
    refetch,
} = useWorkoutStore(offset);

const loggingKind = computed<"exercise" | "cardio" | "mobility">(() => {
    switch (route.name) {
        case "logging-exercise":
            return "exercise";
        case "logging-cardio":
            return "cardio";
        case "logging-mobility":
            return "mobility";
        default:
            throw new Error(
                `Unexpected route for logging page: ${String(route.name)}`,
            );
    }
});

const exerciseGroupForShell = computed(() => {
    if (loggingKind.value !== "exercise") return null;
    const id = parseExerciseIdParam(route);
    if (id == null || pending.value) return null;
    return findExerciseGroupByExerciseId(log.value, id);
});

const mobilityLoggedForRoute = computed(() => {
    if (loggingKind.value !== "mobility") return null;
    const raw = route.params.slot;
    const slot = typeof raw === "string" ? raw : "";
    if (slot === "post") return loggedPostMobility.value;
    if (slot === "pre") return loggedPreMobility.value;
    return null;
});

const emptyMessage = computed(() => {
    if (loggingKind.value === "mobility") return "Mobility not found";
    return "Logging item not found";
});

// An exercise that isn't in the day's plan or log (stale URL, resumed on a later day) has nothing to
// show: leave for the gym home, rendering nothing meanwhile so the empty state never flashes.
const staleExercise = computed(
    () => loggingKind.value === "exercise" && !pending.value && !error.value && !exerciseGroupForShell.value,
);
watch(
    staleExercise,
    (stale) => {
        if (stale) router.replace({ name: "gym" });
    },
    { immediate: true },
);

const isEmpty = computed(() => {
    if (pending.value) return false;
    if (loggingKind.value === "mobility") return !mobilityLoggedForRoute.value;
    return false;
});
</script>

<template>
    <template v-if="staleExercise" />
    <LoggingPageShell
        v-else-if="loggingKind === 'exercise'"
        :pending="pending"
        :error="error"
        :empty="isEmpty"
        :empty-message="emptyMessage"
        @back="goBackToLogging"
        @retry="refetch()"
    >
        <ExerciseLoggingFlow />
    </LoggingPageShell>
    <LoggingPageShell
        v-else
        :pending="pending"
        :error="error"
        :empty="isEmpty"
        :empty-message="emptyMessage"
        @back="goBackToLogging"
        @retry="refetch()"
    >
        <div class="flex w-full flex-col gap-4 self-stretch">
            <CardioLoggingView
                v-if="loggingKind === 'cardio'"
                :planned-cardio="plannedCardio"
                :logged-cardio="loggedCardio"
            />
            <MobilityLoggingView
                v-else
                :logged-pre-mobility="loggedPreMobility"
                :logged-post-mobility="loggedPostMobility"
                :save-pre-mobility="savePreMobility"
                :save-post-mobility="savePostMobility"
            />
        </div>
    </LoggingPageShell>
</template>
