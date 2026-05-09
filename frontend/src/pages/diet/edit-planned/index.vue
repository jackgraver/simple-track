<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQueries, useQuery, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import { addPlannedMealFromSaved, getDietLogsToday } from "~/api/diet/api";
import type {
    MealItem,
    PlannedMeal,
    SavedMeal,
    SavedMealItem,
} from "~/types/diet";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import { useDeletePlannedMeal } from "~/pages/home/queries/useMealMutations";
import { toast } from "~/composables/toast/useToast";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { homeKeys } from "~/pages/home/queries/keys";
import {
    dateAtDietDayOffset,
    dietDayOffsetFromLocalDate,
    formatDateLong,
    parseDietDayOffsetQuery,
} from "~/utils/dateUtil";

type MacroSummary = {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
};
const ZERO_MACROS: MacroSummary = {
    calories: 0,
    protein: 0,
    fat: 0,
    carbs: 0,
    fiber: 0,
};

function macroTotals(
    items: (MealItem | SavedMealItem)[] | undefined,
): MacroSummary {
    if (!items?.length) return ZERO_MACROS;
    return items.reduce(
        (acc, i) => ({
            calories: acc.calories + (i.food?.calories ?? 0) * i.amount,
            protein: acc.protein + (i.food?.protein ?? 0) * i.amount,
            carbs: acc.carbs + (i.food?.carbs ?? 0) * i.amount,
            fat: acc.fat + (i.food?.fat ?? 0) * i.amount,
            fiber: acc.fiber + (i.food?.fiber ?? 0) * i.amount,
        }),
        { ...ZERO_MACROS },
    );
}

function daysInMonth(year: number, month: number): number {
    return new Date(year, month + 1, 0).getDate();
}

function parseIsoLocal(
    iso: string,
): { y: number; m: number; d: number } | null {
    const [y, mo, day] = iso.split("-").map(Number);
    if (!y || !mo || !day) return null;
    return { y, m: mo - 1, d: day };
}

function enumerateOffsetsInclusive(
    fromIso: string,
    toIso: string,
): number[] | null {
    const a = parseIsoLocal(fromIso);
    const b = parseIsoLocal(toIso);
    if (!a || !b) return null;
    let t1 = new Date(a.y, a.m, a.d);
    let t2 = new Date(b.y, b.m, b.d);
    if (t1.getTime() > t2.getTime()) [t1, t2] = [t2, t1];
    const out: number[] = [];
    const cur = new Date(t1);
    while (cur.getTime() <= t2.getTime()) {
        out.push(
            dietDayOffsetFromLocalDate(
                cur.getFullYear(),
                cur.getMonth(),
                cur.getDate(),
            ),
        );
        cur.setDate(cur.getDate() + 1);
    }
    return out;
}

