<script setup lang="ts">
import Chart from "primevue/chart";
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
import axios from "axios";
import NumberSlider from "~/pages/gym/logging/exercise/components/NumberSlider.vue";

const WEIGHT_SLIDER_PADDING = 25;

const route = useRoute();
const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const dateStr = computed(() => ymdForDayOffset(dateOffset.value));
const gymHomeQuery = computed(() =>
    dateOffset.value === 0 ? {} : { offset: String(dateOffset.value) },
);
const weightLbs = ref(0);
const { data: logs, isPending, isError, error } = useWeightLogs();
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
const weightLineColor = "hsl(215, 85%, 55%)";
const weightChart = computed(() => {
    const s = weightSeries.value;
    return buildTrackingLineChart(
        s.labels,
        s.values,
        "Weight (lbs)",
        weightLineColor,
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
    try {
        await saveMutation.mutateAsync({ date: dateStr.value, weightLbs: w });
        toast.push("Saved", "success");
    } catch (err: unknown) {
        let msg = "Failed to save";
        if (
            axios.isAxiosError(err) &&
            err.response?.data &&
            typeof err.response.data === "object" &&
            "error" in err.response.data
        ) {
            const e0 = (err.response.data as { error?: string }).error;
            if (e0) msg = e0;
        } else if (err instanceof Error) {
            msg = err.message;
        }
        toast.push(msg, "error");
    }
};
const saving = computed(() => saveMutation.isPending.value);
</script>
<template>
    <div class="flex w-full min-w-0 flex-col gap-6">
        <div
            class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
        >
            <h1 class="m-0 text-lg font-semibold text-textPrimary">
                Body weight
            </h1>
            <router-link
                :to="{ name: 'gym', query: gymHomeQuery }"
                class="text-sm text-textSecondary transition-colors hover:text-textPrimary"
                >Back</router-link
            >
        </div>
        <form
            class="flex flex-col gap-3 rounded-md border border-(--color-border) bg-firstBg p-4"
            @submit.prevent="handleSave"
        >
            <NumberSlider
                v-model="weightLbs"
                :label="weightLabel"
                :step="0.1"
                :min="sliderMin"
                :max="sliderMax"
            />
            <button
                type="submit"
                class="rounded-md bg-secondBg px-4 py-2 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg disabled:opacity-50"
                :disabled="saving"
            >
                {{ saving ? "Saving…" : "Save" }}
            </button>
        </form>
        <div v-if="isError" class="text-sm text-(--color-cf-red)">
            {{ error?.message ?? "Failed to load" }}
        </div>
        <div v-else-if="isPending" class="text-sm text-textSecondary">
            Loading…
        </div>
        <template v-else>
            <div class="flex min-w-0 flex-col gap-2">
                <h2 class="m-0 text-base font-semibold text-textPrimary">
                    Previous
                </h2>
                <div
                    v-if="hasChartData"
                    class="h-72 min-h-72 w-full min-w-0 lg:h-80 lg:min-h-80"
                >
                    <Chart
                        type="line"
                        :data="weightChart.data"
                        :options="weightChart.options"
                        class="h-full w-full"
                    />
                </div>
                <p v-else class="m-0 text-sm text-textSecondary">
                    No history through this day yet.
                </p>
            </div>
        </template>
    </div>
</template>
