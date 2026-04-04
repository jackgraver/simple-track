<script setup lang="ts">
import { useId } from "vue";

const modelValue = defineModel<string>({ required: true });

const props = withDefaults(
    defineProps<{
        label: string;
        type: string;
        error?: string;
        required?: boolean;
        autocomplete?: string;
        placeholder?: string;
    }>(),
    {
        error: "",
        required: false,
        autocomplete: "",
        placeholder: "",
    },
);

const inputId = useId();
</script>

<template>
    <label :for="inputId">{{ props.label }}</label>
    <input
        :id="inputId"
        v-model="modelValue"
        :type="props.type"
        :required="props.required"
        :autocomplete="props.autocomplete"
        :placeholder="props.placeholder"
        class="rounded-md"
    />
    <p v-if="props.error" class="error-message">{{ props.error }}</p>
</template>

<style scoped>
.form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

label {
    color: rgb(200, 200, 200);
    font-size: 0.9rem;
    font-weight: 500;
}

input {
    padding: 0.5rem;
    background: rgb(35, 35, 35);
    border: 1px solid rgb(56, 56, 56);
    color: white;
    font-size: 1rem;
    transition:
        border-color 0.2s,
        background-color 0.2s;
}

input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(40, 40, 40);
}

input::placeholder {
    color: rgb(120, 120, 120);
}

p {
    padding: 0.75rem;
    background: rgb(60, 20, 20);
    border: 1px solid rgb(120, 40, 40);
    border-radius: 4px;
    color: rgb(255, 150, 150);
    font-size: 0.9rem;
    text-align: center;
}
</style>
