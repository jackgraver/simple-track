<script setup lang="ts">
import axios from "axios";
import { computed } from "vue";
import LogWaterDialog from "~/pages/diet/components/dialog/LogWaterDialog.vue";
import {
    useDeleteWater,
    useWaterLogs,
} from "~/api/tracking/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { ymdForDayOffset } from "~/utils/dateUtil";

const props = defineProps<{
    dateOffset: number;
}>();

const dateStr = computed(() => ymdForDayOffset(props.dateOffset));
const { goalOz, formatVolumeFromOz } = useWaterPrefs();
const { data: logs, isPending, isError, error } = useWaterLogs(dateStr);
const deleteMutation = useDeleteWater();

const totalOz = computed(() =>
    (logs.value ?? []).reduce((s, row) => s + row.amount_oz, 0),
);

const totalLabel = computed(() => {
    const f = formatVolumeFromOz(totalOz.value);
    return `${f.value} ${f.unit}`;
});

const goalLabel = computed(() => {
    const f = formatVolumeFromOz(goalOz.value);
    return `${f.value} ${f.unit}`;
});

function barWidthPct(): number {
    const g = goalOz.value;
    const t = totalOz.value;
    if (g <= 0) return t > 0 ? 100 : 0;
    return Math.min(100, (t / g) * 100);
}

async function openLog() {
    await dialogManager.custom<boolean>({
        title: "Log water",
        component: LogWaterDialog,
        componentProps: { dateStr: dateStr.value },
    });
}

async function removeEntry(id: number) {
    const ok = await dialogManager.confirm({
        title: "Remove drink?",
        message: "This entry will be removed from today's log.",
        confirmText: "Remove",
        cancelText: "Cancel",
    });
    if (!ok) return;
    try {
        await deleteMutation.mutateAsync({ id, date: dateStr.value });
        toast.push("Removed", "success");
    } catch (err: unknown) {
        let msg = "Failed to remove";
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
}

function rowLabel(row: { amount_oz: number; preset?: { name: string } | null }) {
    const f = formatVolumeFromOz(row.amount_oz);
    const amt = `${f.value} ${f.unit}`;
    if (row.preset?.name) return `${row.preset.name} (${amt})`;
    return amt;
}
</script>

<template>
    <section
        class="flex flex-col gap-3 rounded-md border border-zinc-700 bg-zinc-900/60 p-4"
    >
        <div class="flex flex-wrap items-center justify-between gap-2">
            <h2 class="m-0 text-sm font-semibold tracking-tight text-zinc-100">
                Water
            </h2>
            <div class="flex flex-wrap items-center gap-2">
                <router-link
                    :to="{ name: 'diet-water-presets' }"
                    class="text-xs font-medium text-sky-400 underline-offset-2 hover:text-sky-300 hover:underline"
                >
                    Sizes & goal
                </router-link>
                <button
                    type="button"
                    class="rounded-md border border-sky-700/60 bg-sky-950/40 px-3 py-1.5 text-xs font-semibold text-sky-100 hover:bg-sky-950/70"
                    @click="openLog"
                >
                    Log water
                </button>
            </div>
        </div>
        <div
            class="overflow-hidden rounded border border-zinc-700 bg-zinc-950"
            role="progressbar"
            :aria-valuenow="Math.round(barWidthPct())"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="`Water progress ${totalLabel} of ${goalLabel}`"
        >
            <div
                class="min-h-[22px] bg-sky-600 px-2 py-1 text-center text-xs font-bold tabular-nums text-white transition-[width] duration-300"
                :style="{ width: `${barWidthPct()}%` }"
            >
                <span class="whitespace-nowrap">{{ totalLabel }} / {{ goalLabel }}</span>
            </div>
        </div>
        <div v-if="isError" class="text-xs text-red-400">
            {{ error?.message ?? "Failed to load water log" }}
        </div>
        <p v-else-if="isPending" class="m-0 text-xs text-zinc-500">Loading water…</p>
        <ul v-else class="m-0 flex list-none flex-col gap-1 p-0">
            <li
                v-for="row in logs ?? []"
                :key="row.ID"
                class="flex items-center justify-between gap-2 rounded border border-zinc-800 bg-zinc-900/80 px-2 py-1.5 text-xs text-zinc-200"
            >
                <span class="min-w-0 truncate">{{ rowLabel(row) }}</span>
                <button
                    type="button"
                    class="shrink-0 rounded px-2 py-0.5 text-xs text-zinc-400 hover:bg-zinc-800 hover:text-red-400"
                    @click="removeEntry(row.ID)"
                >
                    Remove
                </button>
            </li>
            <li v-if="!(logs ?? []).length" class="text-xs text-zinc-500">
                Nothing logged for this day yet.
            </li>
        </ul>
    </section>
</template>
