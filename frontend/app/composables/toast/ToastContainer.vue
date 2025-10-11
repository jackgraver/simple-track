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
    z-index: 999;
}

.toast-container > div {
    padding-left: 1rem;
    padding-right: 1rem;
    border-radius: 0.25rem;
    color: white;
}

.toast {
    padding: 1.5rem 2rem;
    margin-top: 0.5rem;
    font-size: 1.5rem;
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

@keyframes slideInRight {
    from {
        transform: translateX(100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}

.slideInRight {
    animation-name: slideInRight;
    animation-duration: 0.25s;
    animation-timing-function: ease-out;
    animation-iteration-count: 1;
    animation-fill-mode: forwards;
}
</style>
