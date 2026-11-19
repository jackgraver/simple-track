<script setup lang="ts">
import { computed } from "vue";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import { Button, Field, Textarea } from "~/shared/ui";
import LoggingHeader from "./LoggingHeader.vue";
import NumberSlider from "./NumberSlider.vue";

const props = defineProps<{
    exerciseName: string;
    session: ExerciseLoggingSessionViewModel;
    previousRep: number | null;
    repRollover: number | undefined;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "logged", setNumber: number): void;
}>();

const { displayText, isActive } = useGlobalRestTimer();

const sliderMarkers = computed(() => {
    const markers: {
        value: number;
        label: string;
        color?: "orange";
        labelPosition?: "bottom";
    }[] = [];
    if (props.previousRep != null && Number.isFinite(props.previousRep)) {
        markers.push({
            value: props.previousRep,
            label: "Last time",
            color: "orange",
            labelPosition: "bottom",
        });
    }
    if (props.repRollover != null && Number.isFinite(props.repRollover)) {
        markers.push({ value: props.repRollover, label: "Rollover" });
    }
    return markers;
});

const logSet = async () => {
    const setNumber = props.session.currentSetNumber;
    props.session.addNextSet();
    emit("logged", setNumber);
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" :subtitle="`Set ${session.currentSetNumber}`" @back="emit('back')">
            <template #right>
                <span class="font-medium tabular-nums text-fg">{{ isActive ? displayText : "0:00" }}</span>
            </template>
        </LoggingHeader>
        <div class="flex flex-col gap-5">
            <NumberSlider
                label="Reps"
                :model-value="session.currentReps"
                :markers="sliderMarkers"
                integer-only
                :max="50"
                @update:model-value="session.commitRepsFromInput"
            />
            <Field label="Notes">
                <Textarea
                    :model-value="session.notes"
                    placeholder="Add any notes about this exercise..."
                    :rows="3"
                    @update:model-value="session.updateNotes"
                />
            </Field>
        </div>
        <Button variant="primary" size="lg" block @click="logSet">Log set</Button>
    </div>
</template>
