<script setup lang="ts">
import { computed } from "vue";
import { TriangleAlert } from "lucide-vue-next";
import { useDietLogsToday } from "~/api/diet/queries";
import { useWaterLogs } from "~/api/tracking/queries";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { formatDateLong, ymdForDayOffset } from "~/utils/dateUtil";
import { Button, Card, ErrorState, Skeleton, useDelayedFlag } from "~/shared/ui";
import { formatInt, isMeaningfullyOver, remaining, type MacroKey } from "./dietDay";
import WaterQuickLog from "./WaterQuickLog.vue";

const props = defineProps<{ dateOffset: number }>();

const { data, isPending, error, refetch } = useDietLogsToday(() => props.dateOffset);
const showSkeleton = useDelayedFlag(isPending);

const dateStr = computed(() => ymdForDayOffset(props.dateOffset));
const dayLabelLong = computed(() => formatDateLong(`${dateStr.value}T12:00:00`));
const dayShort = computed(() =>
    new Date(`${dateStr.value}T12:00:00`).toLocaleDateString(undefined, {
        weekday: "short",
        month: "short",
        day: "numeric",
    }),
);
const overline = computed(() => (props.dateOffset === 0 ? "Left today" : `Left on ${dayShort.value}`));

const plan = computed(() => data.value?.day.plan);
const eaten = computed(() => ({
    calories: data.value?.totalCalories ?? 0,
    protein: data.value?.totalProtein ?? 0,
    carbs: data.value?.totalCarbs ?? 0,
    fat: data.value?.totalFat ?? 0,
    fiber: data.value?.totalFiber ?? 0,
}));
const target = (k: MacroKey) => plan.value?.[k] ?? 0;

const kcal = computed(() => remaining(eaten.value.calories, target("calories")));
const kcalOver = computed(() => isMeaningfullyOver(eaten.value.calories, target("calories")));

type Row = { key: MacroKey; label: string; fill: string };
const rows: Row[] = [
    { key: "protein", label: "Protein", fill: "bg-macro-protein" },
    { key: "carbs", label: "Carbs", fill: "bg-macro-carbs" },
    { key: "fat", label: "Fat", fill: "bg-macro-fat" },
];
const rowState = (k: MacroKey) => {
    const r = remaining(eaten.value[k], target(k));
    return { ...r, warn: isMeaningfullyOver(eaten.value[k], target(k)) };
};

// Water lives in tracking, with its own goal in local prefs.
const { goalOz, formatVolumeFromOz } = useWaterPrefs();
const water = useWaterLogs(dateStr);
const waterOz = computed(() => (water.data.value ?? []).reduce((s, row) => s + row.amount_oz, 0));
const waterUnit = computed(() => formatVolumeFromOz(goalOz.value).unit);
const waterFill = computed(() => remaining(waterOz.value, goalOz.value).fill);
const fiber = computed(() => remaining(eaten.value.fiber, target("fiber")));
</script>

