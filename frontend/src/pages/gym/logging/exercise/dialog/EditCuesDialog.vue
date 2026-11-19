<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import { ref } from "vue";
import { Exercise } from "~/types/workout";
import { apiPUT } from "~/api/client";
import { Button, Field, Textarea } from "~/shared/ui";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    exercise: Exercise;
    currentCues: string;
    onResolve?: (value: string) => void;
}>();

const cues = ref(props.currentCues);
const submitting = ref(false);

const saveCues = async () => {
    submitting.value = true;
    try {
        const nextCues = cues.value.trim();
        await apiPUT<{ exercise: Exercise }>(
            `/workout/exercises/${props.exercise.ID}/cues`,
            {
                cues: nextCues,
            },
        );
        toast.push("Cues saved", "success");

        props.onResolve?.(nextCues);
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push(message, "error");
    } finally {
        submitting.value = false;
    }
};
</script>

<template>
    <div class="flex flex-col gap-3">
        <Field label="Cues">
            <Textarea v-model="cues" :rows="4" placeholder="Optional cues (e.g. bullet points)" />
        </Field>
        <div class="flex flex-col gap-2 lg:flex-row lg:justify-end">
            <Button variant="primary" :loading="submitting" @click="saveCues">
                {{ submitting ? "Saving…" : "Save" }}
            </Button>
        </div>
    </div>
</template>
