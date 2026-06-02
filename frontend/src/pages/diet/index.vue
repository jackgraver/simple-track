<script setup lang="ts">
import "./macro-colors.css";
import { computed } from "vue";
import { useRoute } from "vue-router";
import DietDayView from "./components/DietDayView.vue";
import { CREATE_TYPE, LOG_TYPE } from "~/pages/diet/logmeal/logmealMode";
import {
    parseDietDayOffsetQuery,
    withDietDayOffsetQuery,
    ymdForDayOffset,
} from "~/utils/dateUtil";
import { dialogManager } from "~/composables/dialog/useDialog";
import LogWaterDialog from "./components/dialog/LogWaterDialog.vue";
import QuickLogDialog from "./components/dialog/QuickLogDialog.vue";
import ModuleTitleBar from "~/shared/ModuleTitleBar.vue";
import type { ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";

const route = useRoute();
const isDietHome = computed(() => route.name === "diet");

const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const dateStr = computed(() => ymdForDayOffset(dateOffset.value ?? 0));

const logMealLinkQuery = computed(() =>
    withDietDayOffsetQuery(dateOffset.value ?? 0, { type: LOG_TYPE }),
);
const createSavedMealLinkQuery = computed(() =>
    withDietDayOffsetQuery(dateOffset.value ?? 0, { type: CREATE_TYPE }),
);

async function openWaterLog() {
    await dialogManager.custom<boolean>({
        title: "Log water",
        component: LogWaterDialog,
        componentProps: { dateStr: dateStr.value },
    });
}
async function openQuickLog() {
    await dialogManager.custom<boolean>({
        title: "Quick log",
        component: QuickLogDialog,
        componentProps: {
            dateOffset: dateOffset.value ?? 0,
        },
    });
}

const dietNavLinks = computed<ModuleTitleLink[]>(() => [
    { name: "Log meal", to: "diet-log", query: logMealLinkQuery.value },
    {
        name: "New saved meal",
        to: "diet-log",
        query: createSavedMealLinkQuery.value,
    },
    { name: "Macro targets", to: "diet-targets" },
    { name: "Quick log", onClick: openQuickLog },
    { name: "Log Water", onClick: openWaterLog },
    { name: "Grocery List", to: "grocery" },
]);
</script>

<template>
    <div class="diet-module flex w-full flex-col gap-6">
        <template v-if="isDietHome">
            <ModuleTitleBar title="Diet" :links="dietNavLinks" />
            <DietDayView :date-offset="dateOffset" />
        </template>
        <router-view v-else />
    </div>
</template>
