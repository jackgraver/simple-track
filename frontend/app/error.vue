<script setup lang="ts">
import type { NuxtError } from "#app";
import { Home } from "lucide-vue-next";

const props = defineProps({
    error: Object as () => NuxtError,
});
</script>

<template>
    <div class="error-page">
        <div class="error-content">
            <div class="error-message">
                <h1>{{ error?.statusCode || "Error" }}</h1>
                <p>{{ error?.message || "Something went wrong." }}</p>
            </div>
            <NuxtLink to="/" class="home-button">
                <Home :size="28" />
                <span>Go back home</span>
            </NuxtLink>
        </div>
    </div>
</template>

<style scoped>
.error-page {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 80vh;
    width: 100%;
    color: #fff;
    text-align: center;
}

.error-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 2rem;
}

.error-message {
    display: flex;
    flex-direction: column;
}

.error-message h1 {
    font-size: 4rem;
    margin: 0;
    color: #ce3b3b;
}

.error-message p {
    font-size: 1.25rem;
    color: #ccc;
    max-width: 400px;
    margin-top: 0;
}

.home-button {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    margin-top: 1rem;
    padding: 0.75rem 1.5rem;
    font-size: 1.1rem;
    font-weight: bold;
    background-color: #2f6;
    color: #2e2e2e;
    border-radius: 8px;
    text-decoration: none;
    transition: background-color 0.2s ease;
}

.home-button:hover {
    background-color: #2c2;
}
</style>
