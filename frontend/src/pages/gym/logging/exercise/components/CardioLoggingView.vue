<script setup lang="ts">
import type { Cardio, PlannedCardio } from "~/types/workout";
import LoggingHeader from "./LoggingHeader.vue";
import NumericStepper from "./NumericStepper.vue";
import { ref, watch, computed, nextTick, useId } from "vue";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import { toast } from "~/composables/toast/useToast";

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

const currentMinutes = ref(0);
const notes = ref("");

const editMode = ref(false);
const inputValue = ref("");
const error = ref("");
const inputId = useId();

watch(
    () => props.loggedCardio,
    (c) => {
        currentMinutes.value = c?.minutes ?? 0;
        notes.value = c?.notes ?? "";
    },
    { immediate: true },
);

const increment = () => {
    error.value = "";
    currentMinutes.value = (currentMinutes.value || 0) + 1;
};

const decrement = () => {
    error.value = "";
    currentMinutes.value = Math.max(0, (currentMinutes.value || 0) - 1);
};

const enterEdit = () => {
    editMode.value = true;
    inputValue.value = currentMinutes.value.toString();
    nextTick(() => {
        const el = document.getElementById(inputId) as HTMLInputElement | null;
        if (el) {
            el.focus();
            el.select();
        }
    });
};

const exitEdit = () => {
    const trimmed = inputValue.value.trim();
    if (trimmed !== "") {
        const n = Number(trimmed);
        if (Number.isFinite(n) && n >= 0) {
            currentMinutes.value = Math.round(n);
            error.value = "";
        } else {
            error.value = "Enter a valid number of minutes.";
        }
    }
    editMode.value = false;
};

const updateInputValue = (v: string) => {
    inputValue.value = v;
    error.value = "";
};

const finish = async () => {
    if (currentMinutes.value <= 0) {
        error.value = "Enter minutes before saving.";
        return;
    }
    try {
        await saveCardio(currentMinutes.value, undefined, notes.value);
        toast.push("Cardio saved", "success");
        goBackToLogging();
    } catch (err: any) {
        toast.push(err.message || "Failed to save cardio", "error");
    }
};
</script>

<template>
    <div class="logging-view">
        <LoggingHeader :title="cardioName" @back="goBackToLogging">
            <template #right>
                <span v-if="isLogged" class="logged-badge">Logged</span>
            </template>
        </LoggingHeader>
        <div class="input-group">
            <NumericStepper
                label="Time (minutes)"
                :model-value="currentMinutes"
                :edit-mode="editMode"
                :input-value="inputValue"
                :error="error"
                @increment="increment"
                @decrement="decrement"
                @update:input-value="updateInputValue"
                @enter-edit="enterEdit"
                @exit-edit="exitEdit"
            />
        </div>
        <div class="input-container">
            <label>Notes</label>
            <textarea
                class="notes-input"
                :value="notes"
                placeholder="What did you watch? Any thoughts..."
                rows="3"
                @input="notes = ($event.target as HTMLTextAreaElement).value"
            ></textarea>
        </div>
        <button
            class="bg-green-600 hover:bg-green-500 w-full rounded py-3 text-base font-medium text-white"
            type="button"
            @click="finish"
        >
            Finish Cardio
        </button>
    </div>
</template>

<style scoped>
.logging-view {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    width: 100%;
    max-width: 100%;
    padding: 0 0.75rem 0 0.375rem;
    box-sizing: border-box;
}
.logged-badge {
    color: rgb(100, 200, 120);
    font-size: 0.85rem;
    font-weight: 500;
    white-space: nowrap;
}
.input-group {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}
.input-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}
.input-container label {
    font-weight: 500;
    font-size: 0.9rem;
    color: rgb(150, 150, 150);
}
.notes-input {
    padding: 0.75rem 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
    transition:
        border-color 0.2s,
        background-color 0.2s;
}
.notes-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}
.notes-input::placeholder {
    color: rgb(100, 100, 100);
}
@media (max-width: 767px) {
    .logging-view {
        padding: 0 1rem 0 0.5rem;
    }
}
</style>
