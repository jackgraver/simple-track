<script setup lang="ts">
import { useWorkoutStore } from "../store/useWorkoutStore";
import { computed } from "vue";
import { useRoute } from "vue-router";
import ExerciseLoggingView from "./components/ExerciseLoggingView.vue";
import CardioLoggingView from "./components/CardioLoggingView.vue";
import MobilityLoggingView from "./components/MobilityLoggingView.vue";
import LoggingPageShell from "./components/LoggingPageShell.vue";
import { useLoggingRouteContext } from "../composables/useLoggingRouteContext";
import {
    findExerciseGroupByExerciseId,
    parseExerciseIdParam,
} from "./domain/exerciseRouteGroup";

const route = useRoute();
const { goBackToLogging, offset } = useLoggingRouteContext();

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
    if (loggingKind.value === "exercise") return "Exercise not found";
    if (loggingKind.value === "mobility") return "Mobility not found";
    return "Logging item not found";
});

const isEmpty = computed(() => {
    if (pending.value) return false;
    if (loggingKind.value === "exercise") return !exerciseGroupForShell.value;
    if (loggingKind.value === "mobility") return !mobilityLoggedForRoute.value;
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
            class="flex flex-col gap-4 w-full self-stretch md:p-0 md:px-2 md:py-0"
        >
            <ExerciseLoggingView v-if="loggingKind === 'exercise'" />
            <CardioLoggingView
                v-else-if="loggingKind === 'cardio'"
                :planned-cardio="plannedCardio"
                :logged-cardio="loggedCardio"
            />
            <MobilityLoggingView
                v-else-if="loggingKind === 'mobility'"
                :logged-pre-mobility="loggedPreMobility"
                :logged-post-mobility="loggedPostMobility"
                :save-pre-mobility="savePreMobility"
                :save-post-mobility="savePostMobility"
            />
        </div>
    </LoggingPageShell>
</template>
