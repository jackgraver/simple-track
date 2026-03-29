<script setup lang="ts">
import type { Exercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "~/utils/axios";
import { ref } from "vue";

const props = defineProps<{
    onResolve?: (ok: boolean) => void;
}>();

const queryClient = useQueryClient();
const name = ref("");
const repRollover = ref(10);
const cues = ref("");
const submitting = ref(false);

const submit = async () => {
    const n = name.value.trim();
    if (!n) {
        toast.push("Name is required", "error");
        return;
    }
    submitting.value = true;
    try {
        const createRes = await apiClient.post<{ exercise: Exercise }>("/workout/exercises", {
            name: n,
            rep_rollover: repRollover.value || 10,
            cues: cues.value.trim(),
        });
        const exercise = createRes.data.exercise;
        toast.push(`Created ${exercise.name}`, "success");
        await queryClient.invalidateQueries({ queryKey: ["searchList"] });
        props.onResolve?.(true);
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push(message, "error");
    } finally {
        submitting.value = false;
    }
};
</script>

<template>
    <div class="create-ex-form">
        <label class="field">
            <span>Name</span>
            <input v-model="name" type="text" autocomplete="off" class="input" />
        </label>
        <label class="field">
            <span>Rep rollover</span>
            <input v-model.number="repRollover" type="number" min="1" class="input" />
        </label>
        <label class="field">
            <span>Cues</span>
            <textarea v-model="cues" class="input textarea" rows="4" placeholder="Optional cues (e.g. bullet points)"></textarea>
        </label>
        <div class="actions">
            <button type="button" class="btn primary" :disabled="submitting" @click="submit">
                {{ submitting ? "Saving…" : "Create exercise" }}
            </button>
        </div>
    </div>
</template>

<style scoped>
.create-ex-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-width: 260px;
}
.field {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    font-size: 0.9rem;
    color: #ccc;
}
.input {
    padding: 0.5rem 0.65rem;
    background: #2a2a2a;
    border: 1px solid #444;
    border-radius: 0.25rem;
    color: #fff;
    font-size: 1rem;
}
.input:focus {
    outline: none;
    border-color: #4a9eff;
}
.textarea {
    resize: vertical;
    font-family: inherit;
}
.actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 0.25rem;
}
.btn.primary {
    padding: 0.5rem 1rem;
    background: #4a9eff;
    color: #fff;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-size: 0.9rem;
}
.btn.primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
</style>
