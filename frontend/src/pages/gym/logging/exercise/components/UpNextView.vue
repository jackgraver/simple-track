<script setup lang="ts">
import { computed } from "vue";
import { Check, ChevronRight, PartyPopper } from "lucide-vue-next";
import { Badge, Button, Card, EmptyState, List, ListRow } from "~/shared/ui";
import { formatWeightDelta, type ProgressVerdict } from "../domain/liveLogging";
import type { UpNextItem } from "../domain/upNext";

// Shown in place after "Finish": what you just did, then pick what's next without going back to the list.
const props = defineProps<{
    finishedName: string;
    finishedSets: string;
    verdict: ProgressVerdict | null;
    items: UpNextItem[];
}>();

const emit = defineEmits<{
    (e: "go", item: UpNextItem): void;
    (e: "list"): void;
}>();

const next = computed(() => props.items[0] ?? null);
const rest = computed(() => props.items.slice(1));

const badge = (item: UpNextItem) => {
    const o = item.overload;
    if (o.kind === "up") return { text: `${o.earned ? "↑ " : ""}${formatWeightDelta(o.delta)}`, tone: o.earned ? "success" : "neutral" } as const;
    return null;
};
const settingsLine = (item: UpNextItem) =>
    item.machineSettings
        .slice(0, 3)
        .map((s) => `${s.label} ${s.value.trim() || "—"}`)
        .join(" · ");
</script>

<template>
    <div class="flex flex-col gap-6">
        <Card padding="sm" aria-label="Finished exercise">
            <div class="flex flex-col gap-1">
                <div class="flex items-baseline justify-between gap-3">
                    <span class="flex min-w-0 items-center gap-2 font-semibold text-fg">
                        <Check class="size-4 shrink-0 text-success" aria-hidden="true" />
                        <span class="truncate">{{ finishedName }}</span>
                    </span>
                    <span class="shrink-0 text-sm text-fg-muted tabular-nums">{{ finishedSets }}</span>
                </div>
                <p v-if="verdict" class="m-0 text-sm" :class="verdict.onTrack ? 'text-success' : 'text-fg-muted'">
                    {{ verdict.hit }}/{{ verdict.total }} hit {{ verdict.rollover }}
                    {{ verdict.onTrack ? "· up next time" : "· same weight next time" }}
                </p>
            </div>
        </Card>

        <EmptyState
            v-if="!next"
            :icon="PartyPopper"
            title="All logged"
            description="Nice session."
        >
            <template #action>
                <Button variant="primary" @click="emit('list')">Back to session</Button>
            </template>
        </EmptyState>

        <template v-else>
            <section class="flex flex-col gap-2" aria-label="Up next">
                <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Up next</p>
                <Card as="button" type="button" interactive class="border-fg" @click="emit('go', next)">
                    <span class="flex flex-col gap-2">
                        <span class="flex items-baseline justify-between gap-3">
                            <span class="min-w-0 truncate text-xl font-semibold text-fg">{{ next.name }}</span>
                            <span v-if="next.weight != null" class="flex shrink-0 items-center gap-2 tabular-nums">
                                <span class="font-semibold text-fg">{{ next.weight }} lb</span>
                                <Badge v-if="badge(next)" :tone="badge(next)!.tone">{{ badge(next)!.text }}</Badge>
                            </span>
                        </span>
                        <span v-if="next.detail" class="text-sm text-fg-muted tabular-nums">{{ next.detail }}</span>
                        <span v-if="next.machineSettings.length" class="flex flex-wrap gap-1.5">
                            <Badge v-for="s in next.machineSettings" :key="s.label" variant="outline">
                                {{ s.label }} {{ s.value.trim() || "—" }}
                            </Badge>
                        </span>
                    </span>
                </Card>
            </section>

            <section v-if="rest.length" class="flex flex-col gap-2" aria-label="Also left">
                <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Also left · {{ rest.length }}</p>
                <List class="rounded-card border border-line bg-surface">
                    <ListRow v-for="item in rest" :key="item.key" @click="emit('go', item)">
                        {{ item.name }}
                        <template #subtitle>
                            <span class="tabular-nums">
                                {{
                                    [item.weight != null ? `${item.weight} lb` : null, settingsLine(item) || item.detail]
                                        .filter(Boolean)
                                        .join(" · ")
                                }}
                            </span>
                        </template>
                        <template #trailing><ChevronRight class="size-4 text-fg-subtle" aria-hidden="true" /></template>
                    </ListRow>
                </List>
            </section>

            <Button variant="ghost" block @click="emit('list')">
                Session list <ChevronRight class="size-4" aria-hidden="true" />
            </Button>
        </template>
    </div>
</template>
