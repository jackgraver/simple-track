<script setup lang="ts">
import { computed, ref } from "vue";
import { useWorkoutActivity, type WorkoutActivityQueryOpts } from "~/api/workout/queries";
import { formatDateLong } from "~/utils/dateUtil";

type ViewMode = "year" | "weeks52" | "days365";

const viewMode = ref<ViewMode>("weeks52");

const activityOpts = computed<WorkoutActivityQueryOpts>(() => {
    if (viewMode.value === "year") {
        return { mode: "year", weeks: null, days: null };
    }
    if (viewMode.value === "days365") {
        return { mode: "rolling", weeks: null, days: 365 };
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
    const rangeStartKey = dateKey(rangeStart);
    const rangeEndKey = dateKey(rangeEnd);
    const cells: { key: string; inRange: boolean; isFuture: boolean; active: boolean }[] = [];
    for (let d = new Date(gridStart); d <= gridEnd; d.setDate(d.getDate() + 1)) {
        const k = dateKey(new Date(d));
        const inRange = k >= rangeStartKey && k <= rangeEndKey;
        const isFuture = k > todayKey.value;
        const active = activeSet.value.has(k);
        cells.push({ key: k, inRange, isFuture, active });
    }
    return cells;
});

function cellClass(c: {
    inRange: boolean;
    isFuture: boolean;
    active: boolean;
}): string {
    if (!c.inRange) {
        return "bg-zinc-800/40 border border-zinc-700/50";
    }
    if (c.active) {
        return "bg-emerald-600 border border-emerald-500/80";
    }
    if (c.isFuture) {
        return "bg-zinc-800/60 border border-zinc-700/50";
    }
    return "bg-zinc-700/80 border border-zinc-600/60";
}

function titleForKey(k: string): string {
    return formatDateLong(`${k}T12:00:00`);
}
</script>

<template>
    <div class="flex flex-col gap-2 w-full min-w-0">
        <p class="m-0 text-sm text-zinc-400">Activity</p>
        <div class="flex flex-wrap gap-1.5">
            <button
                type="button"
                class="rounded-md px-2 py-1 text-xs border transition-colors"
                :class="
                    viewMode === 'year'
                        ? 'border-emerald-600 bg-emerald-900/40 text-zinc-100'
                        : 'border-zinc-600 bg-zinc-800 text-zinc-300'
                "
                @click="viewMode = 'year'"
            >
                This year
            </button>
            <button
                type="button"
                class="rounded-md px-2 py-1 text-xs border transition-colors"
                :class="
                    viewMode === 'weeks52'
                        ? 'border-emerald-600 bg-emerald-900/40 text-zinc-100'
                        : 'border-zinc-600 bg-zinc-800 text-zinc-300'
                "
                @click="viewMode = 'weeks52'"
            >
                52 weeks
            </button>
            <button
                type="button"
                class="rounded-md px-2 py-1 text-xs border transition-colors"
                :class="
                    viewMode === 'days365'
                        ? 'border-emerald-600 bg-emerald-900/40 text-zinc-100'
                        : 'border-zinc-600 bg-zinc-800 text-zinc-300'
                "
                @click="viewMode = 'days365'"
            >
                365 days
            </button>
        </div>
        <div v-if="isPending" class="text-sm text-zinc-500">Loading activity…</div>
        <div v-else-if="isError" class="text-sm text-red-400">
            {{ error?.message ?? "Failed to load activity" }}
        </div>
        <div v-else class="overflow-x-auto pb-1 -mx-0.5">
            <div
                class="grid w-max gap-1 grid-rows-7 grid-flow-col auto-cols-max p-0.5"
                role="grid"
                :aria-label="'Workout activity'"
            >
                <div
                    v-for="(c, i) in gridCells"
                    :key="c.key + '-' + i"
                    role="gridcell"
                    class="h-2.5 w-2.5 rounded-sm shrink-0"
                    :class="cellClass(c)"
                    :title="titleForKey(c.key)"
                />
            </div>
        </div>
    </div>
</template>
