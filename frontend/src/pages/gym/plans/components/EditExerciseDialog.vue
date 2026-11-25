<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import type { Exercise, ExerciseLoadType } from "~/types/workout";
import { DEFAULT_EXERCISE_LOAD_TYPE } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { useUpdateExercise } from "~/api/workout/routines";
import { Button, Field, Input, NumberField, Select, Textarea } from "~/shared/ui";
import { computed, ref, watch } from "vue";
import {
    DEFAULT_WEIGHT_STEP_LBS,
    defaultBaseWeight,
    isPlateLoaded,
    normalizeLoadType,
} from "~/pages/gym/logging/exercise/domain/plates";

const props = defineProps<{
    exercise: Exercise;
    onResolve?: (ok: boolean) => void;
}>();

const updateExercise = useUpdateExercise();
const name = ref("");
const repRollover = ref<number | null>(10);
const cues = ref("");
const loadType = ref<ExerciseLoadType>(DEFAULT_EXERCISE_LOAD_TYPE);
/** Empty (null) means the load type's default, shown as the placeholder. */
const baseWeight = ref<number | null>(null);
const weightStep = ref<number | null>(null);
const submitting = ref(false);
const loadTypeOptions: { label: string; value: ExerciseLoadType }[] = [
    { label: "Plates on each side (barbell, sled)", value: "plate_loaded_per_side" },
    { label: "Plates in total (single-peg machine)", value: "plate_loaded_total" },
    { label: "Weight stack machine", value: "weight_stack" },
    { label: "Free weights", value: "free_weights" },
];
const plateLoaded = computed(() => isPlateLoaded(loadType.value));
const baseDefault = computed(() => defaultBaseWeight(loadType.value));

watch(
    () => props.exercise,
    (e) => {
        name.value = e.name;
        repRollover.value = e.rep_rollover || 10;
        cues.value = e.cues ?? "";
        loadType.value = normalizeLoadType(e.load_type ?? DEFAULT_EXERCISE_LOAD_TYPE);
        // A legacy "without bar" exercise (stale cache) means base 0.
        baseWeight.value = e.base_weight ?? (e.load_type === "plate_loaded_without_bar" ? 0 : null);
        weightStep.value = e.weight_step ?? null;
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
        await updateExercise.mutateAsync({
            id: props.exercise.ID,
            name: n,
            rep_rollover: repRollover.value || 10,
            cues: cues.value.trim(),
            load_type: loadType.value,
            base_weight: plateLoaded.value && baseWeight.value != null && baseWeight.value >= 0 ? baseWeight.value : null,
            weight_step: weightStep.value != null && weightStep.value > 0 ? weightStep.value : null,
        });
        toast.push(`Updated ${n}`, "success");
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
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Field
                v-if="plateLoaded"
                label="Bar / base weight (lb)"
                hint="Loaded before any plates. 0 when the bar or sled isn't counted."
            >
                <NumberField
                    v-model="baseWeight"
                    label="base weight"
                    :min="0"
                    :step="5"
                    :decimals="1"
                    unit="lb"
                    :placeholder="`${baseDefault} (default)`"
                />
            </Field>
            <Field
                label="Weight increase (lb)"
                :hint="`Added when every set hits the rollover, and the +/− step. Empty uses ${DEFAULT_WEIGHT_STEP_LBS}; barbells often use 10 (5 a side).`"
            >
                <NumberField
                    v-model="weightStep"
                    label="weight increase"
                    :min="0"
                    :step="5"
                    :decimals="1"
                    unit="lb"
                    :placeholder="`${DEFAULT_WEIGHT_STEP_LBS} (default)`"
                />
            </Field>
        </div>
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
