<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { useDietDays } from "~/api/diet/queries";
import { weekLandingOffset, withDietDayOffsetQuery } from "~/utils/dateUtil";
import { Card, ErrorState, IconButton, Skeleton, useDelayedFlag, WeekCarousel } from "~/shared/ui";
import { averageLogged, dietWeek, formatInt, isMeaningfullyOver, remaining, type DietWeekDay } from "./dietDay";

const props = defineProps<{ offset: number }>();

const week = computed(() => dietWeek(props.offset));
const queries = useDietDays(() => week.value.map((d) => d.offset));

const isPending = computed(() => queries.value.some((q) => q.isPending));
const failed = computed(() => queries.value.find((q) => q.isError));
// Only the first load shows a skeleton. After that a new week slides in right away and each day's
// numbers fill in as they arrive, so the strip never unmounts mid-slide.
const loadedOnce = ref(false);
watch(isPending, (p) => (loadedOnce.value ||= !p), { immediate: true });
const firstLoad = computed(() => isPending.value && !loadedOnce.value);
const showSkeleton = useDelayedFlag(firstLoad);

type Tile = DietWeekDay & { kcal: number | null; target: number; fill: number; over: boolean };

const tiles = computed<Tile[]>(() =>
    week.value.map((d, i) => {
        const data = queries.value[i]?.data;
        const kcal = data ? data.totalCalories : null;
        const target = data?.day.plan?.calories ?? 0;
        return {
            ...d,
            kcal,
            target,
            fill: kcal == null ? 0 : remaining(kcal, target).fill,
            over: kcal != null && isMeaningfullyOver(kcal, target),
        };
    }),
);

const isCurrentWeek = computed(() => week.value.some((d) => d.isToday));
const fmtShort = (d: Date) => d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
const title = computed(() => {
    if (isCurrentWeek.value) return "This week";
    const first = week.value[0]!.date;
    const last = week.value[6]!.date;
    return `${fmtShort(first)} – ${fmtShort(last)}`;
});

// Future days have nothing eaten yet, so they don't pull the average down.
const average = computed(() => averageLogged(tiles.value.filter((t) => !t.isFuture).map((t) => t.kcal)));
const target = computed(() => tiles.value.find((t) => t.isSelected)?.target ?? tiles.value[0]?.target ?? 0);
const daysLogged = computed(() => tiles.value.filter((t) => !t.isFuture && (t.kcal ?? 0) > 0).length);

const to = (d: DietWeekDay) => ({ name: "diet", query: withDietDayOffsetQuery(d.offset, {}) });
const router = useRouter();
// The arrows (and swipes) land on the new week's Sunday, or today for the current week.
const goToWeek = (step: 1 | -1) =>
    router.push({ name: "diet", query: withDietDayOffsetQuery(weekLandingOffset(week.value[0]!.offset, step), {}) });

const dayLabel = (t: Tile) => {
    const date = t.date.toLocaleDateString(undefined, { weekday: "long", month: "long", day: "numeric" });
    if (t.isFuture || t.kcal == null) return date;
    if (t.kcal === 0) return `${date}, nothing logged`;
    const vs = t.target > 0 ? ` of ${formatInt(t.target)}` : "";
    return `${date}, ${formatInt(t.kcal)}${vs} kcal${t.over ? ", over target" : ""}`;
};

const tileClass = (t: Tile) => [
    // Selection is drawn by WeekCarousel's sliding outline.
    t.isSelected ? "" : "hover:bg-surface-3",
    t.isFuture && !t.isSelected ? "text-fg-subtle" : "text-fg",
];
</script>

<template>
    <Card padding="sm">
        <div class="flex flex-col gap-3">
            <div class="flex items-center gap-2">
                <div class="flex min-w-0 flex-1 flex-col">
                    <h2 class="m-0 text-base font-semibold text-fg">{{ title }}</h2>
                    <p class="m-0 text-xs text-fg-muted tabular-nums">
                        <template v-if="average != null">
                            avg {{ formatInt(average) }}<template v-if="target > 0"> / {{ formatInt(target) }}</template>
                            kcal · {{ daysLogged }} {{ daysLogged === 1 ? "day" : "days" }} logged
                        </template>
                        <template v-else-if="!isPending">Nothing logged this week</template>
                    </p>
                </div>
                <IconButton label="Previous week" size="sm" @click="goToWeek(-1)">
                    <ChevronLeft />
                </IconButton>
                <IconButton label="Next week" size="sm" @click="goToWeek(1)">
                    <ChevronRight />
                </IconButton>
            </div>

            <ErrorState
                v-if="failed"
                compact
                :error="failed.error"
                title="Couldn't load this week"
                @retry="queries.forEach((q) => q.isError && q.refetch())"
            />
            <template v-else-if="firstLoad">
                <Skeleton v-if="showSkeleton" shape="block" height="4.5rem" />
            </template>
            <WeekCarousel v-else :days="tiles" label="Days this week" @prev="goToWeek(-1)" @next="goToWeek(1)">
                <template #day="{ day: t }">
                    <router-link
                        :to="to(t)"
                        :aria-label="dayLabel(t)"
                        :aria-current="t.isSelected ? 'date' : undefined"
                        class="flex min-h-touch flex-col items-center gap-1 rounded-control border border-transparent px-0.5 py-2 no-underline transition-colors duration-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
                        :class="tileClass(t)"
                    >
                        <span
                            class="text-2xs font-medium uppercase tracking-wide"
                            :class="t.isToday ? 'text-accent' : 'text-fg-muted'"
                        >
                            {{ t.isToday ? "Today" : t.date.toLocaleDateString(undefined, { weekday: "short" }) }}
                        </span>
                        <span class="text-base font-semibold tabular-nums">{{ t.date.getDate() }}</span>
                        <span class="h-1 w-3/4 overflow-hidden rounded-full bg-surface-3" aria-hidden="true">
                            <span
                                class="block h-full rounded-full transition-[width] duration-250 ease-out"
                                :class="t.over ? 'bg-warn' : 'bg-macro-calories'"
                                :style="{ width: `${t.fill}%` }"
                            />
                        </span>
                        <span class="text-2xs tabular-nums" :class="t.over ? 'text-warn' : 'text-fg-muted'" aria-hidden="true">
                            {{ t.isFuture || !t.kcal ? "–" : formatInt(t.kcal) }}
                        </span>
                    </router-link>
                </template>
            </WeekCarousel>
        </div>
    </Card>
</template>
