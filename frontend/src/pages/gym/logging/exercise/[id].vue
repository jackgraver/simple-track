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

const exerciseNotInDay = computed(
    () => loggingKind.value === "exercise" && !pending.value && !error.value && !exerciseGroupForShell.value,
);

watch(
    exerciseNotInDay,
    (notInDay) => {
        if (notInDay) router.replace({ name: "gym" });
    },
    { immediate: true },
);

const isMobilityMissing = computed(
    () => !pending.value && loggingKind.value === "mobility" && !mobilityLoggedForRoute.value,
);
</script>

<template>
    <template v-if="exerciseNotInDay" />
    <LoggingPageShell
        v-else-if="loggingKind === 'exercise'"
        :pending="pending"
        :error="error"
        :empty="isMobilityMissing"
        empty-message="Mobility not found"
        @back="goBackToLogging"
        @retry="refetch()"
    >
        <ExerciseLoggingFlow />
    </LoggingPageShell>
    <LoggingPageShell
        v-else
        :pending="pending"
        :error="error"
        :empty="isMobilityMissing"
        empty-message="Mobility not found"
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
