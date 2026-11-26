<script setup lang="ts">
import { computed, ref, watch } from "vue";
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

// Slide the days the way the week moved: later weeks come in from the right. Keyed on the week's
// first day, so picking a day inside the same week doesn't animate.
const weekKey = computed(() => days.value[0]?.key ?? "");
const direction = ref<"next" | "prev">("next");
watch(weekKey, (next, prev) => {
    if (next && prev) direction.value = next > prev ? "next" : "prev";
});

// The arrows land on the new week's first day (Sunday), or on today when that week is this one.
// A week runs Sunday (days[0], the largest offset) to Saturday (offset 6 smaller).
const goToWeek = (step: 1 | -1) => {
    const sunday = (days.value[0]?.offset ?? props.offset) - 7 * step;
    emit("update:offset", sunday >= 0 && sunday - 6 <= 0 ? 0 : sunday);
};

// One outline for the selected day that slides between tiles. Each step is a tile width plus the
// gap-1 between tiles; percentages in translate are of the outline's own (one-tile) width.
const selectedIndex = computed(() => days.value.findIndex((d) => d.isSelected));
const outlineStyle = computed(() => ({ translate: `calc(${selectedIndex.value} * (100% + 0.25rem)) 0` }));

const tileClass = (d: WeekDay) => [
    d.isSelected ? "" : "hover:bg-surface-3",
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
                <Badge
                    v-if="week.streak.value > 1"
                    tone="warn"
                    size="md"
                    :title="`${week.streak.value} sessions in a row. Rest days don't break a streak, missed training days do.`"
                >
                    <Flame aria-hidden="true" class="size-3.5" />
                    <span class="tabular-nums">{{ week.streak.value }}</span>
                    <span>in a row</span>
                </Badge>
                <IconButton label="Previous week" size="sm" @click="goToWeek(-1)">
                    <ChevronLeft />
                </IconButton>
                <IconButton label="Next week" size="sm" @click="goToWeek(1)">
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
            <!--
                A carousel track: both weeks share one grid cell and move together by one week plus the
                tile gap, so the new week arrives right behind the old one, like one continuous row of
                days. Same duration and easing on both sides keep them locked together. The clip (with
                room for focus rings) hides them past the card edge.
            -->
            <div v-else class="relative -mx-1 grid overflow-x-clip px-1">
                <!--
                    The selected-day outline stays out of the sliding weeks: it only moves between
                    columns, so on a week change it glides to the new day while the weeks slide under
                    it, with the same duration so both land together. left-1 and the extra 0.5rem in
                    the width undo this wrapper's px-1. Painted under the tiles, since the week
                    wrappers are positioned and come later.
                -->
                <div
                    v-if="selectedIndex >= 0"
                    aria-hidden="true"
                    class="pointer-events-none absolute inset-y-0 left-1 w-[calc((100%-2rem)/7)] rounded-control border border-accent bg-accent-soft transition-[translate] duration-350 ease-out"
                    :style="outlineStyle"
                />
                <Transition
                    enter-active-class="transition-transform duration-350 ease-out"
                    leave-active-class="transition-transform duration-350 ease-out"
                    :enter-from-class="direction === 'next' ? 'translate-x-[calc(100%+0.25rem)]' : '-translate-x-[calc(100%+0.25rem)]'"
                    :leave-to-class="direction === 'next' ? '-translate-x-[calc(100%+0.25rem)]' : 'translate-x-[calc(100%+0.25rem)]'"
                >
                    <div :key="weekKey" class="relative col-start-1 row-start-1">
                        <ol class="relative m-0 grid list-none grid-cols-7 gap-1 p-0" aria-label="Days this week">
                            <li v-for="d in days" :key="d.key" class="min-w-0">
                                <router-link
                                    :to="to(d)"
                                    :aria-label="dayLabel(d)"
                                    :aria-current="d.isToday ? 'date' : undefined"
                                    class="flex min-h-touch flex-col items-center gap-0.5 rounded-control border border-transparent px-0.5 py-2 no-underline transition-colors duration-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
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
                </Transition>
            </div>
        </div>
    </Card>
</template>
