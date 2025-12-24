<script setup lang="ts">
import { toast } from "~/composables/toast/useToast";
</script>

<template>
    <div class="toast-container">
        <div
            v-for="t in toast.toasts.value"
            :key="t.id"
            :class="{
                slideInRight: true,
                toast: true,
                success: t.type === 'success',
                error: t.type === 'error',
                info: t.type === 'info',
            }"
        >
            {{ t.message }}
        </div>
    </div>
</template>

<style scoped>
.toast-container {
    position: fixed;
    bottom: 1rem;
    right: 1rem;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.5rem;
    pointer-events: none; /* allow clicks through */
}

.toast {
    background-color: gray;
    color: white;
    border-radius: 6px;
    padding: 0.4rem 0.8rem;
    font-size: 0.85rem;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
    pointer-events: auto;
}

.success {
    background-color: #28a745;
}
.error {
    background-color: #dc3545;
}
.info {
    background-color: #007bff;
}
</style>
