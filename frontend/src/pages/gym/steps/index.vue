<script setup lang="ts">
import Chart from "~/shared/chart/Chart.vue";
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { useStepLogs, useSaveSteps } from "~/api/tracking/queries";
import {
    formatDateLong,
    parseDietDayOffsetQuery,
    ymdForDayOffset,
} from "~/utils/dateUtil";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import {
    buildTrackingBarChart,
    seriesLastDaysThroughSelectedDay,
    trackingDayKey,
} from "~/pages/gym/trackingLogChart";
import { Button, Card, ErrorState, Field, Input, PageHeader, Section, Skeleton, useDelayedFlag } from "~/shared/ui";

const route = useRoute();
const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const dateStr = computed(() => ymdForDayOffset(dateOffset.value));
const gymHomeQuery = computed(() =>
    dateOffset.value === 0 ? {} : { offset: String(dateOffset.value) },
);
const stepsInput = ref("");
const { data: logs, isPending, isError, error, refetch } = useStepLogs();
const showSkeleton = useDelayedFlag(() => isPending.value);
const saveMutation = useSaveSteps();
const rows = computed(() => logs.value ?? []);
const existingLogForDate = computed(() =>
    rows.value.find((r) => trackingDayKey(r.date) === dateStr.value),
);
const stepSeries = computed(() =>
    seriesLastDaysThroughSelectedDay(
        rows.value.map((r) => ({ date: r.date, value: r.steps })),
        dateStr.value,
    ),
);
const stepChart = computed(() => {
    const s = stepSeries.value;
    return buildTrackingBarChart(s.labels, s.values, "Steps");
});
const handleSave = async () => {
    const n = Number.parseInt(stepsInput.value.replace(/\s/g, ""), 10);
    if (Number.isNaN(n) || n < 0) {
        toast.push("Enter a valid step count", "error");
        return;
    }
    const existing = existingLogForDate.value;
    if (existing) {
        const dateLabel = formatDateLong(`${dateStr.value}T12:00:00`);
        const confirmed = await dialogManager.confirm({
            title: "Overwrite steps?",
            message: `You already logged ${existing.steps.toLocaleString()} steps for ${dateLabel}. Are you sure you want to overwrite?`,
            confirmText: "Overwrite",
        });
        if (!confirmed) return;
    }
    // Optimistic: the chart and list update now; a failure rolls back with a Retry toast.
    saveMutation.mutate({ date: dateStr.value, steps: n });
    toast.push("Saved", "success");
        stepsInput.value = "";
};
</script>
<template>
    <div class="flex w-full min-w-0 flex-col gap-6">
        <PageHeader title="Steps">
            <template #actions>
                <Button variant="ghost" :to="{ name: 'gym', query: gymHomeQuery }">Back</Button>
            </template>
        </PageHeader>
        <Card>
            <form class="flex flex-col gap-3" @submit.prevent="handleSave">
                <Field :label="`Steps (${formatDateLong(`${dateStr}T12:00:00`)})`">
                    <Input v-model="stepsInput" inputmode="numeric" autocomplete="off" placeholder="e.g. 8420" />
                </Field>
                <Button type="submit" variant="primary">Save</Button>
            </form>
        </Card>
        <ErrorState v-if="isError" :error="error" title="Couldn't load your step history" @retry="refetch()" />
        <template v-else-if="isPending">
            <Skeleton v-if="showSkeleton" shape="block" height="18rem" />
        </template>
        <Section v-else title="Previous">
            <div class="h-72 min-h-72 w-full min-w-0 lg:h-80 lg:min-h-80">
                <Chart type="bar" :data="stepChart.data" :options="stepChart.options" class="h-full w-full" />
            </div>
        </Section>
    </div>
</template>
