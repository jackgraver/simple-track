<script setup lang="ts">
import { computed } from "vue";
import { RotateCcw } from "lucide-vue-next";
import type { SetSide } from "~/types/workout";
import { Button, Card, IconButton } from "~/shared/ui";
import { useOnline } from "~/api/online";
import type { LoggedSetWithStatus } from "../../store/useWorkoutStore";
import { formatWeight, sideLabel, type SetPair } from "../domain/unilateral";

// One unilateral set number: a folder tab ("Set 1 · 30 lb") over a card split into Left | Right.
// Each half is one big tap target that edits that side. Deleting a side lives in edit mode.
type Last = { weight: number; reps: number };

const props = defineProps<{
    number: number;
    pair: SetPair<LoggedSetWithStatus>;
    /** Last session's set for each side (a bilateral set counts for both). */
    last: { left?: Last; right?: Last };
    editingIndex: number | null;
}>();

const emit = defineEmits<{
    (e: "select", index: number, el: HTMLElement): void;
    (e: "retry", index: number): void;
}>();

const online = useOnline();

const halves = computed(() =>
    (["left", "right"] as const).map((side) => ({ side, entry: props.pair[side] ?? props.pair.both })),
);
const sharedWeight = computed(() => {
    const sets = halves.value.map((h) => h.entry?.set).filter((s) => s != null);
    return sets.every((s) => s.weight === sets[0]?.weight) ? (sets[0]?.weight ?? null) : null;
});
const failed = computed(() => halves.value.filter((h) => h.entry?.set.status === "error"));

function diff(set: Last, last: Last | undefined): { text: string; tone: string } {
    if (!last) return { text: "new", tone: "text-fg-subtle" };
    if (set.weight !== last.weight) {
        const d = set.weight - last.weight;
        return { text: `${d > 0 ? "+" : "−"}${Math.abs(d)} lb`, tone: d > 0 ? "text-success" : "text-fg-muted" };
    }
    const d = set.reps - last.reps;
    if (d === 0) return { text: "=", tone: "text-fg-subtle" };
    return { text: `${d > 0 ? "+" : "−"}${Math.abs(d)}`, tone: d > 0 ? "text-success" : "text-warn" };
}

function lastText(side: SetSide): string {
    const l = props.last[side];
    if (!l) return "Last —";
    return sharedWeight.value != null && l.weight === sharedWeight.value ? `Last ${l.reps}` : `Last ${formatWeight(l.weight)} × ${l.reps}`;
}

function select(index: number, ev: MouseEvent) {
    const card = (ev.currentTarget as HTMLElement).closest("[data-set-card]") as HTMLElement | null;
    emit("select", index, card ?? (ev.currentTarget as HTMLElement));
}
</script>

<template>
    <div class="relative" data-set-card>
        <Card padding="none" :aria-label="`Set ${number}`" as="section">
            <template #tab>
                <span class="tabular-nums">Set {{ number }}<template v-if="sharedWeight != null"> · {{ formatWeight(sharedWeight) }}</template></span>
            </template>
            <!-- Each side is a sunken tile: inset, on a darker fill, with the card showing around it. No borders. -->
            <div class="grid grid-cols-2 gap-1.5 p-1.5">
                <template v-for="h in halves" :key="h.side">
                    <!-- The tile look lives on a wrapper: Button's own (transparent) fill and border win over
                         classes passed to it. The ghost Button fills the tile as the tap target. -->
                    <div
                        v-if="h.entry"
                        class="flex rounded-control"
                        :class="[
                            editingIndex === h.entry.index ? 'bg-surface-3' : 'bg-canvas',
                            h.entry.set.status === 'pending' ? 'opacity-60' : '',
                        ]"
                    >
                        <Button
                            variant="ghost"
                            block
                            class="min-h-16 justify-start px-3 py-2 text-left"
                            :aria-pressed="editingIndex === h.entry.index"
                            :aria-label="`Edit set ${number} ${h.side}`"
                            @click="(ev: MouseEvent) => select(h.entry!.index, ev)"
                        >
                            <span class="flex min-w-0 flex-col gap-0.5 tabular-nums">
                                <span class="text-xs font-semibold tracking-wide uppercase" :class="editingIndex === h.entry.index ? 'text-accent' : 'text-fg-muted'">
                                    {{ editingIndex === h.entry.index ? `Editing ${sideLabel(h.side).toLowerCase()}` : sideLabel(h.side) }}
                                </span>
                                <span class="flex items-baseline gap-2 text-base text-fg">
                                    {{ sharedWeight != null ? `${h.entry.set.reps} reps` : `${formatWeight(h.entry.set.weight)} × ${h.entry.set.reps}` }}
                                    <span class="text-sm font-medium" :class="diff(h.entry.set, last[h.side]).tone">{{ diff(h.entry.set, last[h.side]).text }}</span>
                                </span>
                                <span v-if="h.entry.set.status === 'error'" class="text-xs text-danger">Didn't save</span>
                                <span v-else-if="h.entry.set.status === 'pending' && !online" class="text-xs text-fg-muted">Waiting for connection</span>
                                <span v-else class="text-xs text-fg-muted">{{ lastText(h.side) }}</span>
                            </span>
                        </Button>
                    </div>
                    <div v-else class="flex min-h-16 flex-col justify-center gap-0.5 rounded-control bg-canvas px-3 py-2">
                        <span class="text-xs font-semibold tracking-wide text-fg-subtle uppercase">{{ sideLabel(h.side) }}</span>
                        <span class="text-sm text-fg-subtle">Up next</span>
                    </div>
                </template>
            </div>
        </Card>
        <!-- Retry sits on the tab line, outside the halves (they're buttons; buttons don't nest). -->
        <div v-if="failed.length" class="absolute top-0 right-0 flex h-8 items-center">
            <IconButton
                :label="`Retry saving set ${number}`"
                variant="danger"
                @click="failed.forEach((h) => emit('retry', h.entry!.index))"
            >
                <RotateCcw />
            </IconButton>
        </div>
    </div>
</template>
