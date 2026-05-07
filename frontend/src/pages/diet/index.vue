<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import DietDayView from "./components/DietDayView.vue";
import { CREATE_TYPE, LOG_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { parseDietDayOffsetQuery, ymdForDayOffset } from "~/utils/dateUtil";
import { dialogManager } from "~/composables/dialog/useDialog";
import LogWaterDialog from "./components/dialog/LogWaterDialog.vue";

const route = useRoute();
const isDietHome = computed(() => route.name === "diet");

const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const dateStr = computed(() => ymdForDayOffset(dateOffset.value ?? 0));

async function openWaterLog() {
    await dialogManager.custom<boolean>({
        title: "Log water",
        component: LogWaterDialog,
        componentProps: { dateStr: dateStr.value },
    });
}
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <template v-if="isDietHome">
            <div
                class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
            >
                <h1 class="m-0 text-lg font-semibold text-textPrimary">Diet</h1>
                <nav class="flex items-center gap-3 text-sm text-textSecondary">
                    <router-link
                        :to="{ name: 'diet-log', query: { type: LOG_TYPE } }"
                        class="transition-colors hover:text-textPrimary"
                        >Log meal</router-link
                    >
                    <span aria-hidden="true" class="text-textSecondary/50"
                        >·</span
                    >
                    <router-link
                        :to="{ name: 'diet-log', query: { type: CREATE_TYPE } }"
                        class="transition-colors hover:text-textPrimary"
                        >New saved meal</router-link
                    >
                    <span aria-hidden="true" class="text-textSecondary/50"
                        >·</span
                    >
                    <router-link
                        :to="{ name: 'diet-targets' }"
                        class="transition-colors hover:text-textPrimary"
                        >Macro targets</router-link
                    >
                    <span aria-hidden="true" class="text-textSecondary/50"
                        >·</span
                    >
                    <button
                        class="transition-colors hover:text-textPrimary p-0!"
                        @click="openWaterLog"
                    >
                        Log Water
                    </button>
                </nav>
            </div>
            <DietDayView :date-offset="dateOffset" />
        </template>
        <router-view v-else />
    </div>
</template>
