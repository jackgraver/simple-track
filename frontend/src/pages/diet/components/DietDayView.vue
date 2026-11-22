<script setup lang="ts">
import { computed } from "vue";
import { Plus, Zap } from "lucide-vue-next";
import { useDietLogsToday } from "~/api/diet/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { LOG_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { withDietDayOffsetQuery } from "~/utils/dateUtil";
import { Button, ErrorState } from "~/shared/ui";
import DaySummaryCard from "../home/DaySummaryCard.vue";
import DietWeekStrip from "../home/DietWeekStrip.vue";
import QuickLogDialog from "./dialog/QuickLogDialog.vue";
import LoggedMealsDisplay from "./LoggedMealsDisplay.vue";
import PlannedMealsDisplay from "./PlannedMealsDisplay.vue";

const props = defineProps<{
    dateOffset: number;
}>();

const { error, refetch } = useDietLogsToday(() => props.dateOffset);

const logMealTo = computed(() => ({
    name: "diet-log",
    query: withDietDayOffsetQuery(props.dateOffset, { type: LOG_TYPE }),
}));

function openQuickLog() {
    void dialogManager.custom<boolean>({
        title: "Quick log",
        component: QuickLogDialog,
        componentProps: { dateOffset: props.dateOffset },
    });
}
</script>

<template>
    <div class="flex w-full flex-col gap-6 pb-above-tabbar lg:pb-10">
        <DietWeekStrip :offset="dateOffset" />

        <ErrorState v-if="error" title="Couldn't load this day" :error="error" @retry="refetch()" />
        <!-- Mobile: summary, lists, then the log buttons in thumb reach.
             Desktop: summary + log buttons on the left, lists on the right. -->
        <div v-else class="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)] lg:items-start">
            <div class="flex flex-col gap-4">
                <DaySummaryCard :date-offset="dateOffset" />
                <div class="hidden gap-2 lg:flex">
                    <Button variant="primary" class="flex-1" :to="logMealTo">
                        <template #icon-left><Plus class="size-4" /></template>
                        Log meal
                    </Button>
                    <Button variant="secondary" @click="openQuickLog">
                        <template #icon-left><Zap class="size-4" /></template>
                        Quick log
                    </Button>
                </div>
            </div>
            <div class="flex min-w-0 flex-col gap-6">
                <PlannedMealsDisplay :date-offset="dateOffset" />
                <LoggedMealsDisplay :date-offset="dateOffset" />
            </div>
            <div class="flex gap-2 lg:hidden">
                <Button variant="primary" size="lg" class="flex-1" :to="logMealTo">
                    <template #icon-left><Plus class="size-4" /></template>
                    Log meal
                </Button>
                <Button variant="secondary" size="lg" @click="openQuickLog">
                    <template #icon-left><Zap class="size-4" /></template>
                    Quick log
                </Button>
            </div>
        </div>
    </div>
</template>
