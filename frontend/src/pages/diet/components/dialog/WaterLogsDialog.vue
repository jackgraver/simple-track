<script setup lang="ts">
import axios from "axios";
import { computed, ref } from "vue";
import type { WaterLog } from "~/api/tracking/types";
import { useDeleteWater, useWaterLogs } from "~/api/tracking/queries";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    dateStr: string;
    dayLabel: string;
}>();

const { data: logs, isPending } = useWaterLogs(() => props.dateStr);
const deleteMutation = useDeleteWater();
const deletingId = ref<number | null>(null);

const { displayUnit, formatVolumeFromOz } = useWaterPrefs();

const timeFmt = new Intl.DateTimeFormat(undefined, { timeStyle: "short" });

function createdMs(log: WaterLog): number {
    const raw = log.CreatedAt;
    if (raw) {
        const t = Date.parse(raw);
        if (!Number.isNaN(t)) return t;
    }
    return log.ID;
}

function formatTime(log: WaterLog): string {
    const raw = log.CreatedAt;
    if (!raw) return "—";
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) return "—";
    return timeFmt.format(d);
}

function amountLine(log: WaterLog): string {
    const fv = formatVolumeFromOz(log.amount_oz);
    return `${fv.value} ${displayUnit.value}`;
}

const sortedLogs = computed(() =>
    [...(logs.value ?? [])].sort((a, b) => createdMs(a) - createdMs(b)),
);

function errMsg(err: unknown): string {
    if (
        axios.isAxiosError(err) &&
        err.response?.data &&
        typeof err.response.data === "object" &&
        "error" in err.response.data
    ) {
        const e0 = (err.response.data as { error?: string }).error;
        if (e0) return e0;
    }
    if (err instanceof Error) return err.message;
    return "Failed to delete";
}

async function removeLog(log: WaterLog) {
    if (deletingId.value != null || deleteMutation.isPending.value) return;
    deletingId.value = log.ID;
    try {
        await deleteMutation.mutateAsync({ id: log.ID, date: props.dateStr });
        toast.push("Removed", "success");
    } catch (err: unknown) {
        toast.push(errMsg(err), "error");
    } finally {
        deletingId.value = null;
    }
}
</script>

<template>
    <div class="flex flex-col gap-3 text-left text-zinc-100">
        <p class="m-0 text-xs mx-1 text-zinc-400">{{ dayLabel }}</p>
        <p v-if="isPending" class="m-0 text-sm text-zinc-500">Loading…</p>
        <p v-else-if="!sortedLogs.length" class="m-0 text-sm text-zinc-500">
            No water logged for this day.
        </p>
        <ul v-else class="m-0 list-none space-y-1.5 p-0">
            <li v-for="log in sortedLogs" :key="log.ID">
                <button
                    type="button"
                    class="flex w-full cursor-pointer items-baseline justify-between gap-4 rounded border! border-transparent! py-2 px-1! text-left text-sm transition-colors hover:border-red-500! shadow-none! m-0!"
                    :disabled="
                        deletingId === log.ID || deleteMutation.isPending.value
                    "
                    :aria-label="`Delete water entry ${amountLine(log)}`"
                    @click="removeLog(log)"
                >
                    <span class="font-semibold tabular-nums text-zinc-100">{{
                        amountLine(log)
                    }}</span>
                    <span class="shrink-0 tabular-nums text-zinc-400">{{
                        formatTime(log)
                    }}</span>
                </button>
            </li>
        </ul>
    </div>
</template>
