<script setup lang="ts">
import { computed } from "vue";
import { Check, ChevronLeft, ChevronRight, Flame } from "lucide-vue-next";
import { Badge, Card, ErrorState, IconButton, Skeleton, useDelayedFlag } from "~/shared/ui";
import type { useGymWeek, WeekDay } from "./useGymWeek";

const props = defineProps<{ week: ReturnType<typeof useGymWeek>; offset: number }>();
const emit = defineEmits<{ "update:offset": [offset: number] }>();

const showSkeleton = useDelayedFlag(() => props.week.isPending.value);
const days = computed(() => props.week.days.value);

const status = (d: WeekDay): "done" | "missed" | "planned" | "rest" => {
    if (d.trained) return "done";
    if (!d.scheduled) return "rest";
    return d.isFuture || d.isToday ? "planned" : "missed";
};

const dayLabel = (d: WeekDay) => {
    const words = { done: "trained", missed: "missed", planned: "planned", rest: "rest day" }[status(d)];
    const plan = d.scheduled?.name ?? (d.trained ? "workout" : "");
    return [d.date.toLocaleDateString(undefined, { weekday: "long", month: "long", day: "numeric" }), plan, words]
        .filter(Boolean)
        .join(", ");
};

const to = (d: WeekDay) => ({ name: "gym", query: d.offset === 0 ? {} : { offset: String(d.offset) } });

const rangeLabel = computed(() => {
    const first = days.value[0]?.date;
    const last = days.value[6]?.date;
    if (!first || !last) return "";
    const fmt = (d: Date) => d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
    return `${fmt(first)} – ${fmt(last)}`;
});

const tileClass = (d: WeekDay) => [
    d.isSelected ? "border-accent bg-accent-soft" : "border-transparent hover:bg-surface-3",
    status(d) === "rest" && !d.isSelected ? "text-fg-subtle" : "text-fg",
];
</script>

<template>
    <Card padding="sm">
        <div class="flex flex-col gap-3">
            <div class="flex items-center gap-2">
                <div class="flex min-w-0 flex-1 flex-col">
                    <h2 class="m-0 text-base font-semibold text-fg">
                        {{ week.isCurrentWeek.value ? "This week" : rangeLabel }}
                    </h2>
                    <p class="m-0 text-xs text-fg-muted tabular-nums">
                        <template v-if="week.activeProgram.value">
                            {{ week.sessionsThisWeek.value }} of {{ week.weeklyTarget.value }} sessions ·
                            {{ week.activeProgram.value.name }}
                        </template>
                        <template v-else>{{ week.sessionsThisWeek.value }} sessions · no active program</template>
                    </p>
                </div>
                <Badge v-if="week.weekStreak.value > 1" tone="warn" size="md">
                    <Flame aria-hidden="true" class="size-3.5" />
                    <span class="tabular-nums">{{ week.weekStreak.value }} wk streak</span>
                </Badge>
                <IconButton label="Previous week" size="sm" @click="emit('update:offset', offset + 7)">
                    <ChevronLeft />
                </IconButton>
                <IconButton label="Next week" size="sm" @click="emit('update:offset', offset - 7)">
                    <ChevronRight />
                </IconButton>
            </div>

            <ErrorState
                v-if="week.isError.value"
                compact
                :error="week.error.value"
                title="Couldn't load your week"
                @retry="week.refetch()"
            />
            <template v-else-if="week.isPending.value">
                <Skeleton v-if="showSkeleton" shape="block" height="4.75rem" />
            </template>
            <ol v-else class="m-0 grid list-none grid-cols-7 gap-1 p-0" aria-label="Days this week">
                <li v-for="d in days" :key="d.key" class="min-w-0">
                    <router-link
                        :to="to(d)"
                        :aria-label="dayLabel(d)"
                        :aria-current="d.isToday ? 'date' : undefined"
                        class="flex min-h-touch flex-col items-center gap-0.5 rounded-control border px-0.5 py-2 no-underline transition-colors duration-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
                        :class="tileClass(d)"
                    >
                        <span
                            class="text-2xs font-medium uppercase tracking-wide"
                            :class="d.isToday ? 'text-accent' : 'text-fg-muted'"
                        >
                            {{ d.isToday ? "Today" : d.weekday }}
                        </span>
                        <span class="text-base font-semibold tabular-nums">{{ d.dayOfMonth }}</span>
                        <span class="w-full truncate text-center text-2xs">
                            {{ d.scheduled?.name ?? "Rest" }}
                        </span>
                        <span class="flex h-4 items-center" aria-hidden="true">
                            <Check v-if="status(d) === 'done'" class="size-3.5 text-success" stroke-width="3" />
                            <span v-else-if="status(d) === 'planned'" class="size-1.5 rounded-full bg-fg-muted" />
                            <span
                                v-else-if="status(d) === 'missed'"
                                class="size-1.5 rounded-full border border-fg-subtle"
                            />
                        </span>
                    </router-link>
                </li>
            </ol>
        </div>
    </Card>
</template>
