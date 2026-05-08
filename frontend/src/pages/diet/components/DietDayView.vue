<script setup lang="ts">
import MacroBar from "~/pages/diet/components/macrobar/MacroBar.vue";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import WaterLogsDialog from "~/pages/diet/components/dialog/WaterLogsDialog.vue";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import LoggedMealsDisplay from "./LoggedMealsDisplay.vue";
import PlannedMealsDisplay from "./PlannedMealsDisplay.vue";
import { useWaterLogs } from "~/api/tracking/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { formatDateLong, ymdForDayOffset } from "~/utils/dateUtil";
import { computed, ref } from "vue";

const props = defineProps<{
    dateOffset: number;
}>();

function getTotalNumber(
    pending: boolean,
    err: unknown,
    num: number | undefined,
): number {
    return pending || err ? 0 : (num ?? 0);
}

function planNum(
    pending: boolean,
    err: unknown,
    plan: unknown,
    key: "calories" | "protein" | "fat" | "carbs" | "fiber",
): number {
    if (pending || err) return 0;
    const r = plan as Record<string, number> | undefined;
    const v = r?.[key];
    return typeof v === "number" ? v : 0;
}

const macroPane = ref(0);

const {
    data,
    isPending: pending,
    error,
} = useDietLogsToday(() => props.dateOffset);

const dateStr = computed(() => ymdForDayOffset(props.dateOffset));
const { goalOz, displayUnit, formatVolumeFromOz } = useWaterPrefs();
const {
    data: waterLogs,
    isPending: waterPending,
    isError: waterError,
    error: waterQueryError,
} = useWaterLogs(dateStr);

const totalOz = computed(() =>
    (waterLogs.value ?? []).reduce((s, row) => s + row.amount_oz, 0),
);

const waterTotalDisplay = computed(
    () => formatVolumeFromOz(totalOz.value).value,
);
const waterPlannedDisplay = computed(
    () => formatVolumeFromOz(goalOz.value).value,
);
const waterUnitSuffix = computed(() => ` ${displayUnit.value}`);
const waterDayLabelLong = computed(() =>
    formatDateLong(`${dateStr.value}T12:00:00`),
);
function openWaterLogsDialog() {
    void dialogManager.custom<void>({
        title: "Water log",
        component: WaterLogsDialog,
        componentProps: {
            dateStr: dateStr.value,
            dayLabel: waterDayLabelLong.value,
        },
    });
}
</script>

<template>
    <div class="flex w-full flex-col gap-4">
        <div class="w-full">
            <div
                class="w-full overflow-hidden rounded-md pb-2"
                role="tablist"
                aria-label="Macros and hydration summary"
            >
                <div v-show="macroPane === 0" class="box-border px-0.5 pt-1">
                    <MacroBars
                        :totalCalories="
                            getTotalNumber(pending, error, data?.totalCalories)
                        "
                        :totalProtein="
                            getTotalNumber(pending, error, data?.totalProtein)
                        "
                        :totalFat="
                            getTotalNumber(pending, error, data?.totalFat)
                        "
                        :totalCarbs="
                            getTotalNumber(pending, error, data?.totalCarbs)
                        "
                        :plannedCalories="
                            planNum(pending, error, data?.day.plan, 'calories')
                        "
                        :plannedProtein="
                            planNum(pending, error, data?.day.plan, 'protein')
                        "
                        :plannedFat="
                            planNum(pending, error, data?.day.plan, 'fat')
                        "
                        :plannedCarbs="
                            planNum(pending, error, data?.day.plan, 'carbs')
                        "
                    />
                </div>
                <div
                    v-show="macroPane === 1"
                    class="box-border flex flex-col gap-1.5 px-0.5 pt-1"
                >
                    <div v-if="waterError" class="text-xs text-red-400">
                        {{ waterQueryError?.message ?? "Failed to load water" }}
                    </div>
                    <p
                        v-else-if="waterPending"
                        class="m-0 text-xs text-textSecondary"
                    >
                        Loading water…
                    </p>
                    <div v-else class="flex flex-row gap-2">
                        <button
                            type="button"
                            class="min-w-0 flex-1 cursor-pointer rounded text-left transition-opacity hover:opacity-90 focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[rgb(26,26,26)] m-0! p-0!"
                            :aria-label="'Water log for ' + waterDayLabelLong"
                            title="View today's water entries"
                            @click="openWaterLogsDialog"
                        >
                            <MacroBar
                                type="water"
                                class="min-w-0 flex-1"
                                :total="waterTotalDisplay"
                                :planned="waterPlannedDisplay"
                                :value-suffix="waterUnitSuffix"
                            />
                        </button>
                        <MacroBar
                            type="fiber"
                            class="min-w-0 flex-1"
                            :total="
                                getTotalNumber(pending, error, data?.totalFiber)
                            "
                            :planned="
                                planNum(pending, error, data?.day.plan, 'fiber')
                            "
                        />
                    </div>
                </div>
                <div class="mt-2 flex justify-center gap-3">
                    <button
                        type="button"
                        role="tab"
                        :aria-selected="macroPane === 0"
                        class="rounded-full border-none p-2 transition-colors"
                        :class="
                            macroPane === 0
                                ? 'bg-textSecondary opacity-90'
                                : 'bg-secondBg opacity-55 hover:opacity-80'
                        "
                        aria-label="Macros overview"
                        @click="macroPane = 0"
                    />
                    <button
                        type="button"
                        role="tab"
                        :aria-selected="macroPane === 1"
                        class="rounded-full border-none p-2 transition-colors"
                        :class="
                            macroPane === 1
                                ? 'bg-textSecondary opacity-90'
                                : 'bg-secondBg opacity-55 hover:opacity-80'
                        "
                        aria-label="Water and fiber"
                        @click="macroPane = 1"
                    />
                </div>
            </div>
            <div class="flex w-full flex-col gap-8 sm:flex-row sm:gap-6 pb-16">
                <LoggedMealsDisplay :date-offset="dateOffset" />
                <PlannedMealsDisplay :date-offset="dateOffset" />
            </div>
        </div>
    </div>
</template>
