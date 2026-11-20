import { computed, toValue, type MaybeRefOrGetter } from "vue";
import { useWorkoutActivity, useWorkoutPrograms } from "~/api/workout/queries";
import type { WorkoutPlan } from "~/types/workout";

export type WeekDay = {
    key: string; // YYYY-MM-DD
    date: Date;
    /** Gym `offset` query for this day: 0 = today, positive = past. */
    offset: number;
    weekday: string; // "Mon"
    dayOfMonth: number;
    scheduled: WorkoutPlan | null;
    trained: boolean;
    isToday: boolean;
    isFuture: boolean;
    isSelected: boolean;
};

const DAY_MS = 86_400_000;
const WEEKDAY = new Intl.DateTimeFormat(undefined, { weekday: "short" });
const WEEKDAY_LONG = new Intl.DateTimeFormat(undefined, { weekday: "long" });

export function dateKey(d: Date): string {
    const mo = String(d.getMonth() + 1).padStart(2, "0");
    const da = String(d.getDate()).padStart(2, "0");
    return `${d.getFullYear()}-${mo}-${da}`;
}

function startOfDay(d: Date): Date {
    return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

function addDays(d: Date, n: number): Date {
    return new Date(d.getFullYear(), d.getMonth(), d.getDate() + n);
}

/** Training weeks start on Monday. */
function startOfWeek(d: Date): Date {
    return addDays(d, -((d.getDay() + 6) % 7));
}

function offsetOf(d: Date, today: Date): number {
    return Math.round((today.getTime() - d.getTime()) / DAY_MS);
}

export function weekdayLong(d: Date): string {
    return WEEKDAY_LONG.format(d);
}

/**
 * The week around the selected gym day: what the active program schedules on each day and
 * whether a session was logged. Also derives "sessions this week" and a weekly streak from the
 * same activity query the heatmap uses (shared cache, no extra request).
 */
export function useGymWeek(offset: MaybeRefOrGetter<number>) {
    const programs = useWorkoutPrograms();
    // Same options as WorkoutActivityGraph's default view, so both read one cached response.
    const activity = useWorkoutActivity({ mode: "rolling", weeks: 52, days: null });

    const today = computed(() => startOfDay(new Date()));
    const selected = computed(() => addDays(today.value, -toValue(offset)));

    const activeProgram = computed(() => {
        const list = programs.data.value?.programs ?? [];
        return list.find((p) => p.is_active && p.plans?.length) ?? null;
    });

    /** Plan scheduled per weekday (0 = Sunday), from the active program. */
    const schedule = computed(() => {
        const map = new Map<number, WorkoutPlan>();
        for (const plan of activeProgram.value?.plans ?? []) {
            const days = plan.assigned_days?.length
                ? plan.assigned_days
                : plan.day_of_week != null
                  ? [plan.day_of_week]
                  : [];
            for (const d of days) if (!map.has(d)) map.set(d, plan);
        }
        return map;
    });

    const trainedSet = computed(() => new Set(activity.data.value?.active_dates ?? []));

    const days = computed<WeekDay[]>(() => {
        const start = startOfWeek(selected.value);
        const selKey = dateKey(selected.value);
        const todayKey = dateKey(today.value);
        return Array.from({ length: 7 }, (_, i) => {
            const date = addDays(start, i);
            const key = dateKey(date);
            return {
                key,
                date,
                offset: offsetOf(date, today.value),
                weekday: WEEKDAY.format(date),
                dayOfMonth: date.getDate(),
                scheduled: schedule.value.get(date.getDay()) ?? null,
                trained: trainedSet.value.has(key),
                isToday: key === todayKey,
                isFuture: key > todayKey,
                isSelected: key === selKey,
            };
        });
    });

    const isCurrentWeek = computed(() => days.value.some((d) => d.isToday));
    /** Sessions the program asks for per week (at least 1, so a streak means something without a program). */
    const weeklyTarget = computed(() => Math.max(1, schedule.value.size));
    const sessionsThisWeek = computed(() => days.value.filter((d) => d.trained).length);

    /** Consecutive weeks (ending at the current one) that hit the target. The current week counts once it's hit. */
    const weekStreak = computed(() => {
        const set = trainedSet.value;
        let streak = 0;
        let weekStart = startOfWeek(today.value);
        for (let w = 0; w < 52; w++) {
            let count = 0;
            for (let i = 0; i < 7; i++) if (set.has(dateKey(addDays(weekStart, i)))) count++;
            const hit = count >= weeklyTarget.value;
            if (hit) streak++;
            else if (w > 0) break; // the current week may still be in progress
            weekStart = addDays(weekStart, -7);
        }
        return streak;
    });

    /** The next scheduled training day after the selected one, within two weeks. */
    const nextScheduled = computed(() => {
        for (let i = 1; i <= 14; i++) {
            const date = addDays(selected.value, i);
            const plan = schedule.value.get(date.getDay());
            if (plan) return { date, plan, inDays: i };
        }
        return null;
    });

    return {
        days,
        selected,
        activeProgram,
        isCurrentWeek,
        weeklyTarget,
        sessionsThisWeek,
        weekStreak,
        nextScheduled,
        scheduledFor: (d: Date) => schedule.value.get(d.getDay()) ?? null,
        isPending: computed(() => programs.isPending.value || activity.isPending.value),
        isError: computed(() => programs.isError.value || activity.isError.value),
        error: computed(() => programs.error.value ?? activity.error.value),
        refetch: () => Promise.all([programs.refetch(), activity.refetch()]),
    };
}
