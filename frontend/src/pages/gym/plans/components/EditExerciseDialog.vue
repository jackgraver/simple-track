<script setup lang="ts">
import { workoutKeys } from "~/api/workout/keys";
import { searchListKeys } from "~/shared/searchListKeys";
import { getApiErrorMessage } from "~/api/httpError";
import type { Exercise, ExerciseLoadType } from "~/types/workout";
import { DEFAULT_EXERCISE_LOAD_TYPE } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import { Button, Field, Input, NumberField, Select, Textarea } from "~/shared/ui";
import { ref, watch } from "vue";

const props = defineProps<{
    exercise: Exercise;
    onResolve?: (ok: boolean) => void;
}>();

const queryClient = useQueryClient();
const name = ref("");
const repRollover = ref<number | null>(10);
const cues = ref("");
const loadType = ref<ExerciseLoadType>(DEFAULT_EXERCISE_LOAD_TYPE);
const submitting = ref(false);
const loadTypeOptions: { label: string; value: ExerciseLoadType }[] = [
    { label: "Plate loaded — count 45 lb bar", value: "plate_loaded_with_bar" },
    { label: "Plate loaded — do not count bar", value: "plate_loaded_without_bar" },
    { label: "Plate-loaded machine — total plates", value: "plate_loaded_total" },
    { label: "Weight stack machine", value: "weight_stack" },
    { label: "Free weights", value: "free_weights" },
];

watch(
    () => props.exercise,
    (e) => {
        name.value = e.name;
        repRollover.value = e.rep_rollover || 10;
        cues.value = e.cues ?? "";
        loadType.value = e.load_type ?? DEFAULT_EXERCISE_LOAD_TYPE;
    },
    { immediate: true },
);

const submit = async () => {
    const n = name.value.trim();
    if (!n) {
        toast.push("Name is required", "error");
        return;
    }
    submitting.value = true;
    try {
        await apiClient.put<{ exercise: Exercise }>(
            `/workout/exercises/${props.exercise.ID}`,
            {
                name: n,
                rep_rollover: repRollover.value || 10,
                cues: cues.value.trim(),
                load_type: loadType.value,
            },
        );
        toast.push(`Updated ${n}`, "success");
        await queryClient.invalidateQueries({ queryKey: searchListKeys.all });
        await queryClient.invalidateQueries({ queryKey: workoutKeys.plans.all });
        props.onResolve?.(true);
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push(message, "error");
    } finally {
        submitting.value = false;
    }
};
</script>

<template>
    <form class="flex flex-col gap-4" @submit.prevent="submit">
        <Field label="Name">
            <Input v-model="name" autocomplete="off" />
        </Field>
        <Field label="Rep rollover">
            <NumberField v-model="repRollover" label="rep rollover" :min="1" :step="1" :decimals="0" />
        </Field>
        <Field label="Weight type">
            <Select v-model="loadType" :options="loadTypeOptions" />
        </Field>
        <Field label="Cues">
            <Textarea v-model="cues" :rows="4" placeholder="Optional cues (e.g. bullet points)" />
        </Field>
        <div class="flex flex-col gap-2 lg:flex-row lg:justify-end">
            <Button type="submit" variant="primary" :loading="submitting">
                {{ submitting ? "Saving…" : "Save" }}
            </Button>
        </div>
    </form>
</template>
