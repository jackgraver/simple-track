<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
    weightSetup: string;
    onResolve?: (value: string) => void;
    onCancel?: () => void;
}>();

const value = ref(props.weightSetup);

const confirm = () => props.onResolve?.(value.value);
const cancel = () => props.onCancel?.();
</script>

<template>
    <div class="container">
        <label>Weight Setup</label>
        <input type="text" v-model="value" class="input" />

        <div class="buttons">
            <button class="confirm-button" @click="confirm">Confirm</button>
            <button @click="cancel">Cancel</button>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.5rem;
}

.buttons {
    display: flex;
    justify-content: center;
    padding-top: 0.5rem;
    gap: 0.75rem;
}
</style>
