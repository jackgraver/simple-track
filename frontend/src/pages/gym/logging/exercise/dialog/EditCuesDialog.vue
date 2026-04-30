<script setup lang="ts">
import { ref } from "vue";
import { Exercise } from "~/types/workout";
import { apiPUT } from "~/api/client";
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
        const message = err instanceof Error ? err.message : String(err);
        toast.push(message, "error");
    } finally {
        submitting.value = false;
    }
};
</script>

<template>
    <div class="flex flex-col gap-2">
        <textarea
            v-model="cues"
            class="input textarea border border-secondBg rounded-md p-2"
            rows="4"
            placeholder="Optional cues (e.g. bullet points)"
        ></textarea>
        <button
            type="button"
            class="bg-green-500 text-white px-4 py-2 rounded-md disabled:opacity-60"
            :disabled="submitting"
            @click="saveCues"
        >
            {{ submitting ? "Saving…" : "Save" }}
        </button>
    </div>
</template>
