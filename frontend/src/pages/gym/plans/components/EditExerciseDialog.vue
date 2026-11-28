<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import type { Exercise, ExerciseLoadType, MachineSetting } from "~/types/workout";
import { DEFAULT_EXERCISE_LOAD_TYPE } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { useSaveMachineSettings, useUpdateExercise } from "~/api/workout/routines";
import { Button, Field, IconButton, Input, NumberField, Select, Textarea, Toggle } from "~/shared/ui";
import { Plus, X } from "lucide-vue-next";
import { computed, ref, watch } from "vue";

const props = defineProps<{
    exercise: Exercise;
    onResolve?: (ok: boolean) => void;
}>();

const updateExercise = useUpdateExercise();
const saveSettings = useSaveMachineSettings();
// The exercise is shared: only its creator (or an admin) may change it. Machine settings are the
// caller's own, so anyone can edit those.
const canEdit = computed(() => props.exercise.can_edit !== false);
const settings = ref<MachineSetting[]>([]);
const name = ref("");
const repRollover = ref<number | null>(10);
const cues = ref("");
const unilateral = ref(false);
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
        unilateral.value = e.unilateral ?? false;
        settings.value = (e.machine_settings ?? []).map((m) => ({ ...m }));
    },
    { immediate: true },
);

const submit = async () => {
    const n = name.value.trim();
    if (!n) {
        toast.push("Name is required", "error");
        return;
    }
    const rows = settings.value.filter((m) => m.label.trim() || m.value.trim());
    if (rows.some((m) => !m.label.trim())) {
        toast.push("Every machine setting needs a label", "error");
        return;
    }
    submitting.value = true;
    try {
        if (canEdit.value) {
            await updateExercise.mutateAsync({
                id: props.exercise.ID,
                name: n,
                rep_rollover: repRollover.value || 10,
                cues: cues.value.trim(),
                load_type: loadType.value,
                unilateral: unilateral.value,
            });
        }
        await saveSettings.mutateAsync({ id: props.exercise.ID, machine_settings: rows });
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
        <p v-if="!canEdit" class="m-0 text-sm text-fg-muted">
            This exercise is shared with everyone, and only the person who added it (or an admin) can change it. Your machine settings below are yours alone.
        </p>
        <Field label="Name">
            <Input v-model="name" autocomplete="off" :disabled="!canEdit" />
        </Field>
        <Field label="Rep rollover">
            <NumberField v-model="repRollover" label="rep rollover" :min="1" :step="1" :decimals="0" :disabled="!canEdit" />
        </Field>
        <Field label="Weight type">
            <Select v-model="loadType" :options="loadTypeOptions" :disabled="!canEdit" />
        </Field>
        <div class="flex flex-col gap-1">
            <Toggle v-model="unilateral" label="Unilateral" :disabled="!canEdit" />
            <p class="m-0 text-sm text-fg-muted">Log left and right separately, with a short switch timer between sides.</p>
        </div>
        <Field label="Cues">
            <Textarea v-model="cues" :rows="4" placeholder="Optional cues (e.g. bullet points)" :disabled="!canEdit" />
        </Field>
        <div class="flex flex-col gap-2">
            <span class="text-sm font-medium text-fg">Your machine settings</span>
            <p v-if="!settings.length" class="m-0 text-sm text-fg-muted">Seat, pad, pulley… shown while you log this exercise.</p>
            <div v-for="(m, i) in settings" :key="i" class="flex items-center gap-2">
                <div class="min-w-0 flex-1"><Input v-model="m.label" placeholder="Label (e.g. Seat)" aria-label="Setting label" autocomplete="off" /></div>
                <div class="min-w-0 flex-1"><Input v-model="m.value" placeholder="Value (e.g. 4)" aria-label="Setting value" autocomplete="off" /></div>
                <IconButton label="Remove setting" @click="settings.splice(i, 1)"><X /></IconButton>
            </div>
            <div>
                <Button size="sm" variant="secondary" @click="settings.push({ label: '', value: '' })">
                    <Plus class="size-4" aria-hidden="true" /> Add setting
                </Button>
            </div>
        </div>
        <div class="flex flex-col gap-2 lg:flex-row lg:justify-end">
            <Button type="submit" variant="primary" :loading="submitting">
                {{ submitting ? "Saving…" : "Save" }}
            </Button>
        </div>
    </form>
</template>
