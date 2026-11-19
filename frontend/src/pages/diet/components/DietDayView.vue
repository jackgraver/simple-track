<script setup lang="ts">
import MacroBar from "~/pages/diet/components/macrobar/MacroBar.vue";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import WaterLogsDialog from "~/pages/diet/components/dialog/WaterLogsDialog.vue";
import { useDietLogsToday } from "~/api/diet/queries";
import PlannedMealsDisplay from "./PlannedMealsDisplay.vue";
import LoggedMealsDisplay from "./LoggedMealsDisplay.vue";
import { useWaterLogs } from "~/api/tracking/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { formatDateLong, ymdForDayOffset } from "~/utils/dateUtil";
import { computed, ref } from "vue";
import {
    Button,
    Card,
    ErrorState,
    SegmentedControl,
    Skeleton,
    useDelayedFlag,
} from "~/shared/ui";

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

const macroPane = ref<"macros" | "water">("macros");
const paneOptions = [
    { label: "Macros", value: "macros" as const },
    { label: "Water & fiber", value: "water" as const },
];

const {
    data,
    isPending: pending,
    error,
    refetch,
} = useDietLogsToday(() => props.dateOffset);

const showSkeleton = useDelayedFlag(pending);

const dateStr = computed(() => ymdForDayOffset(props.dateOffset));
const { goalOz, displayUnit, formatVolumeFromOz } = useWaterPrefs();
const {
    data: waterLogs,
    isPending: waterPending,
    isError: waterError,
    error: waterQueryError,
    refetch: refetchWater,
} = useWaterLogs(dateStr);

const showWaterSkeleton = useDelayedFlag(waterPending);

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
    <div class="flex w-full flex-col gap-6 pb-above-tabbar lg:pb-10">
        <Card>
            <div class="flex flex-col gap-4">
                <SegmentedControl
                    v-model="macroPane"
                    :options="paneOptions"
                    label="Macros and hydration summary"
                />
                <ErrorState
                    v-if="error"
                    compact
                    title="Couldn't load this day"
                    :error="error"
                    @retry="refetch()"
                />
                <div
                    v-else-if="pending"
                    class="grid grid-cols-2 gap-x-4 gap-y-3 lg:grid-cols-4"
                >
                    <Skeleton
                        v-for="n in showSkeleton ? 4 : 0"
                        :key="n"
                        shape="block"
                        height="3.5rem"
                    />
                </div>
                <MacroBars
                    v-else-if="macroPane === 'macros'"
                    :total-calories="getTotalNumber(pending, error, data?.totalCalories)"
                    :total-protein="getTotalNumber(pending, error, data?.totalProtein)"
                    :total-fat="getTotalNumber(pending, error, data?.totalFat)"
                    :total-carbs="getTotalNumber(pending, error, data?.totalCarbs)"
                    :planned-calories="planNum(pending, error, data?.day.plan, 'calories')"
                    :planned-protein="planNum(pending, error, data?.day.plan, 'protein')"
                    :planned-fat="planNum(pending, error, data?.day.plan, 'fat')"
                    :planned-carbs="planNum(pending, error, data?.day.plan, 'carbs')"
                />
                <div v-else class="flex flex-col gap-3">
                    <ErrorState
                        v-if="waterError"
                        compact
                        title="Couldn't load water"
                        :error="waterQueryError"
                        @retry="refetchWater()"
                    />
                    <div
                        v-else-if="waterPending"
                        class="grid grid-cols-2 gap-x-4"
                    >
                        <Skeleton
                            v-for="n in showWaterSkeleton ? 2 : 0"
                            :key="n"
                            shape="block"
                            height="3.5rem"
                        />
                    </div>
                    <div v-else class="grid grid-cols-2 gap-x-4">
                        <MacroBar
                            type="water"
                            :total="waterTotalDisplay"
                            :planned="waterPlannedDisplay"
                            :value-suffix="waterUnitSuffix"
                        />
                        <MacroBar
                            type="fiber"
                            :total="getTotalNumber(pending, error, data?.totalFiber)"
                            :planned="planNum(pending, error, data?.day.plan, 'fiber')"
                        />
                    </div>
                    <div v-if="!waterError && !waterPending">
                        <Button
                            variant="secondary"
                            size="sm"
                            :aria-label="'Water log for ' + waterDayLabelLong"
                            @click="openWaterLogsDialog"
                        >
                            View water entries
                        </Button>
                    </div>
                </div>
            </div>
        </Card>
        <div class="grid grid-cols-1 items-start gap-6 lg:grid-cols-2">
            <PlannedMealsDisplay :date-offset="dateOffset" />
            <LoggedMealsDisplay :date-offset="dateOffset" />
        </div>
    </div>
</template>
