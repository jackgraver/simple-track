<script setup lang="ts">
import { computed } from "vue";
import { ArrowDown, ArrowUp, Trophy } from "lucide-vue-next";
import { Badge, Card } from "~/shared/ui";
import type { LiftGain, SessionSummary } from "../domain/sessionSummary";

// What the session added up to, compared with last time. Gains are success, drops are warn (a lighter
// day isn't an error, so never danger), unchanged/first-time are neutral. Always a word or sign too.
const props = defineProps<{ summary: SessionSummary; complete: boolean }>();

const fmt = new Intl.NumberFormat();
const volume = computed(() =>
    props.summary.volume >= 10_000 ? `${(props.summary.volume / 1000).toFixed(1)}k` : fmt.format(props.summary.volume),
);
const volumeDelta = computed(() => {
    const p = props.summary.volumeDeltaPct;
    if (p == null) return null;
    return { text: `${p > 0 ? "+" : p < 0 ? "−" : "±"}${Math.abs(p)}% vs last time`, up: p > 0, down: p < 0 };
});

type Delta = { text: string; tone: "up" | "down" | "flat" };
const deltaOf = (l: LiftGain): Delta => {
    if (l.deltaWeight == null) return { text: "First time", tone: "flat" };
    if (l.deltaWeight > 0) return { text: `+${l.deltaWeight} lb`, tone: "up" };
    if (l.deltaWeight < 0) return { text: `−${Math.abs(l.deltaWeight)} lb`, tone: "down" };
    if (l.deltaReps && l.deltaReps > 0) return { text: `+${l.deltaReps} ${l.deltaReps === 1 ? "rep" : "reps"}`, tone: "up" };
    if (l.deltaReps && l.deltaReps < 0) return { text: `−${Math.abs(l.deltaReps)} reps`, tone: "down" };
    return { text: "Same as last", tone: "flat" };
};
const toneClass = { up: "text-success", down: "text-warn", flat: "text-fg-muted" } as const;
</script>

<template>
    <section aria-labelledby="session-summary" class="flex flex-col gap-3">
        <h2 id="session-summary" class="m-0 text-xs font-semibold uppercase tracking-wide text-fg-muted">
            {{ complete ? "Session summary" : "Gains so far" }}
        </h2>
        <Card padding="none">
            <dl class="m-0 grid grid-cols-3 divide-x divide-line border-b border-line">
                <div class="flex flex-col gap-0.5 px-3 py-3 lg:px-4">
                    <dt class="text-xs text-fg-muted">Volume</dt>
                    <dd class="m-0 text-lg font-semibold text-fg tabular-nums">
                        {{ volume }} <span class="text-xs font-normal text-fg-muted">lb</span>
                    </dd>
                    <dd
                        v-if="volumeDelta"
                        class="m-0 flex items-center gap-0.5 text-2xs tabular-nums"
                        :class="volumeDelta.up ? 'text-success' : volumeDelta.down ? 'text-warn' : 'text-fg-muted'"
                    >
                        {{ volumeDelta.text }}
                    </dd>
                </div>
                <div class="flex flex-col gap-0.5 px-3 py-3 lg:px-4">
                    <dt class="text-xs text-fg-muted">Sets</dt>
                    <dd class="m-0 text-lg font-semibold text-fg tabular-nums">{{ summary.sets }}</dd>
                    <dd class="m-0 text-2xs text-fg-muted tabular-nums">{{ summary.reps }} reps</dd>
                </div>
                <div class="flex flex-col gap-0.5 px-3 py-3 lg:px-4">
                    <dt class="text-xs text-fg-muted">PRs</dt>
                    <dd
                        class="m-0 flex items-center gap-1 text-lg font-semibold tabular-nums"
                        :class="summary.prCount ? 'text-accent' : 'text-fg'"
                    >
                        <Trophy v-if="summary.prCount" aria-hidden="true" class="size-4" />
                        {{ summary.prCount }}
                    </dd>
                    <dd class="m-0 text-2xs text-fg-muted">{{ summary.prCount ? "new bests" : "none yet" }}</dd>
                </div>
            </dl>
            <ul class="m-0 flex list-none flex-col divide-y divide-line p-0">
                <li v-for="l in summary.lifts" :key="l.key" class="flex min-h-touch items-center gap-3 px-3 py-2 lg:min-h-row lg:px-4">
                    <div class="flex min-w-0 flex-1 flex-col">
                        <span class="truncate text-sm text-fg">{{ l.name }}</span>
                        <span class="text-xs text-fg-muted tabular-nums">
                            Top set {{ l.top.weight > 0 ? `${l.top.weight} lb` : "BW" }} × {{ l.top.reps }}
                        </span>
                    </div>
                    <Badge v-if="l.isPR" tone="accent">
                        <Trophy aria-hidden="true" class="size-3" />
                        PR
                    </Badge>
                    <span
                        class="flex shrink-0 items-center gap-0.5 text-sm font-medium tabular-nums"
                        :class="toneClass[deltaOf(l).tone]"
                    >
                        <ArrowUp v-if="deltaOf(l).tone === 'up'" aria-hidden="true" class="size-3.5" />
                        <ArrowDown v-else-if="deltaOf(l).tone === 'down'" aria-hidden="true" class="size-3.5" />
                        {{ deltaOf(l).text }}
                    </span>
                </li>
            </ul>
        </Card>
    </section>
</template>