function offsetToHeadingLabel(offset: number): string {
    const d = dateAtDietDayOffset(offset);
    const ymd = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T12:00:00`;
    return formatDateLong(ymd);
}

const route = useRoute();
const router = useRouter();
const queryClient = useQueryClient();

const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));

const calendarYear = ref(new Date().getFullYear());
const calendarMonth = ref(new Date().getMonth());

watch(
    dateOffset,
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

const monthOffsets = computed(() => {
    const y = calendarYear.value;
    const m = calendarMonth.value;
    const dim = daysInMonth(y, m);
    return Array.from({ length: dim }, (_, i) =>
        dietDayOffsetFromLocalDate(y, m, i + 1),
    );
});

const monthQueries = useQueries({
    queries: computed(() =>
        monthOffsets.value.map((offset) => ({
            queryKey: homeKeys.diet.today(offset),
            queryFn: () => getDietLogsToday(offset),
            staleTime: 1000 * 60 * 2,
        })),
    ),
});

function plannedCountForMonthDay(index: number): number {
    return monthQueries.value[index]?.data?.day.plannedMeals?.length ?? 0;
}

function monthDayPending(index: number): boolean {
    return monthQueries.value[index]?.isPending ?? true;
}

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
    const d = dateAtDietDayOffset(dateOffset.value);
    return (
        d.getFullYear() !== calendarYear.value ||
        d.getMonth() !== calendarMonth.value
    );
});

function setOffset(offset: number) {
    router.replace({ query: { ...route.query, offset: String(offset) } });
}

const {
    data: dayData,
    isPending: dayPending,
    error: dayError,
} = useDietLogsToday(dateOffset);
const plannedMeals = computed(
    (): PlannedMeal[] => dayData.value?.day.plannedMeals ?? [],
);

const {
    data: savedRaw,
    isPending: savedPending,
    error: savedError,
} = useQuery({
    queryKey: ["savedMeals", "all"],
    queryFn: () =>
        apiClient
            .get<{ saved_meals: SavedMeal[] }>("/diet/meals/saved-meal/all")
            .then((r) => r.data.saved_meals),
});

const savedMeals = computed((): SavedMeal[] => savedRaw.value ?? []);

const deletePlanned = useDeletePlannedMeal(dateOffset);
const isRemoving = computed(() => deletePlanned.isPending.value);
const singleAdding = ref(false);

async function removePlanned(pm: PlannedMeal) {
    try {
        await deletePlanned.mutateAsync(pm.ID);
        toast.push("Removed from planned", "success");
    } catch {
        toast.push("Could not remove planned meal", "error");
    }
}

async function addSavedToSelectedDay(sm: SavedMeal) {
    const off = dateOffset.value;
    singleAdding.value = true;
    try {
        await addPlannedMealFromSaved(sm.ID, off);
        await queryClient.invalidateQueries({
            queryKey: homeKeys.diet.today(off),
        });
        toast.push(`Added "${sm.name}" to planned`, "success");
    } catch {
        toast.push("Could not add to planned", "error");
    } finally {
        singleAdding.value = false;
    }
}

const bulkFrom = ref("");
const bulkTo = ref("");
const bulkSavedIds = ref(new Set<number>());
const bulkSaving = ref(false);

function toggleBulkSaved(id: number) {
    const next = new Set(bulkSavedIds.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    bulkSavedIds.value = next;
}

async function saveBulkRange() {
    const from = bulkFrom.value.trim();
    const to = bulkTo.value.trim();
    if (!from || !to) {
        toast.push("Choose both start and end dates", "error");
        return;
    }
    const offsets = enumerateOffsetsInclusive(from, to);
    if (!offsets?.length) {
        toast.push("Invalid date range", "error");
        return;
    }
    const mealIds = [...bulkSavedIds.value];
    if (mealIds.length === 0) {
        toast.push("Select at least one saved meal", "error");
        return;
    }
    bulkSaving.value = true;
    try {
        for (const off of offsets) {
            for (const mealId of mealIds) {
                await addPlannedMealFromSaved(mealId, off);
            }
        }
        for (const off of offsets) {
            await queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(off),
            });
        }
        toast.push(
            `Added ${mealIds.length} meal(s) across ${offsets.length} day(s)`,
            "success",
        );
    } catch {
        toast.push("Could not apply range", "error");
        await queryClient.invalidateQueries({ queryKey: ["home", "diet"] });
    } finally {
        bulkSaving.value = false;
    }
}
</script>

<template>
    <div class="flex flex-col gap-8 pb-8 pt-2">
        <h1 class="m-0 text-xl font-semibold tracking-tight text-zinc-100">
            Edit planned meals
        </h1>
        <section
            class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-900/40 px-3 py-4"
        >
            <div class="flex items-center justify-between gap-2">
                <h2 class="m-0 text-base font-semibold text-zinc-200">
                    Calendar
                </h2>
                <div class="flex items-center gap-1">
                    <button
                        type="button"
                        class="rounded-md border border-zinc-600 p-1.5 text-zinc-300 hover:bg-zinc-800"
                        aria-label="Previous month"
                        @click="prevMonth"
                    >
                        <ChevronLeft class="size-5" />
                    </button>
                    <span
                        class="min-w-48 text-center text-sm font-medium text-zinc-100"
                        >{{ monthLabel }}</span
                    >
                    <button
                        type="button"
                        class="rounded-md border border-zinc-600 p-1.5 text-zinc-300 hover:bg-zinc-800"
                        aria-label="Next month"
                        @click="nextMonth"
                    >
                        <ChevronRight class="size-5" />
                    </button>
                </div>
            </div>
            <div
                class="grid grid-cols-7 gap-1 text-center text-xs font-medium text-zinc-500"
            >
                <div v-for="w in WEEKDAY_LABELS" :key="w">{{ w }}</div>
            </div>
            <div class="grid grid-cols-7 gap-1">
                <template v-for="(cell, idx) in calendarCells" :key="idx">
                    <div v-if="cell.kind === 'pad'" />
                    <button
                        v-else
                        type="button"
                        class="flex min-h-12 flex-col items-center justify-center rounded-lg border py-1 text-sm transition-colors"
                        :class="
                            cell.offset === dateOffset
                                ? 'border-amber-600/80 bg-amber-950/40 text-amber-50'
                                : cell.offset === 0
                                  ? 'border-zinc-500 bg-zinc-800/90 text-zinc-100'
                                  : 'border-zinc-700 bg-zinc-900/60 text-zinc-200 hover:bg-zinc-800'
                        "
                        @click="setOffset(cell.offset)"
                    >
                        <span class="font-semibold leading-none">{{
                            cell.day
                        }}</span>
                        <span
                            v-if="monthDayPending(cell.index)"
                            class="mt-0.5 text-[10px] text-zinc-600"
                            >…</span
                        >
                        <span
                            v-else-if="plannedCountForMonthDay(cell.index) > 0"
                            class="mt-0.5 text-[10px] leading-none text-zinc-500"
                        >
                            {{ plannedCountForMonthDay(cell.index) }} planned
                        </span>
                    </button>
                </template>
            </div>
            <p
                v-if="selectedOutsideVisibleMonth"
                class="m-0 text-xs text-zinc-500"
            >
                Selected day is not in this month.
            </p>
        </section>
        <section class="flex flex-col gap-4">
            <h2 class="m-0 text-sm text-zinc-400">
                Editing:
                <span class="font-medium text-zinc-200">{{
                    offsetToHeadingLabel(dateOffset)
                }}</span>
            </h2>
            <div v-if="dayPending" class="text-sm text-zinc-500">
                Loading day…
            </div>
            <div v-else-if="dayError" class="text-sm text-red-400">
                {{ dayError.message }}
            </div>
            <template v-else>
                <div class="flex flex-col gap-3">
                    <h3 class="m-0 text-sm font-semibold text-zinc-200">
                        Currently planned
                    </h3>
                    <p
                        v-if="plannedMeals.length === 0"
                        class="m-0 text-sm text-zinc-500"
                    >
                        Nothing planned yet. Add a saved meal below.
                    </p>
                    <ul v-else class="m-0 flex list-none flex-col gap-2 p-0">
                        <li
                            v-for="pm in plannedMeals"
                            :key="pm.ID"
                            class="flex items-center justify-between gap-2 rounded-lg border border-zinc-700 bg-zinc-900/50 px-3 py-2"
                        >
                            <div class="min-w-0 flex-1">
                                <p class="m-0 font-medium text-zinc-100">
                                    {{ pm.meal.name }}
                                </p>
                                <SimpleMacros
                                    v-if="pm.meal.items?.length"
                                    class="mt-1"
                                    font-size="0.8rem"
                                    v-bind="macroTotals(pm.meal.items)"
                                />
                            </div>
                            <button
                                type="button"
                                class="shrink-0 rounded-md border border-red-900/60 bg-red-950/40 px-2 py-1 text-sm font-medium text-red-200 hover:bg-red-950/70 disabled:opacity-50"
                                :disabled="isRemoving"
                                @click="removePlanned(pm)"
                            >
                                Remove
                            </button>
                        </li>
                    </ul>
                </div>
                <div class="flex flex-col gap-3">
                    <h3 class="m-0 text-sm font-semibold text-zinc-200">
                        Add from saved meals
                    </h3>
                    <p v-if="savedError" class="m-0 text-sm text-red-400">
                        Could not load saved meals.
                    </p>
                    <p
                        v-else-if="savedPending"
                        class="m-0 text-sm text-zinc-500"
                    >
                        Loading saved meals…
                    </p>
                    <p
                        v-else-if="savedMeals.length === 0"
                        class="m-0 text-sm text-zinc-500"
                    >
                        No saved meals yet. Create one from the Diet page.
                    </p>
                    <div v-else class="flex flex-wrap gap-2">
                        <button
                            v-for="sm in savedMeals"
                            :key="sm.ID"
                            type="button"
                            class="flex max-w-full flex-col items-start gap-1 rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-left hover:bg-zinc-700 disabled:opacity-50"
                            :disabled="singleAdding"
                            @click="addSavedToSelectedDay(sm)"
                        >
                            <span class="font-medium text-zinc-100">{{
                                sm.name
                            }}</span>
                            <SimpleMacros
                                v-if="sm.items?.length"
                                font-size="0.8rem"
                                v-bind="macroTotals(sm.items)"
                            />
                        </button>
                    </div>
                </div>
            </template>
        </section>
        <section
            class="flex flex-col gap-4 rounded-lg border border-zinc-700 bg-zinc-900/40 px-3 py-4"
        >
            <h2 class="m-0 text-base font-semibold text-zinc-200">
                Plan a date range
            </h2>
            <p class="m-0 text-xs text-zinc-500">
                Pick dates and saved meals, then save.
            </p>
            <div class="flex flex-wrap items-end gap-3">
                <label
                    class="flex flex-col gap-1 text-xs font-medium text-zinc-400"
                >
                    Start
                    <input
                        v-model="bulkFrom"
                        type="date"
                        class="rounded-md border border-zinc-600 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                    />
                </label>
                <label
                    class="flex flex-col gap-1 text-xs font-medium text-zinc-400"
                >
                    End
                    <input
                        v-model="bulkTo"
                        type="date"
                        class="rounded-md border border-zinc-600 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                    />
                </label>
            </div>
            <div class="flex flex-col gap-2">
                <span class="text-xs font-medium text-zinc-400"
                    >Saved meals to add on each day</span
                >
                <p
                    v-if="savedPending || savedError || savedMeals.length === 0"
                    class="m-0 text-sm text-zinc-500"
                >
                    <template v-if="savedPending">Loading…</template>
                    <template v-else-if="savedError"
                        >Could not load saved meals.</template
                    >
                    <template v-else>No saved meals.</template>
                </p>
                <div v-else class="flex flex-col gap-2">
                    <label
                        v-for="sm in savedMeals"
                        :key="'bulk-' + sm.ID"
                        class="flex cursor-pointer items-start gap-2 rounded-md border border-zinc-700 bg-zinc-900/60 px-3 py-2 hover:bg-zinc-900"
                    >
                        <input
                            type="checkbox"
                            class="mt-1 accent-amber-600"
                            :checked="bulkSavedIds.has(sm.ID)"
                            @change="toggleBulkSaved(sm.ID)"
                        />
                        <span class="min-w-0 flex-1">
                            <span class="font-medium text-zinc-100">{{
                                sm.name
                            }}</span>
                            <SimpleMacros
                                v-if="sm.items?.length"
                                class="mt-1 block"
                                font-size="0.75rem"
                                v-bind="macroTotals(sm.items)"
                            />
                        </span>
                    </label>
                </div>
            </div>
            <button
                type="button"
                class="self-start rounded-md border border-emerald-800/60 bg-emerald-950/50 px-4 py-2 text-sm font-medium text-emerald-100 hover:bg-emerald-950/80 disabled:opacity-50"
                :disabled="
                    bulkSaving || savedPending || savedMeals.length === 0
                "
                @click="saveBulkRange"
            >
                {{ bulkSaving ? "Saving…" : "Save to range" }}
            </button>
        </section>
    </div>
</template>
