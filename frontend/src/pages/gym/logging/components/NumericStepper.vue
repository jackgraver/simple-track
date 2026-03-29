<script setup lang="ts">
import { nextTick, useId } from "vue";

defineProps<{
    label: string;
    modelValue: number;
    editMode: boolean;
    inputValue: string;
    error?: string;
    inputStep?: string;
}>();

const emit = defineEmits<{
    increment: [];
    decrement: [];
    "update:inputValue": [value: string];
    "enter-edit": [];
    "exit-edit": [];
}>();

const inputId = useId();

const enterEdit = () => {
    emit("enter-edit");
    nextTick(() => {
        const input = document.getElementById(inputId) as HTMLInputElement | null;
        if (input) {
            input.focus();
            input.select();
        }
    });
};

const onInput = (e: Event) => {
    emit("update:inputValue", (e.target as HTMLInputElement).value);
};
</script>

<template>
    <div class="stepper-container">
        <label :for="inputId">{{ label }}</label>
        <div :class="['stepper', { 'stepper-error': error }]">
            <button
                class="stepper-button"
                type="button"
                @click="emit('decrement')"
            >
                −
            </button>
            <div
                v-if="!editMode"
                class="stepper-display"
                @click="enterEdit"
            >
                {{ modelValue || 0 }}
            </div>
            <input
                v-else
                :id="inputId"
                type="number"
                class="stepper-input"
                :value="inputValue"
                min="0"
                :step="inputStep ?? '1'"
                @input="onInput"
                @blur="emit('exit-edit')"
                @keyup.enter="emit('exit-edit')"
                @keyup.escape="emit('exit-edit')"
            />
            <button
                class="stepper-button"
                type="button"
                @click="emit('increment')"
            >
                +
            </button>
        </div>
        <p v-if="error" class="stepper-error-text">{{ error }}</p>
    </div>
</template>

<style scoped>
.stepper-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.stepper-container label {
    font-weight: 500;
    font-size: 0.9rem;
    color: rgb(150, 150, 150);
}

.stepper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
}

.stepper-error .stepper-display,
.stepper-error .stepper-input,
.stepper-error .stepper-button {
    border-color: rgb(150, 60, 60);
}

.stepper-button {
    width: 3rem;
    height: 3rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1.5rem;
    font-weight: 300;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition:
        background-color 0.2s,
        border-color 0.2s;
    user-select: none;
}

.stepper-button:hover {
    background: rgb(40, 40, 40);
    border-color: rgb(100, 100, 100);
}

.stepper-button:active {
    background: rgb(50, 50, 50);
}

.stepper-display {
    min-width: 6rem;
    height: 3rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
    font-weight: 500;
    cursor: pointer;
    user-select: none;
    padding: 0 1rem;
    border-radius: 5px;
    transition: background-color 0.2s;
}

.stepper-display:hover {
    background: rgb(30, 30, 30);
}

.stepper-input {
    min-width: 6rem;
    height: 3rem;
    padding: 0 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 2rem;
    font-weight: 500;
    text-align: center;
}

.stepper-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}

.stepper-error-text {
    margin: 0;
    color: rgb(220, 110, 110);
    font-size: 0.85rem;
}
</style>
