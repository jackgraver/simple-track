<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { parseDietDayOffsetQuery } from "~/utils/dateUtil";
import { gymDayOptions, useAdjacentDayPrefetch } from "~/api/prefetchDays";
import GymHome from "./home/GymHome.vue";

const route = useRoute();
const isGymHome = computed(() => route.name === "gym");
const dayOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));

useAdjacentDayPrefetch(dayOffset, gymDayOptions, isGymHome);
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <GymHome v-if="isGymHome" :offset="dayOffset" />
        <router-view v-else />
    </div>
</template>
