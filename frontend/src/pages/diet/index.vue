<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import DietDayView from "./components/DietDayView.vue";
import { CREATE_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { parseDietDayOffsetQuery, withDietDayOffsetQuery } from "~/utils/dateUtil";
import { dietDayOptions, useAdjacentDayPrefetch } from "~/api/prefetchDays";
import ModuleTitleBar from "~/shared/ModuleTitleBar.vue";
import type { ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";

const route = useRoute();
const isDietHome = computed(() => route.name === "diet");

const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
useAdjacentDayPrefetch(dateOffset, dietDayOptions, isDietHome);

// Logging (meal, quick log, water) lives on the page itself; the title bar is for planning.
const dietNavLinks = computed<ModuleTitleLink[]>(() => [
    { name: "Plan meals", to: "diet-edit-planned", query: { offset: String(dateOffset.value) } },
    { name: "Targets", to: "diet-targets" },
    {
        name: "New saved meal",
        to: "diet-log",
        query: withDietDayOffsetQuery(dateOffset.value, { type: CREATE_TYPE }),
    },
]);
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <template v-if="isDietHome">
            <ModuleTitleBar title="Diet" :links="dietNavLinks" />
            <DietDayView :date-offset="dateOffset" />
        </template>
        <router-view v-else />
    </div>
</template>
