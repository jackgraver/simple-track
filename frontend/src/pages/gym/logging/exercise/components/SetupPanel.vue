<script setup lang="ts">
import { computed } from "vue";
import { ChevronRight } from "lucide-vue-next";
import { Badge, Card } from "~/shared/ui";
import type { MachineSetting } from "~/types/workout";
import { formatWeightDelta, type OverloadStatus } from "../domain/liveLogging";

// Today's weight + how to set the machine up. Open before the first set (that's when you set up),
// one line after it. The whole card opens the weight & setup sheet.
const props = defineProps<{
    expanded: boolean;
    weight: number;
    overload: OverloadStatus;
    /** Last session's sets, e.g. "100 lb × 10, 10, 9"; empty when there's no history. */
    lastTime: string;
    machineSettings: MachineSetting[];
    /** Plate summary for plate-loaded lifts, null otherwise. */
    plates: string | null;
    cues: string;
}>();

const emit = defineEmits<{ (e: "edit"): void }>();

// Only progress gets a badge: going up is news, staying the same isn't.
const badge = computed(() => {
    const o = props.overload;
    if (o.kind !== "up") return null;
    return { text: `${o.earned ? "↑ " : ""}${formatWeightDelta(o.delta)}`, tone: o.earned ? "success" : "neutral" } as const;
});
const lastLine = computed(() => {
    if (!props.lastTime) return "No history yet";
    const o = props.overload;
    const allHit = o.kind !== "none" && o.earned && o.rollover != null ? ` · all hit ${o.rollover}` : "";
    return `Last ${props.lastTime}${allHit}`;
});
</script>

<template>
    <Card
        as="button"
        type="button"
        interactive
        :padding="expanded ? 'md' : 'sm'"
        :aria-label="`Weight ${weight} pounds. Change weight and setup`"
        @click="emit('edit')"
    >
        <span v-if="expanded" class="flex flex-col gap-4">
            <span class="flex items-start justify-between gap-3">
                <span class="flex min-w-0 flex-col gap-1">
                    <span class="text-xs font-semibold tracking-wide text-fg-muted uppercase">Setup</span>
                    <span class="flex items-baseline gap-2">
                        <span class="text-display font-semibold text-fg tabular-nums">{{ weight }} lb</span>
                        <Badge v-if="badge" :tone="badge.tone">{{ badge.text }}</Badge>
                    </span>
                    <span class="text-sm text-fg-muted tabular-nums">{{ lastLine }}</span>
                </span>
                <ChevronRight class="size-5 shrink-0 text-fg-subtle" aria-hidden="true" />
            </span>
            <span class="grid grid-cols-2 gap-2 sm:grid-cols-3">
                <span v-for="s in machineSettings" :key="s.label" class="flex flex-col rounded-control bg-surface-2 px-3 py-2">
                    <span class="text-2xs tracking-wide text-fg-muted uppercase">{{ s.label }}</span>
                    <span class="text-lg font-semibold text-fg tabular-nums">{{ s.value.trim() || "—" }}</span>
                </span>
                <span v-if="plates" class="flex flex-col rounded-control bg-surface-2 px-3 py-2">
                    <span class="text-2xs tracking-wide text-fg-muted uppercase">Plates</span>
                    <span class="text-sm font-semibold text-fg tabular-nums">{{ plates }}</span>
                </span>
                <span
                    v-if="!machineSettings.length"
                    class="col-span-2 flex flex-col gap-1 rounded-control border border-dashed border-line px-3 py-2 sm:col-span-3"
                >
                    <span class="flex items-center gap-2 text-2xs tracking-wide text-fg-muted uppercase">
                        Machine settings <Badge tone="warn">Preview</Badge>
                    </span>
                    <span class="text-sm text-fg-subtle">Seat and pad positions show here.</span>
                </span>
            </span>
            <span v-if="cues" class="flex flex-col gap-1">
                <span class="text-2xs tracking-wide text-fg-muted uppercase">Cues</span>
                <span class="text-sm whitespace-pre-wrap text-fg">{{ cues }}</span>
            </span>
        </span>
        <span v-else class="flex items-center gap-3">
            <span class="text-xl font-semibold text-fg tabular-nums">{{ weight }} lb</span>
            <Badge v-if="badge" :tone="badge.tone">{{ badge.text }}</Badge>
            <span class="flex min-w-0 flex-1 flex-wrap justify-end gap-1.5">
                <Badge v-for="s in machineSettings.slice(0, 3)" :key="s.label" variant="outline">
                    {{ s.label }} {{ s.value.trim() || "—" }}
                </Badge>
                <Badge v-if="plates" variant="outline">{{ plates }}</Badge>
            </span>
            <ChevronRight class="size-4 shrink-0 text-fg-subtle" aria-hidden="true" />
        </span>
    </Card>
</template>
