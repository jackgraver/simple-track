<script setup lang="ts">
import { RouterLink } from "vue-router";
</script>

<template>
    <div class="flex flex-col items-center gap-3 py-16 text-center">
        <h1 class="m-0 text-2xl font-semibold">Page not found</h1>
        <p class="m-0 text-zinc-400">That page doesn't exist.</p>
        <RouterLink
            :to="{ name: 'gym' }"
            class="rounded-md bg-zinc-800 px-4 py-2 text-zinc-100 hover:bg-zinc-700"
        >
            Back to gym
        </RouterLink>
    </div>
</template>
