<script setup lang="ts">
import { RotateCcw, X } from "lucide-vue-next";
import { IconButton, List, ListRow } from "~/shared/ui";
import type { LoggedSetWithStatus } from "../../store/useWorkoutStore";
import type { ProgressVerdict } from "../domain/liveLogging";

// Below the fold: every set today next to the same set last session, and whether the weight goes up next time.
// Tap a saved set to edit it with the slider above; tap it again to go back to the current set.
const props = defineProps<{
    sets: LoggedSetWithStatus[];
    editingIndex: number | null;
    previous: { weight: number; reps: number }[];
    verdict: ProgressVerdict | null;
    weight: number;
}>();

const emit = defineEmits<{
    (e: "select", index: number, row: HTMLElement): void;
    (e: "delete", index: number): void;
    (e: "retry", index: number): void;
}>();

function diff(index: number): { text: string; tone: string } {
    const set = props.sets[index];
    const prev = props.previous[index];
    if (!set || !prev) return { text: "new", tone: "text-fg-subtle" };
    if (set.weight !== prev.weight) {
        const d = set.weight - prev.weight;
        return { text: `${d > 0 ? "+" : "−"}${Math.abs(d)} lb`, tone: d > 0 ? "text-success" : "text-fg-muted" };
    }
    const d = set.reps - prev.reps;
    if (d === 0) return { text: "=", tone: "text-fg-subtle" };
    return { text: `${d > 0 ? "+" : "−"}${Math.abs(d)} rep${Math.abs(d) === 1 ? "" : "s"}`, tone: d > 0 ? "text-success" : "text-warn" };
}
</script>

<template>
    <section v-if="sets.length" class="flex flex-col gap-2" aria-label="Today vs last time">
        <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
            <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Today vs last time</p>
            <p v-if="verdict" class="m-0 text-sm" :class="verdict.onTrack ? 'text-success' : 'text-fg-muted'">
                {{ verdict.hit }}/{{ verdict.total }} sets hit {{ verdict.rollover }}
                {{ verdict.onTrack ? "· on track to go up" : `· stay at ${weight} next time` }}
            </p>
        </div>
        <List class="rounded-card border border-line bg-surface">
            <ListRow
                v-for="(set, index) in sets"
                :key="set.tempId || index"
                :pending="set.status === 'pending'"
                :selected="editingIndex === index"
                v-bind="
                    set.status === 'success'
                        ? { onClick: (ev: MouseEvent) => emit('select', index, ev.currentTarget as HTMLElement) }
                        : {}
                "
            >
                <span class="tabular-nums">Set {{ index + 1 }} · {{ set.weight }} × {{ set.reps }}</span>
                <template #subtitle>
                    <span class="tabular-nums">
                        Last {{ previous[index] ? `${previous[index].weight} × ${previous[index].reps}` : "—" }}
                    </span>
                </template>
                <template #trailing>
                    <span v-if="set.status === 'error'" class="text-xs text-danger">Didn't save</span>
                    <span v-else-if="editingIndex === index" class="text-xs font-medium text-accent">Editing</span>
                    <span v-else class="text-sm font-medium tabular-nums" :class="diff(index).tone">{{ diff(index).text }}</span>
                </template>
                <template v-if="set.status === 'error'" #end>
                    <IconButton :label="`Retry saving set ${index + 1}`" variant="danger" @click="emit('retry', index)">
                        <RotateCcw />
                    </IconButton>
                </template>
                <template v-else-if="set.status === 'success'" #end>
                    <IconButton :label="`Delete set ${index + 1}`" variant="danger" @click="emit('delete', index)"><X /></IconButton>
                </template>
            </ListRow>
        </List>
    </section>
</template>