<template>
    <Card>
        <ErrorState v-if="error" compact title="Couldn't load this day" :error="error" @retry="refetch()" />
        <div v-else-if="isPending" class="flex flex-col gap-4" aria-busy="true">
            <template v-if="showSkeleton">
                <Skeleton width="8rem" />
                <Skeleton shape="block" height="3rem" />
                <Skeleton v-for="n in 3" :key="n" shape="line" />
            </template>
        </div>
        <div v-else class="flex flex-col gap-5">
            <!-- Calories: the one number this page leads with. -->
            <div class="flex flex-col gap-2">
                <div class="flex items-baseline justify-between gap-3">
                    <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                        {{ kcal.noTarget ? dayShort : kcalOver ? "Over target" : overline }}
                    </p>
                    <p v-if="plan?.name" class="m-0 truncate text-xs text-fg-muted">{{ plan.name }}</p>
                </div>
                <div class="flex items-baseline justify-between gap-3">
                    <p class="m-0 flex items-baseline gap-2" :class="kcalOver ? 'text-warn' : 'text-fg'">
                        <TriangleAlert v-if="kcalOver" class="size-5 self-center" aria-hidden="true" />
                        <span class="text-display font-semibold tracking-tight tabular-nums">
                            {{ formatInt(kcal.noTarget ? eaten.calories : kcal.over > 0 ? kcal.over : kcal.left) }}
                        </span>
                        <span class="text-sm text-fg-muted">
                            {{ kcal.noTarget ? "kcal eaten" : kcal.over > 0 ? "kcal over" : "kcal left" }}
                        </span>
                    </p>
                    <p v-if="!kcal.noTarget" class="m-0 text-sm text-fg-muted tabular-nums">
                        {{ formatInt(eaten.calories) }} / {{ formatInt(target("calories")) }}
                    </p>
                </div>
                <div
                    class="h-2 overflow-hidden rounded-full bg-surface-3"
                    role="progressbar"
                    aria-label="Calories eaten"
                    aria-valuemin="0"
                    aria-valuemax="100"
                    :aria-valuenow="Math.round(kcal.fill)"
                >
                    <div
                        class="h-full rounded-full"
                        :class="kcalOver ? 'bg-warn' : 'bg-macro-calories'"
                        :style="{ width: `${kcal.fill}%` }"
                    />
                </div>
                <p v-if="kcal.noTarget" class="m-0 text-sm text-fg-muted">
                    No targets for this day.
                    <Button variant="ghost" size="sm" :to="{ name: 'diet-targets' }">Set targets</Button>
                </p>
            </div>

            <!-- Macros: what's left of each, eaten / target on the right. -->
            <ul class="m-0 flex list-none flex-col gap-3 p-0">
                <li v-for="row in rows" :key="row.key" class="flex flex-col gap-1">
                    <div class="flex items-baseline justify-between gap-3 text-sm">
                        <span class="text-fg">
                            {{ row.label }}
                            <span
                                v-if="!rowState(row.key).noTarget"
                                class="tabular-nums"
                                :class="rowState(row.key).warn ? 'text-warn' : 'text-fg-muted'"
                            >
                                ·
                                {{
                                    rowState(row.key).over > 0
                                        ? `${formatInt(rowState(row.key).over)} g over`
                                        : `${formatInt(rowState(row.key).left)} g left`
                                }}
                            </span>
                        </span>
                        <span class="text-xs text-fg-muted tabular-nums">
                            {{ formatInt(eaten[row.key]) }} / {{ formatInt(target(row.key)) }} g
                        </span>
                    </div>
                    <div class="h-1.5 overflow-hidden rounded-full bg-surface-3" aria-hidden="true">
                        <div
                            class="h-full rounded-full"
                            :class="rowState(row.key).warn ? 'bg-warn' : row.fill"
                            :style="{ width: `${rowState(row.key).fill}%` }"
                        />
                    </div>
                </li>
            </ul>

            <!-- Water and fiber: secondary goals, one-tap water underneath. -->
            <div class="flex flex-col gap-3 border-t border-line pt-4">
                <div class="grid grid-cols-2 gap-4">
                    <div class="flex flex-col gap-1">
                        <div class="flex items-baseline justify-between gap-2 text-sm">
                            <span class="text-fg">Water</span>
                            <span v-if="water.isError.value" class="text-xs text-danger">Couldn't load</span>
                            <span v-else-if="!water.isPending.value" class="text-xs text-fg-muted tabular-nums">
                                {{ formatVolumeFromOz(waterOz).value }} / {{ formatVolumeFromOz(goalOz).value }}
                                {{ waterUnit }}
                            </span>
                        </div>
                        <div class="h-1.5 overflow-hidden rounded-full bg-surface-3" aria-hidden="true">
                            <div class="h-full rounded-full bg-macro-water" :style="{ width: `${waterFill}%` }" />
                        </div>
                    </div>
                    <div class="flex flex-col gap-1">
                        <div class="flex items-baseline justify-between gap-2 text-sm">
                            <span class="text-fg">Fiber</span>
                            <span class="text-xs text-fg-muted tabular-nums">
                                {{ formatInt(eaten.fiber) }} / {{ formatInt(target("fiber")) }} g
                            </span>
                        </div>
                        <div class="h-1.5 overflow-hidden rounded-full bg-surface-3" aria-hidden="true">
                            <div class="h-full rounded-full bg-macro-fiber" :style="{ width: `${fiber.fill}%` }" />
                        </div>
                    </div>
                </div>
                <ErrorState
                    v-if="water.isError.value"
                    compact
                    title="Couldn't load water"
                    :error="water.error.value"
                    @retry="water.refetch()"
                />
                <WaterQuickLog v-else :date-str="dateStr" :day-label="dayLabelLong" />
            </div>
        </div>
    </Card>
</template>
