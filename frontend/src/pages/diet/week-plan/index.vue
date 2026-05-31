<script setup lang="ts">
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQueryClient } from "@tanstack/vue-query";
import {
    addPlannedMealFromLabel,
    addPlannedMealFromSaved,
} from "~/api/diet/api";
import type {
    MealItem,
    PlannedMeal,
    SavedMeal,
    SavedMealItem,
} from "~/types/diet";
import { useSavedMeals } from "~/pages/diet/queries/useSavedMeals";
import { useDietWeekLogs } from "~/pages/diet/week-plan/queries/useDietWeekLogs";
import { homeKeys } from "~/pages/home/queries/keys";
import { toast } from "~/composables/toast/useToast";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import {
    dietWeekDays,
    formatDietWeekRangeLabel,
    parseWeekOffsetQuery,
} from "~/utils/dateUtil";

type MacroSummary = {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
};
type DraftEntry =
    | { kind: "saved"; savedMealId: number }
    | { kind: "label"; label: string };

const ZERO_MACROS: MacroSummary = {
    calories: 0,
    protein: 0,
    fat: 0,
    carbs: 0,
    fiber: 0,
};
const WEEKDAY_SHORT = [
    "Sun",
    "Mon",
    "Tue",
    "Wed",
    "Thu",
    "Fri",
    "Sat",
] as const;
const EATING_OUT_LABEL = "Eating out";

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

function formatDayDate(date: Date): string {
    return new Intl.DateTimeFormat(undefined, {
        month: "short",
        day: "numeric",
    }).format(date);
}

