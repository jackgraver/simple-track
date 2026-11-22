<script setup lang="ts">
import { computed } from "vue";
import type { WaterLog } from "~/api/tracking/types";
import { useDeleteWater, useWaterLogs } from "~/api/tracking/queries";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { Trash2 } from "lucide-vue-next";
import { EmptyState, ErrorState, IconButton, List, ListRow, SkeletonList, useDelayedFlag } from "~/shared/ui";

const props = defineProps<{
    dateStr: string;
    dayLabel: string;
}>();

const {
    data: logs,
    isPending,
    isError,
    error,
    refetch,
} = useWaterLogs(() => props.dateStr);
const showSkeleton = useDelayedFlag(isPending);
const deleteMutation = useDeleteWater();

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

/** Rows whose create is still in flight have no server id yet (negative temp id). */
function isUnsaved(log: WaterLog): boolean {
    return log.ID < 0;
}

/** Any row can be deleted: one still being saved is deleted by client_id once its create is through. */
function removeLog(log: WaterLog) {
    deleteMutation.mutate({ id: log.ID, date: props.dateStr, client_id: log.client_id });
}
</script>

<template>
    <div class="flex flex-col gap-3 text-left">
        <p class="m-0 text-xs text-fg-muted">{{ dayLabel }}</p>
        <ErrorState
            v-if="isError"
            compact
            title="Couldn't load water"
            :error="error"
            @retry="refetch()"
        />
        <SkeletonList v-else-if="showSkeleton" :rows="3" :leading="false" />
        <EmptyState
            v-else-if="!isPending && !sortedLogs.length"
            title="No water logged"
            description="Use Log Water to add an entry for this day."
        />
        <List v-else-if="sortedLogs.length">
            <ListRow
                v-for="log in sortedLogs"
                :key="log.ID"
                :pending="isUnsaved(log)"
            >
                <span class="font-semibold tabular-nums">{{ amountLine(log) }}</span>
                <template #subtitle>{{ formatTime(log) }}</template>
                <template #end>
                    <IconButton
                        variant="danger"
                        :label="`Delete water entry ${amountLine(log)}`"
                        @click="removeLog(log)"
                    >
                        <Trash2 class="size-5" />
                    </IconButton>
                </template>
            </ListRow>
        </List>
    </div>
</template>
