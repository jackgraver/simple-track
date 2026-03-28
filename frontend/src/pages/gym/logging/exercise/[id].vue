<script setup lang="ts">
import type { ExerciseGroup } from "../store/useWorkoutStore";
import { useWorkoutStore } from "../store/useWorkoutStore";
import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import ExerciseLoggingView from "../components/ExerciseLoggingView.vue";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";

const router = useRouter();
const route = useRoute();
const { log, data, logExercise, pending } = useWorkoutStore(0);

const exerciseId = computed(() => {
    const id = route.params.id;
    return typeof id === "string"
        ? parseInt(id, 10)
        : Array.isArray(id)
          ? parseInt(id[0], 10)
          : id;
});

const exerciseGroup = computed<ExerciseGroup | null>(() => {
    if (!exerciseId.value || pending.value) return null;
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
    logExercise,
    router,
});
</script>

<template>
    <div v-if="pending" class="container">
        <div>Loading...</div>
    </div>
    <div v-else-if="!exerciseGroup" class="container">
        <div>Exercise not found</div>
    </div>
    <div v-else class="container">
        <ExerciseLoggingView :session="session" />
    </div>
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