function sortedPlannedMeals(list: PlannedMeal[] | undefined): PlannedMeal[] {
    return [...(list ?? [])].sort(
        (a, b) =>
            (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID,
    );
}

function draftLabel(
    entry: DraftEntry,
    savedById: Map<number, SavedMeal>,
): string {
    if (entry.kind === "label") return entry.label;
    return savedById.get(entry.savedMealId)?.name ?? "Saved meal";
}

function isLabelEntry(entry: DraftEntry): boolean {
    return entry.kind === "label";
}

const route = useRoute();
const router = useRouter();
const queryClient = useQueryClient();

const weekOffset = computed(() => parseWeekOffsetQuery(route.query.week));
const weekLabel = computed(() => formatDietWeekRangeLabel(weekOffset.value));
const weekDays = computed(() => dietWeekDays(weekOffset.value));

const { days, queries: weekQueries } = useDietWeekLogs(weekOffset);
const {
    data: savedMealsRaw,
    isPending: savedPending,
    error: savedError,
} = useSavedMeals();
const savedMeals = computed((): SavedMeal[] => savedMealsRaw.value ?? []);
const savedMealsById = computed(
    () => new Map(savedMeals.value.map((s) => [s.ID, s])),
);

const draftByOffset = ref<Record<number, DraftEntry>>({});
const saving = ref(false);
const selectedDayOffset = ref<number | null>(null);

function defaultSelectedOffset(): number {
    const todayInWeek = weekDays.value.find((d) => d.offset === 0);
    return todayInWeek?.offset ?? weekDays.value[0]?.offset ?? 0;
}

watch(
    weekDays,
    (daysList) => {
        if (daysList.length === 0) return;
        const offsets = new Set(daysList.map((d) => d.offset));
        if (
            selectedDayOffset.value == null ||
            !offsets.has(selectedDayOffset.value)
        ) {
            selectedDayOffset.value = defaultSelectedOffset();
        }
    },
    { immediate: true },
);

watch(weekOffset, () => {
    draftByOffset.value = {};
    selectedDayOffset.value = defaultSelectedOffset();
});

const activeDay = computed(
    () =>
        weekDays.value.find((d) => d.offset === selectedDayOffset.value) ??
        weekDays.value[0],
);

const totalDraftCount = computed(() => Object.keys(draftByOffset.value).length);

const weekLoading = computed(() => weekQueries.value.some((q) => q.isPending));
const weekError = computed(
    () => weekQueries.value.find((q) => q.error)?.error ?? null,
);

function dayDataForOffset(offset: number) {
    const idx = days.value.findIndex((d) => d.offset === offset);
    if (idx < 0) return null;
    return weekQueries.value[idx]?.data ?? null;
}

function draftForOffset(offset: number): DraftEntry | undefined {
    return draftByOffset.value[offset];
}

function hasDraftForOffset(offset: number): boolean {
    return draftForOffset(offset) != null;
}

function dinnerDisplayForOffset(offset: number): {
    label: string;
    isDraft: boolean;
    isPlaceholder: boolean;
} | null {
    const draft = draftForOffset(offset);
    if (draft) {
        return {
            label: draftLabel(draft, savedMealsById.value),
            isDraft: true,
            isPlaceholder: isLabelEntry(draft),
        };
    }
    const planned = sortedPlannedMeals(
        dayDataForOffset(offset)?.day.plannedMeals,
    );
    if (planned.length === 0) return null;
    const meal = planned[0].meal;
    return {
        label: meal.name,
        isDraft: false,
        isPlaceholder: !meal.items?.length,
    };
}

function isDraftSelected(offset: number, entry: DraftEntry): boolean {
    const draft = draftForOffset(offset);
    if (!draft) return false;
    if (draft.kind !== entry.kind) return false;
    if (entry.kind === "saved" && draft.kind === "saved") {
        return draft.savedMealId === entry.savedMealId;
    }
    if (entry.kind === "label" && draft.kind === "label") {
        return draft.label === entry.label;
    }
    return false;
}

function selectDay(offset: number) {
    selectedDayOffset.value = offset;
}

function setWeekOffset(next: number) {
    router.replace({
        query: {
            ...route.query,
            week: String(next),
        },
    });
}

function setDraft(offset: number, entry: DraftEntry) {
    draftByOffset.value = { ...draftByOffset.value, [offset]: entry };
}

function clearDraftForDay(offset: number) {
    const next = { ...draftByOffset.value };
    delete next[offset];
    draftByOffset.value = next;
}

function clearDrafts() {
    draftByOffset.value = {};
}

async function saveWeekPlan() {
    const entries = Object.entries(draftByOffset.value).map(
        ([offsetStr, draft]) => ({
            offset: Number(offsetStr),
            draft,
        }),
    );
    if (entries.length === 0) {
        toast.push("Choose at least one dinner for the week", "error");
        return;
    }
    saving.value = true;
    try {
        for (const { offset, draft } of entries) {
            if (draft.kind === "saved") {
                await addPlannedMealFromSaved(draft.savedMealId, offset);
            } else {
                await addPlannedMealFromLabel(draft.label, offset);
            }
        }
        for (const day of days.value) {
            await queryClient.invalidateQueries({
                queryKey: homeKeys.diet.today(day.offset),
            });
        }
        await queryClient.invalidateQueries({
            queryKey: homeKeys.diet.monthPlannedSummaryPrefix,
        });
        clearDrafts();
        toast.push(`Added ${entries.length} dinner(s) to the week`, "success");
    } catch {
        toast.push("Could not save week plan", "error");
        await queryClient.invalidateQueries({ queryKey: ["home", "diet"] });
    } finally {
        saving.value = false;
    }
}
</script>
<template>
    <div class="flex flex-col gap-6 pb-8 pt-2">
        <div class="flex flex-wrap items-center justify-between gap-3">
            <h1 class="m-0 text-xl font-semibold tracking-tight text-zinc-100">
                Plan dinners
            </h1>
            <div class="flex items-center gap-1">
                <button
                    type="button"
                    class="rounded-md border border-zinc-600 p-1.5 text-zinc-300 hover:bg-zinc-800"
                    aria-label="Previous week"
                    :disabled="saving"
                    @click="setWeekOffset(weekOffset - 1)"
                >
                    <ChevronLeft class="size-5" />
                </button>
                <span
                    class="min-w-44 text-center text-sm font-medium text-zinc-100"
                    >{{ weekLabel }}</span
                >
                <button
                    type="button"
                    class="rounded-md border border-zinc-600 p-1.5 text-zinc-300 hover:bg-zinc-800"
                    aria-label="Next week"
                    :disabled="saving"
                    @click="setWeekOffset(weekOffset + 1)"
                >
                    <ChevronRight class="size-5" />
                </button>
            </div>
        </div>
        <div v-if="weekLoading" class="text-sm text-zinc-500">
            Loading week…
        </div>
        <div v-else-if="weekError" class="text-sm text-red-400">
            {{ weekError.message }}
        </div>
        <template v-else>
            <div
                class="grid grid-cols-7 gap-1 rounded-lg border border-zinc-700 bg-zinc-900/40 p-2"
            >
                <button
                    v-for="day in weekDays"
                    :key="day.offset"
                    type="button"
                    class="flex min-w-0 flex-col items-center gap-1 rounded-md border px-1 py-2 text-center transition-colors"
                    :class="
                        day.offset === selectedDayOffset
                            ? 'border-amber-600/80 bg-amber-950/40 text-amber-50'
                            : day.offset === 0
                              ? 'border-zinc-500 bg-zinc-800/90 text-zinc-100 hover:bg-zinc-800'
                              : 'border-zinc-700 bg-zinc-900/60 text-zinc-200 hover:bg-zinc-800'
                    "
                    :disabled="saving"
                    @click="selectDay(day.offset)"
                >
                    <span
                        class="text-[11px] font-semibold uppercase tracking-wide"
                        >{{ WEEKDAY_SHORT[day.dayIndex] }}</span
                    >
                    <span class="text-xs">{{ formatDayDate(day.date) }}</span>
                    <span
                        v-if="dinnerDisplayForOffset(day.offset)"
                        class="line-clamp-2 w-full px-0.5 text-[10px] leading-tight"
                        :class="
                            dinnerDisplayForOffset(day.offset)?.isDraft
                                ? 'text-amber-100/90'
                                : 'text-zinc-400'
                        "
                    >
                        <span
                            v-if="
                                dinnerDisplayForOffset(day.offset)
                                    ?.isPlaceholder
                            "
                            class="italic"
                            >{{
                                dinnerDisplayForOffset(day.offset)?.label
                            }}</span
                        >
                        <span v-else>{{
                            dinnerDisplayForOffset(day.offset)?.label
                        }}</span>
                    </span>
                    <span v-else class="text-[10px] text-zinc-600">—</span>
                </button>
            </div>
            <section
                v-if="activeDay"
                class="flex flex-col gap-5 rounded-lg border border-zinc-700 bg-zinc-900/40 px-4 py-4"
            >
                <div>
                    <h2 class="m-0 text-base font-semibold text-zinc-100">
                        {{ WEEKDAY_SHORT[activeDay.dayIndex] }},
                        {{ formatDayDate(activeDay.date) }}
                        <span
                            v-if="activeDay.offset === 0"
                            class="text-zinc-500"
                            >· Today</span
                        >
                    </h2>
                </div>
                <div
                    v-if="
                        sortedPlannedMeals(
                            dayDataForOffset(activeDay.offset)?.day
                                .plannedMeals,
                        ).length > 0
                    "
                    class="flex flex-col gap-2"
                >
                    <h3
                        class="m-0 text-[11px] font-semibold uppercase tracking-wide text-zinc-500"
                    >
                        Already planned
                    </h3>
                    <ul class="m-0 flex list-none flex-col gap-2 p-0">
                        <li
                            v-for="pm in sortedPlannedMeals(
                                dayDataForOffset(activeDay.offset)?.day
                                    .plannedMeals,
                            )"
                            :key="pm.ID"
                            class="rounded-lg border border-zinc-800 bg-zinc-950/50 px-3 py-2"
                        >
                            <p class="m-0 text-sm font-medium text-zinc-200">
                                {{ pm.meal.name }}
                            </p>
                            <SimpleMacros
                                v-if="pm.meal.items?.length"
                                class="mt-1"
                                font-size="0.75rem"
                                v-bind="macroTotals(pm.meal.items)"
                            />
                            <p
                                v-else
                                class="m-0 mt-1 text-xs italic text-zinc-500"
                            >
                                Placeholder
                            </p>
                        </li>
                    </ul>
                </div>
                <div class="flex flex-col gap-3">
                    <div class="flex items-center justify-between gap-2">
                        <h3 class="m-0 text-sm font-semibold text-zinc-200">
                            Choose dinner
                        </h3>
                        <button
                            v-if="hasDraftForOffset(activeDay.offset)"
                            type="button"
                            class="text-xs text-zinc-400 hover:text-zinc-200 disabled:opacity-50"
                            :disabled="saving"
                            @click="clearDraftForDay(activeDay.offset)"
                        >
                            Clear selection
                        </button>
                    </div>
                    <button
                        type="button"
                        class="flex flex-col items-start rounded-md border px-3 py-2 text-left transition-colors disabled:opacity-50"
                        :class="
                            isDraftSelected(activeDay.offset, {
                                kind: 'label',
                                label: EATING_OUT_LABEL,
                            })
                                ? 'border-amber-600/80 bg-amber-950/40'
                                : 'border-zinc-600 bg-zinc-800/80 hover:bg-zinc-700'
                        "
                        :disabled="saving"
                        @click="
                            setDraft(activeDay.offset, {
                                kind: 'label',
                                label: EATING_OUT_LABEL,
                            })
                        "
                    >
                        <span
                            class="text-sm font-medium italic text-zinc-100"
                            >{{ EATING_OUT_LABEL }}</span
                        >
                        <span class="text-xs text-zinc-500"
                            >Placeholder — log details later</span
                        >
                    </button>
                    <div v-if="savedError" class="text-sm text-red-400">
                        Could not load saved meals.
                    </div>
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
                        No saved meals yet. Use Eating out above or create one
                        from the Diet page.
                    </p>
                    <div v-else class="flex flex-wrap gap-2">
                        <button
                            v-for="sm in savedMeals"
                            :key="sm.ID"
                            type="button"
                            class="flex max-w-full flex-col items-start gap-1 rounded-md border px-3 py-2 text-left transition-colors disabled:opacity-50"
                            :class="
                                isDraftSelected(activeDay.offset, {
                                    kind: 'saved',
                                    savedMealId: sm.ID,
                                })
                                    ? 'border-amber-600/80 bg-amber-950/40'
                                    : 'border-zinc-600 bg-zinc-800 hover:bg-zinc-700'
                            "
                            :disabled="saving"
                            @click="
                                setDraft(activeDay.offset, {
                                    kind: 'saved',
                                    savedMealId: sm.ID,
                                })
                            "
                        >
                            <span class="font-medium text-zinc-100">{{
                                sm.name
                            }}</span>
                            <SimpleMacros
                                v-if="sm.items?.length"
                                font-size="0.75rem"
                                v-bind="macroTotals(sm.items)"
                            />
                        </button>
                    </div>
                </div>
            </section>
            <div
                class="sticky bottom-0 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-zinc-700 bg-zinc-950/95 px-4 py-3 backdrop-blur"
            >
                <p class="m-0 text-sm text-zinc-400">
                    <span v-if="totalDraftCount === 0"
                        >No dinners selected.</span
                    >
                    <span v-else
                        >{{ totalDraftCount }} dinner(s) ready to save.</span
                    >
                </p>
                <div class="flex flex-wrap gap-2">
                    <button
                        type="button"
                        class="rounded-md border border-zinc-600 px-4 py-2 text-sm font-medium text-zinc-200 hover:bg-zinc-800 disabled:opacity-50"
                        :disabled="totalDraftCount === 0 || saving"
                        @click="clearDrafts"
                    >
                        Clear all
                    </button>
                    <button
                        type="button"
                        class="rounded-md border border-emerald-800/60 bg-emerald-950/50 px-4 py-2 text-sm font-medium text-emerald-100 hover:bg-emerald-950/80 disabled:opacity-50"
                        :disabled="totalDraftCount === 0 || saving"
                        @click="saveWeekPlan"
                    >
                        {{ saving ? "Saving…" : "Save week plan" }}
                    </button>
                </div>
            </div>
        </template>
    </div>
</template>
