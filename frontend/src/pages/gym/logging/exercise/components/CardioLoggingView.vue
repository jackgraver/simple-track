<script setup lang="ts">
import type { Cardio, PlannedCardio } from "~/types/workout";
import LoggingHeader from "./LoggingHeader.vue";
import { ref, watch, computed } from "vue";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import { Badge, Button, Field, Textarea } from "~/shared/ui";
import NumberSlider from "./NumberSlider.vue";

const props = defineProps<{
    plannedCardio: PlannedCardio | null;
    loggedCardio: Cardio | null;
}>();
const { offset, goBackToLogging } = useLoggingRouteContext();
const { saveCardio } = useWorkoutStore(offset);

const cardioName = computed(
    () => props.plannedCardio?.type ?? props.loggedCardio?.type ?? "Cardio",
);

const isLogged = computed(() => (props.loggedCardio?.minutes ?? 0) > 0);
const sliderMax = computed(() =>
    Math.max(
        100,
        props.plannedCardio?.minutes ?? 0,
        props.loggedCardio?.minutes ?? 0,
    ),
);

const currentMinutes = ref(25);
const notes = ref("");

const error = ref("");

watch(
    () => [props.loggedCardio, props.plannedCardio] as const,
    ([loggedCardio, plannedCardio]) => {
        currentMinutes.value = loggedCardio?.minutes ?? plannedCardio?.minutes ?? 0;
        notes.value = loggedCardio?.notes ?? "";
    },
    { immediate: true },
);

const onMinutesUpdate = (v: number) => {
    currentMinutes.value = v;
    error.value = "";
};

const finish = () => {
    if (currentMinutes.value <= 0) {
        error.value = "Enter minutes before saving.";
        return;
    }
    // Shows as logged at once; a failed save rolls back with a Retry toast.
    saveCardio(currentMinutes.value, undefined, notes.value);
    goBackToLogging();
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="cardioName" @back="goBackToLogging">
            <template #right>
                <Badge v-if="isLogged" tone="success">Logged</Badge>
            </template>
        </LoggingHeader>
        <NumberSlider
            label="Time (minutes)"
            :model-value="currentMinutes"
            :min="0"
            :max="sliderMax"
            :step="1"
            integer-only
            :error="error"
            @update:model-value="onMinutesUpdate"
        />
        <Field label="Notes">
            <Textarea v-model="notes" placeholder="What did you watch? Any thoughts..." :rows="3" />
        </Field>
        <Button variant="primary" size="lg" block @click="finish">Finish cardio</Button>
    </div>
</template>
