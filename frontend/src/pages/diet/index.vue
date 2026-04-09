<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import DietDayView from "./components/DietDayView.vue";
import { CREATE_TYPE, LOG_TYPE } from "~/pages/diet/logmeal/logmealMode";

const route = useRoute();
const isDietHome = computed(() => route.name === "diet");
</script>

<template>
    <div
        v-if="isDietHome"
        class="mx-auto flex max-w-3xl flex-col gap-4 pb-8 pt-2"
    >
        <div class="flex flex-wrap items-center gap-3">
            <h1 class="m-0 text-xl font-semibold tracking-tight text-zinc-100">
                Diet
            </h1>
            <router-link
                :to="{ name: 'diet-log', query: { type: LOG_TYPE } }"
                class="rounded-md border border-amber-700/50 bg-amber-950/40 px-3 py-2 text-sm font-medium text-amber-100 hover:bg-amber-950/70"
            >
                Log meal
            </router-link>
            <router-link
                :to="{ name: 'diet-log', query: { type: CREATE_TYPE } }"
                class="rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-sm font-medium text-zinc-200 hover:bg-zinc-700"
            >
                New saved meal
            </router-link>
        </div>
        <DietDayView :date-offset="0" />
    </div>
    <router-view v-else />
</template>
