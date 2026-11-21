<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { computed, ref, watch } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { getMonthPlannedSummary } from "~/api/diet/api";
import { dietKeys } from "~/api/diet/keys";
import {
    dateAtDietDayOffset,
    dietDayOffsetFromLocalDate,
} from "~/utils/dateUtil";
import { Card, ErrorState, IconButton, Section } from "~/shared/ui";
import { daysInMonth } from "./planUtils";

const props = defineProps<{ dateOffset: number }>();
const emit = defineEmits<{ select: [offset: number] }>();

const calendarYear = ref(new Date().getFullYear());
const calendarMonth = ref(new Date().getMonth());

watch(
    () => props.dateOffset,
    (off) => {
        const d = dateAtDietDayOffset(off);
        calendarYear.value = d.getFullYear();
        calendarMonth.value = d.getMonth();
    },
    { immediate: true },
);

const monthLabel = computed(() =>
    new Intl.DateTimeFormat(undefined, {
        month: "long",
        year: "numeric",
    }).format(new Date(calendarYear.value, calendarMonth.value, 1)),
);

function prevMonth() {
    if (calendarMonth.value === 0) {
        calendarYear.value -= 1;
        calendarMonth.value = 11;
    } else {
        calendarMonth.value -= 1;
    }
}

function nextMonth() {
    if (calendarMonth.value === 11) {
        calendarYear.value += 1;
        calendarMonth.value = 0;
    } else {
        calendarMonth.value += 1;
    }
}

const dietMonthOffset = computed(() => {
    const now = new Date();
    return (
        (calendarYear.value - now.getFullYear()) * 12 +
        (calendarMonth.value - now.getMonth())
    );
});
const {
    data: monthPlannedCounts,
    isPending: monthPlannedPending,
    isError: monthError,
    error: monthQueryError,
    refetch: refetchMonth,
} = useQuery({
    queryKey: computed(() =>
        dietKeys.days.monthPlannedSummary(dietMonthOffset.value),
    ),
    queryFn: () => getMonthPlannedSummary(dietMonthOffset.value),
    staleTime: 1000 * 60 * 2,
});

type CalendarCell =
    | { kind: "pad" }
    | { kind: "day"; day: number; offset: number; index: number };

const calendarCells = computed((): CalendarCell[] => {
    const y = calendarYear.value;
    const m = calendarMonth.value;
    const firstWd = new Date(y, m, 1).getDay();
    const dim = daysInMonth(y, m);
    const pads: CalendarCell[] = Array.from({ length: firstWd }, () => ({
        kind: "pad",
    }));
    const days: CalendarCell[] = Array.from({ length: dim }, (_, i) => ({
        kind: "day",
        day: i + 1,
        offset: dietDayOffsetFromLocalDate(y, m, i + 1),
        index: i,
    }));
    return [...pads, ...days];
});

const WEEKDAY_LABELS = [
    "Sun",
    "Mon",
    "Tue",
    "Wed",
    "Thu",
    "Fri",
    "Sat",
] as const;

const selectedOutsideVisibleMonth = computed(() => {
    const d = dateAtDietDayOffset(props.dateOffset);
    return (
        d.getFullYear() !== calendarYear.value ||
        d.getMonth() !== calendarMonth.value
    );
});

function plannedCount(index: number): number {
    return monthPlannedCounts.value?.[index] ?? 0;
}
</script>

<template>
    <Section title="Calendar">
        <template #actions>
            <IconButton label="Previous month" @click="prevMonth">
                <ChevronLeft class="size-5" />
            </IconButton>
            <span
                class="min-w-32 text-center text-sm font-medium text-fg tabular-nums"
                aria-live="polite"
                >{{ monthLabel }}</span
            >
            <IconButton label="Next month" @click="nextMonth">
                <ChevronRight class="size-5" />
            </IconButton>
        </template>
        <Card padding="sm">
            <div class="flex flex-col gap-2">
                <ErrorState
                    v-if="monthError"
                    compact
                    title="Couldn't load planned counts"
                    :error="monthQueryError"
                    @retry="refetchMonth()"
                />
                <div
                    class="grid grid-cols-7 gap-1 text-center text-xs font-medium text-fg-muted"
                    aria-hidden="true"
                >
                    <div v-for="w in WEEKDAY_LABELS" :key="w">{{ w }}</div>
                </div>
                <!-- A radio group: one day is selected, arrow keys move between days, and no button
                     CSS is involved. -->
                <fieldset class="m-0 grid min-w-0 grid-cols-7 gap-1 border-0 p-0">
                    <legend class="sr-only">Day to edit, {{ monthLabel }}</legend>
                    <template v-for="(cell, idx) in calendarCells" :key="idx">
                        <div v-if="cell.kind === 'pad'" />
                        <label
                            v-else
                            class="relative flex min-h-12 cursor-pointer flex-col items-center justify-center rounded-control border py-1 text-sm text-fg transition-colors duration-100 has-checked:border-accent has-checked:bg-accent-soft has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-accent"
                            :class="
                                cell.offset === 0
                                    ? 'border-line-strong bg-surface-2 hover:bg-surface-3'
                                    : 'border-line bg-surface hover:bg-surface-3'
                            "
                        >
                            <!-- eslint-disable-next-line vue/no-restricted-html-elements -- sr-only native radio behind a styled day label (radio-card pattern, PRIMITIVES.md) -->
                            <input
                                type="radio"
                                name="plan-day"
                                class="sr-only"
                                :checked="cell.offset === dateOffset"
                                @change="emit('select', cell.offset)"
                            />
                            <span class="leading-none font-semibold">{{
                                cell.day
                            }}</span>
                            <span
                                v-if="
                                    !monthPlannedPending &&
                                    plannedCount(cell.index) > 0
                                "
                                class="mt-0.5 text-2xs leading-none text-fg-muted tabular-nums"
                            >
                                {{ plannedCount(cell.index)
                                }}<span class="sr-only"> planned</span>
                            </span>
                        </label>
                    </template>
                </fieldset>
                <p class="m-0 text-xs text-fg-muted">
                    Numbers show planned meals per day.
                    <template v-if="selectedOutsideVisibleMonth"
                        >The selected day is not in this month.</template
                    >
                </p>
            </div>
        </Card>
    </Section>
</template>
