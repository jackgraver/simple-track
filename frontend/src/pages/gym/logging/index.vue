<script setup lang="ts">
import { useWorkoutStore } from "./store/useWorkoutStore";
import ExerciseListView from "./components/ExerciseListView.vue";
import { useRoute, useRouter } from "vue-router";
import { computed, ref } from "vue";
import { ErrorState, SkeletonList, useDelayedFlag } from "~/shared/ui";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";
import { useLoggingRouteContext } from "./composables/useLoggingRouteContext";
import { buildWorkoutSteps, type WorkoutStep } from "./domain/workoutSteps";

const router = useRouter();
const route = useRoute();
const { offset } = useLoggingRouteContext();
const {
    log,
    plannedCardio,
    loggedCardio,
    plannedPreMobility,
    loggedPreMobility,
    plannedPostMobility,
    loggedPostMobility,
    pending,
    error,
    data,
    addExerciseToWorkout,
    removeExerciseFromLog,
    refetch,
} = useWorkoutStore(offset);

// Fast loads show nothing instead of flashing a skeleton.
const showSkeleton = useDelayedFlag(() => pending.value);
const removingId = ref<number | null>(null);

const workoutName = computed(() => data.value?.day?.workout_plan?.name || "");

const steps = computed(() =>
    buildWorkoutSteps({
        log: log.value,
        plannedCardio: plannedCardio.value,
        loggedCardio: loggedCardio.value,
        plannedPreMobility: plannedPreMobility.value,
        loggedPreMobility: loggedPreMobility.value,
        plannedPostMobility: plannedPostMobility.value,
        loggedPostMobility: loggedPostMobility.value,
    }),
);

const dateLabel = computed(() => {
    if (offset.value === 0) return "Today";
    if (offset.value === 1) return "Yesterday";
    const d = new Date();
    d.setDate(d.getDate() - offset.value);
    return d.toLocaleDateString(undefined, { weekday: "long", month: "short", day: "numeric" });
});

const selectStep = (step: WorkoutStep) => {
    if (step.kind === "lift" && step.exerciseIndex != null) selectExercise(step.exerciseIndex);
    else if (step.kind === "cardio") selectCardio();
    else if (step.kind === "warmup") selectPreMobility();
    else if (step.kind === "cooldown") selectPostMobility();
};

const finish = () => {
    router.push({ name: "gym", query: route.query });
};

const selectExercise = (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;

    const exerciseId =
        exerciseGroup.planned?.ID || exerciseGroup.logged?.exercise_id;
    if (!exerciseId) return;

    router.push({
        name: "logging-exercise",
        params: { id: String(exerciseId) },
        query: route.query,
    });
};

const handleAddExercise = async (exerciseId: number) => {
    try {
        await addExerciseToWorkout(exerciseId);
        toast.push("Exercise added", "success");
    } catch (err: unknown) {
        console.error(err);
        toast.push(getApiErrorMessage(err, "Failed to add exercise"), "error");
    }
};

const selectCardio = () => {
    router.push({
        name: "logging-cardio",
        query: route.query,
    });
};

const selectPreMobility = () => {
    router.push({
        name: "logging-mobility",
        params: { slot: "pre" },
        query: route.query,
    });
};

const selectPostMobility = () => {
    router.push({
        name: "logging-mobility",
        params: { slot: "post" },
        query: route.query,
    });
};

const handleRemoveExercise = async (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;

    const exerciseId =
        exerciseGroup.logged?.exercise_id || exerciseGroup.planned?.ID;
    if (!exerciseId) {
        toast.push("Cannot remove exercise: ID not found", "error");
        return;
    }

    removingId.value = exerciseId;
    try {
        await removeExerciseFromLog(exerciseId);
        toast.push("Exercise removed", "success");
    } catch (err: unknown) {
        console.error(err);
        toast.push(getApiErrorMessage(err, "Failed to remove exercise"), "error");
    } finally {
        removingId.value = null;
    }
};
</script>

<template>
    <div class="mx-auto flex w-full max-w-md flex-col gap-4 py-2 lg:py-0">
        <ErrorState
            v-if="error"
            :error="error"
            title="Couldn't load today's workout"
            @retry="refetch()"
        />
        <SkeletonList v-else-if="pending && showSkeleton" :rows="5" :leading="false" />
        <ExerciseListView
            v-else-if="!pending"
            :workout-name="workoutName"
            :date-label="dateLabel"
            :steps="steps"
            :exercises="log"
            :removing-id="removingId"
            @select-step="selectStep"
            @add-exercise="handleAddExercise"
            @remove-exercise="handleRemoveExercise"
            @finish="finish"
        />
    </div>
</template>
