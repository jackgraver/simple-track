<script setup lang="ts">
import type { Cardio, PlannedCardio } from "~/types/workout";
import { ArrowLeft } from "lucide-vue-next";
import NumericStepper from "./NumericStepper.vue";
import { ref, watch, computed, nextTick, useId } from "vue";

const props = defineProps<{
    plannedCardio: PlannedCardio | null;
    loggedCardio: Cardio | null;
}>();

const emit = defineEmits<{
    (e: "save", minutes: number): void;
    (e: "back"): void;
}>();

const cardioName = computed(
    () => props.plannedCardio?.type ?? props.loggedCardio?.type ?? "Cardio",
);

const isLogged = computed(() => (props.loggedCardio?.minutes ?? 0) > 0);

const currentMinutes = ref(0);

const editMode = ref(false);
const inputValue = ref("");
const error = ref("");
const inputId = useId();

watch(
    () => props.loggedCardio,
    (c) => {
        currentMinutes.value = c?.minutes ?? 0;
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

const finish = () => {
    if (currentMinutes.value <= 0) {
        error.value = "Enter minutes before saving.";
        return;
    }
    emit("save", currentMinutes.value);
};
</script>

<template>
    <div class="logging-view">
        <div class="logging-header">
            <button class="back-button" type="button" @click="emit('back')">
                <ArrowLeft :size="20" />
            </button>
            <h2>{{ cardioName }}</h2>
            <span v-if="isLogged" class="logged-badge">Logged</span>
        </div>
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
.logging-header {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 0.5rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgb(56, 56, 56);
    width: 100%;
    box-sizing: border-box;
}
.back-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2.5rem;
    height: 2.5rem;
    background: transparent;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 0.25rem;
    color: inherit;
    cursor: pointer;
    transition: background-color 0.2s, border-color 0.2s;
    padding: 0;
}
.back-button:hover {
    background: rgb(40, 40, 40);
    border-color: rgb(100, 100, 100);
}
.logging-header h2 {
    margin: 0;
    text-align: center;
    grid-column: 2;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
}
.logged-badge {
    color: rgb(100, 200, 120);
    font-size: 0.85rem;
    font-weight: 500;
    grid-column: 3;
    white-space: nowrap;
}
.input-group {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}
@media (max-width: 767px) {
    .logging-view {
        padding: 0 1rem 0 0.5rem;
    }
    .logging-header {
        gap: 0.75rem;
    }
    .logging-header h2 {
        font-size: 1.25rem;
    }
    .back-button {
        width: 2.25rem;
        height: 2.25rem;
    }
}
</style>
