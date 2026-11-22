<script setup lang="ts">
import { computed, ref } from "vue";
import { useWorkoutTimingStats } from "~/api/workout/queries";
import type { DurationStat } from "~/api/workout/api";

const RANGES = [
    { days: 30, label: "30d" },
    { days: 90, label: "90d" },
    { days: 365, label: "1y" },
] as const;

const days = ref<number>(90);
const showAllExercises = ref(false);
const { data, isPending, isError, error } = useWorkoutTimingStats(days);

function formatDuration(seconds: number): string {
    const s = Math.round(seconds);
    if (s < 60) return `${s}s`;
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const rem = s % 60;
    if (h > 0) return `${h}h ${String(m).padStart(2, "0")}m`;
    return rem === 0 ? `${m}m` : `${m}m ${String(rem).padStart(2, "0")}s`;
}

function fmt(stat: DurationStat | undefined): string {
    return stat && stat.count > 0 ? formatDuration(stat.avg_seconds) : "—";
}

const summary = computed(() => {
    const d = data.value;
    return [
        {
            label: "Workout",
            hint: "first set to last set",
            stat: d?.workout,
            unit: "workouts",
        },
        {
            label: "Set to set",
            hint: "includes rest",
            stat: d?.set_interval,
            unit: "intervals",
        },
        {
            label: "Between exercises",
            hint: "last set to next first set",
            stat: d?.transition,
            unit: "changes",
        },
    ];
});

const EXERCISE_PREVIEW = 6;
const sortedExercises = computed(() =>
    [...(data.value?.exercises ?? [])].sort(
        (a, b) =>
            b.transition.count + b.set_interval.count -
            (a.transition.count + a.set_interval.count),
    ),
);
const visibleExercises = computed(() =>
    showAllExercises.value
        ? sortedExercises.value
        : sortedExercises.value.slice(0, EXERCISE_PREVIEW),
);
</script>

<template>
    <section class="flex w-full min-w-0 flex-col gap-3">
        <div class="flex items-center justify-between">
            <p class="m-0 text-sm text-textSecondary">Timing</p>
            <div class="flex gap-1">
                <button
                    v-for="r in RANGES"
                    :key="r.days"
                    type="button"
                    class="rounded px-2 py-0.5 text-xs transition-colors"
                    :class="
                        days === r.days
                            ? 'bg-secondBg text-textPrimary'
                            : 'text-textSecondary hover:bg-firstBg hover:text-textPrimary'
                    "
                    @click="days = r.days"
                >
                    {{ r.label }}
                </button>
            </div>
        </div>

        <div v-if="isPending" class="text-sm text-textSecondary">
            Loading timing…
        </div>
        <div v-else-if="isError" class="text-sm text-(--color-cf-red)">
            {{ error?.message ?? "Failed to load timing" }}
        </div>
        <template v-else-if="data">
            <div class="grid grid-cols-3 gap-2">
                <div
                    v-for="s in summary"
                    :key="s.label"
                    class="flex min-w-0 flex-col gap-0.5 rounded-md bg-secondBg px-3 py-2"
                    :title="
                        s.stat && s.stat.count > 0
                            ? `Median ${formatDuration(s.stat.median_seconds)} · ${s.stat.count} ${s.unit}`
                            : `No ${s.unit} yet`
                    "
                >
                    <span class="truncate text-xs text-textSecondary">{{
                        s.label
                    }}</span>
                    <span
                        class="text-lg font-semibold tabular-nums text-textPrimary"
                        >{{ fmt(s.stat) }}</span
                    >
                    <span class="truncate text-[11px] text-textSecondary">{{
                        s.hint
                    }}</span>
                </div>
            </div>

            <div
                v-if="sortedExercises.length > 0"
                class="flex flex-col gap-1"
            >
                <div
                    class="grid grid-cols-[1fr_auto_auto] gap-x-4 px-1 text-[11px] text-textSecondary"
                >
                    <span>Exercise</span>
                    <span class="w-16 text-right">Set to set</span>
                    <span class="w-16 text-right">Lead-in</span>
                </div>
                <div
                    v-for="ex in visibleExercises"
                    :key="ex.exercise_id"
                    class="grid grid-cols-[1fr_auto_auto] items-baseline gap-x-4 rounded px-1 py-1 text-sm odd:bg-firstBg"
                >
                    <span class="truncate text-textPrimary">{{ ex.name }}</span>
                    <span
                        class="w-16 text-right tabular-nums text-textPrimary"
                        :title="`${ex.set_interval.count} intervals`"
                        >{{ fmt(ex.set_interval) }}</span
                    >
                    <span
                        class="w-16 text-right tabular-nums text-textPrimary"
                        :title="`Time from the previous exercise's last set · ${ex.transition.count} times`"
                        >{{ fmt(ex.transition) }}</span
                    >
                </div>
                <button
                    v-if="sortedExercises.length > EXERCISE_PREVIEW"
                    type="button"
                    class="self-start rounded px-1 py-0.5 text-xs text-textSecondary hover:text-textPrimary"
                    @click="showAllExercises = !showAllExercises"
                >
                    {{
                        showAllExercises
                            ? "Show fewer"
                            : `Show all ${sortedExercises.length}`
                    }}
                </button>
            </div>
            <p
                v-else-if="data.workout.count === 0"
                class="m-0 text-sm text-textSecondary"
            >
                No live-logged workouts in this range yet.
            </p>
        </template>
    </section>
</template>
