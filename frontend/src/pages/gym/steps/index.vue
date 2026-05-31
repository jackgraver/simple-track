<script setup lang="ts">
import Chart from "primevue/chart";
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
import axios from "axios";

const route = useRoute();
const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const dateStr = computed(() => ymdForDayOffset(dateOffset.value));
const gymHomeQuery = computed(() =>
    dateOffset.value === 0 ? {} : { offset: String(dateOffset.value) },
);
const stepsInput = ref("");
const { data: logs, isPending, isError, error } = useStepLogs();
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
const stepsBarColor = "hsl(142, 71%, 45%)";
const stepChart = computed(() => {
    const s = stepSeries.value;
    return buildTrackingBarChart(s.labels, s.values, "Steps", stepsBarColor);
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
    try {
        await saveMutation.mutateAsync({ date: dateStr.value, steps: n });
        toast.push("Saved", "success");
        stepsInput.value = "";
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
            <h1 class="m-0 text-lg font-semibold text-textPrimary">Steps</h1>
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
            <label class="flex flex-col gap-1 text-xs text-textSecondary"
                >Steps ({{ formatDateLong(`${dateStr}T12:00:00`) }})
                <input
                    v-model="stepsInput"
                    type="text"
                    inputmode="numeric"
                    autocomplete="off"
                    placeholder="e.g. 8420"
                    class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                />
            </label>
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
                <div class="h-72 min-h-72 w-full min-w-0 lg:h-80 lg:min-h-80">
                    <Chart
                        type="bar"
                        :data="stepChart.data"
                        :options="stepChart.options"
                        class="h-full w-full"
                    />
                </div>
            </div>
        </template>
    </div>
</template>
