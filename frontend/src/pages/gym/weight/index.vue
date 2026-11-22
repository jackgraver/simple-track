<script setup lang="ts">
import Chart from "~/shared/chart/Chart.vue";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
    latestWeightLbs,
    useSaveWeight,
    useWeightLogs,
} from "~/api/tracking/queries";
import {
    formatDateLong,
    parseDietDayOffsetQuery,
    ymdForDayOffset,
} from "~/utils/dateUtil";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import {
    buildTrackingLineChart,
    seriesThroughSelectedDay,
    trackingDayKey,
} from "~/pages/gym/trackingLogChart";
import { ClipboardList } from "lucide-vue-next";
import { openImportWeightDialog } from "./dialog/useImportDialog";
import { Button, Card, ErrorState, PageHeader, Section, Skeleton, useDelayedFlag } from "~/shared/ui";
import NumberSlider from "~/pages/gym/logging/exercise/components/NumberSlider.vue";

const WEIGHT_SLIDER_PADDING = 25;

const route = useRoute();
const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const dateStr = computed(() => ymdForDayOffset(dateOffset.value));
const gymHomeQuery = computed(() =>
    dateOffset.value === 0 ? {} : { offset: String(dateOffset.value) },
);
const weightLbs = ref(0);
const { data: logs, isPending, isError, error, refetch } = useWeightLogs();
const showSkeleton = useDelayedFlag(() => isPending.value);
const saveMutation = useSaveWeight();
const rows = computed(() => logs.value ?? []);
const existingLogForDate = computed(() =>
    rows.value.find((r) => trackingDayKey(r.date) === dateStr.value),
);
const defaultWeightLbs = computed(() => {
    const existing = existingLogForDate.value;
    if (existing) return existing.weight_lbs;
    return latestWeightLbs(logs.value) ?? 0;
});
watch(
    defaultWeightLbs,
    (value) => {
        if (value > 0) weightLbs.value = value;
    },
    { immediate: true },
);
const sliderAnchor = computed(() => {
    const d = defaultWeightLbs.value;
    return d > 0 ? d : 150;
});
const sliderMin = computed(() =>
    Math.max(50, Math.round((sliderAnchor.value - WEIGHT_SLIDER_PADDING) * 10) / 10),
);
const sliderMax = computed(() =>
    Math.min(500, Math.round((sliderAnchor.value + WEIGHT_SLIDER_PADDING) * 10) / 10),
);
const weightLabel = computed(
    () => `Weight in lbs (${formatDateLong(`${dateStr.value}T12:00:00`)})`,
);
const weightSeries = computed(() =>
    seriesThroughSelectedDay(
        rows.value.map((r) => ({ date: r.date, value: r.weight_lbs })),
        dateStr.value,
    ),
);
const hasChartData = computed(() => weightSeries.value.values.length > 0);
const weightChart = computed(() => {
    const s = weightSeries.value;
    return buildTrackingLineChart(
        s.labels,
        s.values,
        "Weight (lbs)",
    );
});
const handleSave = async () => {
    const w = weightLbs.value;
    if (w <= 0) {
        toast.push("Enter a valid weight (lbs)", "error");
        return;
    }
    const existing = existingLogForDate.value;
    if (existing) {
        const dateLabel = formatDateLong(`${dateStr.value}T12:00:00`);
        const confirmed = await dialogManager.confirm({
            title: "Overwrite weight?",
            message: `You already logged ${existing.weight_lbs} lbs for ${dateLabel}. Are you sure you want to overwrite?`,
            confirmText: "Overwrite",
        });
        if (!confirmed) return;
    }
    // Optimistic: the chart and list update now; a failure rolls back with a Retry toast.
    saveMutation.mutate({ date: dateStr.value, weightLbs: w });
    toast.push("Saved", "success");
};

const metricsForDay = computed(() => {
    const l = existingLogForDate.value;
    if (!l) return [];
    const fmt = (v: number | null | undefined, suffix = "") =>
        v == null ? null : `${v}${suffix}`;
    return [
        { label: "Body fat", value: fmt(l.body_fat_pct, "%") },
        { label: "Muscle", value: fmt(l.muscle_pct, "%") },
        { label: "Body water", value: fmt(l.body_water_pct, "%") },
        { label: "BMI", value: fmt(l.bmi) },
        { label: "Visceral fat", value: fmt(l.visceral_fat) },
        { label: "BMR", value: fmt(l.bmr_kcal, " kcal") },
    ].filter((m): m is { label: string; value: string } => m.value != null);
});
</script>
<template>
    <div class="flex w-full min-w-0 flex-col gap-6">
        <PageHeader title="Body weight">
            <template #actions>
                <Button variant="secondary" @click="openImportWeightDialog">
                    <ClipboardList class="size-4" aria-hidden="true" />
                    Import
                </Button>
                <Button variant="ghost" :to="{ name: 'gym', query: gymHomeQuery }">Back</Button>
            </template>
        </PageHeader>
        <Card>
            <form class="flex flex-col gap-3" @submit.prevent="handleSave">
                <NumberSlider v-model="weightLbs" :label="weightLabel" :step="0.1" :min="sliderMin" :max="sliderMax" />
                <Button type="submit" variant="primary">Save</Button>
            </form>
        </Card>
        <Card v-if="metricsForDay.length">
            <div class="grid grid-cols-3 gap-3">
                <div v-for="m in metricsForDay" :key="m.label" class="flex flex-col">
                    <span class="text-xs text-fg-muted">{{ m.label }}</span>
                    <span class="text-sm font-semibold text-fg">{{ m.value }}</span>
                </div>
            </div>
        </Card>
        <ErrorState v-if="isError" :error="error" title="Couldn't load your weight history" @retry="refetch()" />
        <template v-else-if="isPending">
            <Skeleton v-if="showSkeleton" shape="block" height="18rem" />
        </template>
        <Section v-else title="Previous">
            <div v-if="hasChartData" class="h-72 min-h-72 w-full min-w-0 lg:h-80 lg:min-h-80">
                <Chart type="line" :data="weightChart.data" :options="weightChart.options" class="h-full w-full" />
            </div>
            <p v-else class="m-0 text-sm text-fg-muted">No history through this day yet.</p>
        </Section>
    </div>
</template>
