<script setup lang="ts">
import { toast } from "~/composables/toast/useToast";
</script>

<template>
    <div class="toast-container">
        <div
            v-for="t in toast.toasts.value"
            :key="t.id"
            :class="{
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
    z-index: 999;
}

.toast-container > div {
    padding-left: 1rem;
    padding-right: 1rem;
    border-radius: 0.25rem;
    color: white;
}

.success {
    background-color: green;
}
.error {
    background-color: red;
}
.info {
    background-color: blue;
}
</style>
