<script setup lang="ts">
import { AlertTriangle, ArrowDown, ArrowUp, TrendingUp } from "lucide-vue-next";
import { Badge, Card } from "~/shared/ui";
import { mockInsightCallouts, mockInsightStats, type InsightTone } from "./mockInsights";

// Design preview on mock data (see mockInsights.ts). The "Preview" badge stays until it's real.
const toneClass: Record<InsightTone, string> = { up: "text-success", down: "text-warn", flat: "text-fg-muted" };
</script>

<template>
    <section aria-labelledby="gym-insights" class="flex flex-col gap-3">
        <div class="flex items-center gap-2">
            <h2 id="gym-insights" class="m-0 text-base font-semibold text-fg">Insights</h2>
            <Badge>Preview</Badge>
        </div>
        <Card padding="none" class="overflow-hidden">
            <!-- gap-px over a line-coloured background draws the grid dividers at any column count -->
            <dl class="m-0 grid grid-cols-2 gap-px bg-line lg:grid-cols-4">
                <div v-for="s in mockInsightStats" :key="s.key" class="flex flex-col gap-1 bg-surface p-3 lg:p-4">
                    <dt class="text-xs text-fg-muted">{{ s.label }}</dt>
                    <dd class="m-0 text-xl font-semibold tracking-tight text-fg tabular-nums">
                        {{ s.value }}<span v-if="s.unit" class="ml-0.5 text-xs font-normal text-fg-muted">{{ s.unit }}</span>
                    </dd>
                    <dd class="m-0 flex items-center gap-0.5 truncate text-xs tabular-nums" :class="toneClass[s.tone]">
                        <ArrowUp v-if="s.tone === 'up'" aria-hidden="true" class="size-3 shrink-0" />
                        <ArrowDown v-else-if="s.tone === 'down'" aria-hidden="true" class="size-3 shrink-0" />
                        {{ s.note }}
                    </dd>
                </div>
            </dl>
            <ul class="m-0 flex list-none flex-col divide-y divide-line border-t border-line p-0">
                <li v-for="c in mockInsightCallouts" :key="c.key" class="flex items-start gap-3 px-3 py-3 lg:px-4">
                    <span
                        class="flex size-8 shrink-0 items-center justify-center rounded-full"
                        :class="c.kind === 'attention' ? 'bg-warn-soft text-warn' : 'bg-success-soft text-success'"
                    >
                        <AlertTriangle v-if="c.kind === 'attention'" aria-hidden="true" class="size-4" />
                        <TrendingUp v-else aria-hidden="true" class="size-4" />
                    </span>
                    <div class="flex min-w-0 flex-col">
                        <p class="m-0 text-sm font-medium text-fg">{{ c.title }}</p>
                        <p class="m-0 text-sm text-fg-muted">{{ c.body }}</p>
                    </div>
                </li>
            </ul>
        </Card>
    </section>
</template>
