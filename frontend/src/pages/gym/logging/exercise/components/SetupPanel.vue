<script setup lang="ts">
import { computed } from "vue";
import { ChevronRight } from "lucide-vue-next";
import { Badge, Card, SegmentedControl } from "~/shared/ui";
import type { MachineSetting } from "~/types/workout";
import { formatWeightDelta, type OverloadStatus } from "../domain/liveLogging";
import type { WeightProgression } from "../domain/weightProgression";

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
    /** Earned increase still open to change (before the first set); null hides the choice. */
    progression: WeightProgression | null;
}>();

const emit = defineEmits<{ (e: "edit"): void; (e: "choose", choice: "up" | "stay"): void }>();

// Going up is the default (don't miss progressive overload); staying is one tap and can be undone.
// The switch takes the "+5" badge's place on the weight line, so the number and the choice sit together.
const showChoice = computed(() => props.expanded && props.progression != null);
const choice = computed({
    get: () => (props.weight === props.progression?.nextWeight ? "up" : "stay"),
    set: (v: "up" | "stay") => emit("choose", v),
});
const choiceOptions = computed(() => {
    const p = props.progression;
    if (!p) return [];
    return [
        { value: "stay" as const, label: String(p.previousWeight) },
        { value: "up" as const, label: String(p.nextWeight) },
    ];
});

// Only progress gets a badge: going up is news, staying the same isn't.
const badge = computed(() => {
    const o = props.overload;
    if (o.kind !== "up" || showChoice.value) return null;
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
    <!-- The switch can't live inside the card (a button), so it's a sibling laid over the weight line:
         same grid cell, and an overlay that repeats the card's padding, label and weight (invisibly) so
         its row is exactly as tall as the card's and the switch centres on the number. -->
    <div class="grid">
        <Card
            class="col-start-1 row-start-1"
            as="button"
            type="button"
            interactive
            :padding="expanded ? 'md' : 'sm'"
            :aria-label="`Weight ${weight} pounds. Change weight and setup`"
            @click="emit('edit')"
        >
            <span v-if="expanded" class="flex flex-col gap-4">
                <!-- No chevron here: the whole (big) card is the tap target, and the corner belongs to the
                     go-up/stay switch when it shows. -->
                <span class="flex min-w-0 flex-col gap-1">
                    <span class="text-xs font-semibold tracking-wide text-fg-muted uppercase">Exercise Setup</span>
                    <span class="flex items-baseline gap-2">
                        <span class="text-display font-semibold text-fg tabular-nums">{{ weight }} lb</span>
                        <Badge v-if="badge" :tone="badge.tone">{{ badge.text }}</Badge>
                    </span>
                    <span class="text-sm text-fg-muted tabular-nums">{{ lastLine }}</span>
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
                        <span class="text-sm text-fg-subtle">Seat, pad and pulley positions will show here.</span>
                    </span>
                </span>
                <span v-if="cues" class="flex flex-col gap-1">
                    <span class="text-2xs tracking-wide text-fg-muted uppercase">Cues</span>
                    <span class="line-clamp-3 text-sm whitespace-pre-wrap text-fg">{{ cues }}</span>
                </span>
            </span>
            <span v-else class="flex flex-col gap-1">
                <span class="text-xs font-semibold tracking-wide text-fg-muted uppercase">Exercise Setup</span>
                <span class="flex items-center gap-3">
                    <span class="text-xl font-semibold text-fg tabular-nums">{{ weight }} lb</span>
                    <Badge v-if="badge" :tone="badge.tone">{{ badge.text }}</Badge>
                    <!-- One line: badges that don't fit wrap to a second row, which the fixed height hides whole
                         (clipping one mid-word read as "at 4"). The sheet behind the chevron shows them all. -->
                    <span class="flex h-6 min-w-0 flex-1 flex-wrap items-center justify-end gap-x-1.5 gap-y-2 overflow-hidden">
                        <Badge v-for="s in machineSettings.slice(0, 3)" :key="s.label" variant="outline">
                            {{ s.label }} {{ s.value.trim() || "—" }}
                        </Badge>
                        <Badge v-if="plates" variant="outline">{{ plates }}</Badge>
                    </span>
                    <ChevronRight class="size-4 shrink-0 text-fg-subtle" aria-hidden="true" />
                </span>
            </span>
        </Card>
        <div
            v-if="showChoice"
            class="pointer-events-none col-start-1 row-start-1 flex flex-col gap-1 border border-transparent p-4 lg:p-5"
        >
            <span class="invisible text-xs font-semibold uppercase" aria-hidden="true">Exercise Setup</span>
            <span class="flex items-center justify-between gap-3">
                <span class="invisible text-display font-semibold tabular-nums" aria-hidden="true">{{ weight }} lb</span>
                <span class="pointer-events-auto">
                    <SegmentedControl v-model="choice" :options="choiceOptions" size="sm" label="Today's weight">
                        <template #option="{ option }">
                            <span class="flex items-center gap-0.5 tabular-nums">
                                <span v-if="option.value === 'up'" class="text-success" aria-hidden="true">↑</span>
                                {{ option.label }}
                            </span>
                        </template>
                    </SegmentedControl>
                </span>
            </span>
        </div>
    </div>
</template>
