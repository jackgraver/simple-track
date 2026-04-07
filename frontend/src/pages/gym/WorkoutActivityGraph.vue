<script setup lang="ts">
import { computed, ref, watch, nextTick } from "vue";
import { useRouter } from "vue-router";
import { useWorkoutActivity, type WorkoutActivityQueryOpts } from "~/api/workout/queries";
import { formatDateLong } from "~/utils/dateUtil";

const router = useRouter();

type ViewMode = "year" | "weeks52";

const viewMode = ref<ViewMode>("weeks52");
const scrollContainer = ref<HTMLElement | null>(null);

const activityOpts = computed<WorkoutActivityQueryOpts>(() => {
    if (viewMode.value === "year") {
        return { mode: "year", weeks: null, days: null };
    }
    return { mode: "rolling", weeks: 52, days: null };
});

const { data, isPending, isError, error } = useWorkoutActivity(activityOpts);

function parseDateOnly(s: string): Date {
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(s);
    if (m) {
        return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
    }
    const d = new Date(s);
    return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

function dateKey(d: Date): string {
    const y = d.getFullYear();
    const mo = String(d.getMonth() + 1).padStart(2, "0");
    const da = String(d.getDate()).padStart(2, "0");
    return `${y}-${mo}-${da}`;
}

function startOfWeekSunday(d: Date): Date {
    const x = new Date(d.getFullYear(), d.getMonth(), d.getDate());
    const day = x.getDay();
    x.setDate(x.getDate() - day);
    return x;
}

function endOfWeekSaturday(d: Date): Date {
    const start = startOfWeekSunday(d);
    const end = new Date(start);
    end.setDate(end.getDate() + 6);
    return end;
}

const activeSet = computed(() => {
    const s = new Set<string>();
    for (const x of data.value?.active_dates ?? []) {
        s.add(x);
    }
    return s;
});

const todayKey = computed(() => {
    const n = new Date();
    return dateKey(new Date(n.getFullYear(), n.getMonth(), n.getDate()));
});

const gridCells = computed(() => {
    const res = data.value;
    if (!res?.range?.start || !res?.range?.end) return [];
    const rangeStart = parseDateOnly(res.range.start);
    const rangeEnd = parseDateOnly(res.range.end);
    const gridStart = startOfWeekSunday(rangeStart);
    const gridEnd = endOfWeekSaturday(rangeEnd);
    const cells: { key: string; isFuture: boolean; active: boolean }[] = [];
    for (let d = new Date(gridStart); d <= gridEnd; d.setDate(d.getDate() + 1)) {
        const k = dateKey(new Date(d));
        const isFuture = k > todayKey.value;
        const active = activeSet.value.has(k);
        cells.push({ key: k, isFuture, active });
    }
    return cells;
});

const todayWeekIndex = computed(() => {
    const tk = todayKey.value;
    const idx = gridCells.value.findIndex((c) => c.key === tk);
    if (idx < 0) return -1;
    return Math.floor(idx / 7);
});

function scrollToToday() {
    nextTick(() => {
        const el = scrollContainer.value;
        if (!el) return;
        const weekIdx = todayWeekIndex.value;
        if (weekIdx < 0) {
            el.scrollLeft = el.scrollWidth;
            return;
        }
        const cellSize = 10;
        const gap = 3;
        const targetX = weekIdx * (cellSize + gap);
        const maxScroll = el.scrollWidth - el.clientWidth;
        el.scrollLeft = Math.min(targetX, maxScroll);
    });
}

function cellClass(c: { isFuture: boolean; active: boolean }): string {
    if (c.active) {
        return "bg-emerald-500";
    }
    if (c.isFuture) {
        return "bg-secondBg/45";
    }
    return "bg-firstBg";
}

function titleForKey(k: string): string {
    return formatDateLong(`${k}T12:00:00`);
}

/** Matches gym `offset` query: 0 = today, positive = that many days in the past, negative = future. */
function offsetForDateKey(dateKeyStr: string): number {
    const clicked = parseDateOnly(dateKeyStr);
    const today = new Date();
    const t0 = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();
    const t1 = new Date(clicked.getFullYear(), clicked.getMonth(), clicked.getDate()).getTime();
    return Math.round((t0 - t1) / 86400000);
}

function goToGymDay(key: string) {
    const offset = offsetForDateKey(key);
    router.push({
        name: "gym",
        query: offset === 0 ? {} : { offset: String(offset) },
    });
}

watch(gridCells, () => scrollToToday());
</script>

<template>
    <div class="flex flex-col gap-2 w-full min-w-0">
        <div class="flex items-center justify-between">
            <p class="m-0 text-sm text-textSecondary">Activity</p>
            <div class="flex gap-1">
                <button
                    type="button"
                    class="rounded px-2 py-0.5 text-xs transition-colors"
                    :class="
                        viewMode === 'year'
                            ? 'bg-secondBg text-textPrimary'
                            : 'text-textSecondary hover:bg-firstBg hover:text-textPrimary'
                    "
                    @click="viewMode = 'year'"
                >
                    This year
                </button>
                <button
                    type="button"
                    class="rounded px-2 py-0.5 text-xs transition-colors"
                    :class="
                        viewMode === 'weeks52'
                            ? 'bg-secondBg text-textPrimary'
                            : 'text-textSecondary hover:bg-firstBg hover:text-textPrimary'
                    "
                    @click="viewMode = 'weeks52'"
                >
                    52 weeks
                </button>
            </div>
        </div>
        <div v-if="isPending" class="text-sm text-textSecondary">Loading activity…</div>
        <div v-else-if="isError" class="text-sm text-(--color-cf-red)">
            {{ error?.message ?? "Failed to load activity" }}
        </div>
        <div
            v-else
            ref="scrollContainer"
            class="w-full min-w-0 overflow-x-auto scrollbar-thin"
        >
            <div
                class="grid grid-rows-7 grid-flow-col gap-[3px] w-max"
                role="grid"
                aria-label="Workout activity"
            >
                <button
                    v-for="(c, i) in gridCells"
                    :key="c.key + '-' + i"
                    type="button"
                    role="gridcell"
                    class="m-0! min-h-0! size-[10px] rounded-sm p-0! border-0 cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-thirdBg"
                    :class="[
                        cellClass(c),
                        c.key === todayKey
                            ? 'shadow-[inset_0_0_0_2px_#fb923c]!'
                            : 'shadow-none!',
                    ]"
                    :title="titleForKey(c.key)"
                    :aria-label="titleForKey(c.key)"
                    @click="goToGymDay(c.key)"
                />
            </div>
        </div>
    </div>
</template>
