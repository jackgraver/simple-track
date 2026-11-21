<script setup lang="ts">
import { ChevronLeft, ChevronRight, X } from "lucide-vue-next";
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
    useAddPlannedFromSavedAt,
    useCopyPlannedDay,
    useDeletePlannedMealAt,
    useDietDays,
    useReorderPlannedMealsAt,
} from "~/api/diet/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { dietWeek, formatInt } from "~/pages/diet/home/dietDay";
import { CREATE_TYPE } from "~/pages/diet/logmeal/logmealMode";
import type { SavedMeal } from "~/types/diet";
import { parseDietDayOffsetQuery, withDietDayOffsetQuery } from "~/utils/dateUtil";
import {
    Button,
    ErrorState,
    IconButton,
    PageHeader,
    Skeleton,
    useDelayedFlag,
    useIsDesktop,
} from "~/shared/ui";
import AddMealOverlay from "./AddMealOverlay.vue";
import CopyDayOverlay from "./CopyDayOverlay.vue";
import PlanDayCard from "./PlanDayCard.vue";
import {
    buildPlanDay,
    formatWeekSummary,
    moveId,
    savedMealMacros,
    shortDayLabel,
    weekRangeLabel,
    weekSummary,
    type PlanDay,
} from "./domain/planWeek";
import { useSavedMeals } from "./useSavedMeals";

const route = useRoute();
const router = useRouter();

// `?offset=` is any day in the week to show (0 = today, positive = past), as on /diet.
const offset = computed(() => parseDietDayOffsetQuery(route.query.offset));
const week = computed(() => dietWeek(offset.value));
const queries = useDietDays(() => week.value.map((d) => d.offset));

const loading = computed(() => queries.value.some((q) => q.isPending));
const failed = computed(() => queries.value.find((q) => q.isError));
const showSkeleton = useDelayedFlag(loading);

const days = computed<PlanDay[]>(() => week.value.map((d, i) => buildPlanDay(d, queries.value[i]?.data)));
const summary = computed(() => formatWeekSummary(weekSummary(days.value)));
const isCurrentWeek = computed(() => week.value.some((d) => d.isToday));
const rangeLabel = computed(() => weekRangeLabel(week.value[0]!.date, week.value[6]!.date));

function shiftWeek(delta: number) {
    router.replace({ query: { ...route.query, offset: String(offset.value + delta) } });
}
function thisWeek() {
    router.replace({ query: { ...route.query, offset: "0" } });
}

// ---- Saved meals: the desktop tray and the add overlay ----------------------------------------
const { savedMeals, isPending: savedPending, error: savedError, refetch: refetchSaved } = useSavedMeals();
const createSavedTo = computed(() => ({ name: "diet-log", query: withDietDayOffsetQuery(0, { type: CREATE_TYPE }) }));

// Desktop: pick a meal in the tray, then each day's add button adds it. Esc or ✕ drops the pick.
const isDesktop = useIsDesktop();
const armed = ref<SavedMeal | null>(null);
watch(isDesktop, (d) => {
    if (!d) armed.value = null;
});
watch(savedMeals, (list) => {
    if (armed.value && !list.some((s) => s.ID === armed.value?.ID)) armed.value = null;
});
function toggleArmed(sm: SavedMeal) {
    armed.value = armed.value?.ID === sm.ID ? null : sm;
}
function onKey(e: KeyboardEvent) {
    if (e.key === "Escape" && armed.value && !addOpen.value && !copyOpen.value) armed.value = null;
}
onMounted(() => window.addEventListener("keydown", onKey));
onBeforeUnmount(() => window.removeEventListener("keydown", onKey));

// ---- Mutations -------------------------------------------------------------------------------
const addMeal = useAddPlannedFromSavedAt();
const removeMeal = useDeletePlannedMealAt();
const reorder = useReorderPlannedMealsAt();
const copyDay = useCopyPlannedDay();

// Several days can be busy at once, so pending is tracked per day / meal, not per mutation.
const busyDays = reactive(new Set<number>());
const busyMeals = reactive(new Set<number>());
const addingSaved = reactive(new Set<number>());

async function addTo(day: PlanDay, sm: SavedMeal) {
    busyDays.add(day.offset);
    addingSaved.add(sm.ID);
    try {
        // No success toast: the day's totals and chips update, and rapid adds would stack toasts.
        await addMeal.mutateAsync({ savedMealId: sm.ID, offset: day.offset });
    } catch {
        toast.push(`Couldn't add ${sm.name}`, "error");
    } finally {
        busyDays.delete(day.offset);
        addingSaved.delete(sm.ID);
    }
}

async function remove(day: PlanDay, mealId: number) {
    busyMeals.add(mealId);
    try {
        await removeMeal.mutateAsync({ plannedMealId: mealId, offset: day.offset });
    } catch {
        toast.push("Couldn't remove that meal", "error");
    } finally {
        busyMeals.delete(mealId);
    }
}

async function move(day: PlanDay, index: number, dir: -1 | 1) {
    const ids = moveId(
        day.meals.map((m) => m.id),
        index,
        dir,
    );
    if (!ids) return;
    busyDays.add(day.offset);
    try {
        await reorder.mutateAsync({ ids, offset: day.offset });
    } catch {
        toast.push("Couldn't change the order", "error");
    } finally {
        busyDays.delete(day.offset);
    }
}

async function clear(day: PlanDay) {
    const n = day.meals.length;
    const ok = await dialogManager.confirm({
        title: `Clear ${shortDayLabel(day.date)}?`,
        message: `Removes ${n} planned ${n === 1 ? "meal" : "meals"} (${formatInt(day.planned.calories)} kcal). Logged meals stay.`,
        confirmText: "Clear day",
        cancelText: "Cancel",
        tone: "danger",
    });
    if (!ok) return;
    busyDays.add(day.offset);
    try {
        for (const m of day.meals) {
            await removeMeal.mutateAsync({ plannedMealId: m.id, offset: day.offset });
        }
    } catch {
        toast.push("Couldn't clear the whole day", "error");
    } finally {
        busyDays.delete(day.offset);
    }
}

async function copy(from: PlanDay, to: number[]) {
    if (!to.length) return;
    for (const o of to) busyDays.add(o);
    try {
        await copyDay.mutateAsync({ from: from.offset, to });
        toast.push(
            `Copied ${shortDayLabel(from.date)} to ${to.length} ${to.length === 1 ? "day" : "days"}`,
            "success",
        );
        copyOpen.value = false;
    } catch {
        toast.push("Couldn't copy that day", "error");
    } finally {
        for (const o of to) busyDays.delete(o);
    }
}

// ---- Overlays --------------------------------------------------------------------------------
const addOpen = ref(false);
const addOffset = ref<number | null>(null);
// Read the day live from the week, so the overlay's totals follow each add.
const addDay = computed(() => days.value.find((d) => d.offset === addOffset.value) ?? null);

function onAdd(day: PlanDay) {
    if (armed.value) {
        void addTo(day, armed.value);
        return;
    }
    addOffset.value = day.offset;
    addOpen.value = true;
}

const copyOpen = ref(false);
const copySource = ref<PlanDay | null>(null);
function onCopyTo(day: PlanDay) {
    copySource.value = day;
    copyOpen.value = true;
}
</script>

<template>
    <div class="flex w-full flex-col gap-6 pt-2 pb-above-tabbar lg:pb-10">
        <PageHeader title="Plan meals">
            <template #meta>
                <span class="tabular-nums">{{ isCurrentWeek ? "This week" : rangeLabel }}</span>
                <template v-if="summary && !loading">
                    ·
                    <span :class="summary.tone === 'warn' ? 'text-warn' : summary.tone === 'success' ? 'text-success' : ''">{{
                        summary.text
                    }}</span>
                </template>
            </template>
            <template #actions>
                <IconButton label="Previous week" variant="secondary" @click="shiftWeek(7)">
                    <ChevronLeft class="size-5" />
                </IconButton>
                <Button v-if="!isCurrentWeek" size="sm" variant="ghost" @click="thisWeek">This week</Button>
                <span v-else class="min-w-28 text-center text-sm text-fg-muted tabular-nums">{{ rangeLabel }}</span>
                <IconButton label="Next week" variant="secondary" @click="shiftWeek(-7)">
                    <ChevronRight class="size-5" />
                </IconButton>
            </template>
        </PageHeader>

        <!-- Desktop tray: pick a saved meal, then add it to days with one click each. -->
        <div class="hidden flex-col gap-2 lg:flex">
            <div class="flex items-center gap-2">
                <h2 class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Saved meals</h2>
                <span class="text-xs text-fg-muted">
                    {{ armed ? `Click a day's button to add ${armed.name}. Esc to stop.` : "Pick one to add it to several days." }}
                </span>
            </div>
            <ErrorState
                v-if="savedError"
                compact
                title="Couldn't load saved meals"
                :error="savedError"
                @retry="refetchSaved()"
            />
            <div v-else-if="savedPending" class="flex gap-2">
                <Skeleton v-for="i in 3" :key="i" shape="block" width="9rem" height="2.5rem" />
            </div>
            <p v-else-if="!savedMeals.length" class="m-0 text-sm text-fg-muted">
                No saved meals yet.
                <Button :to="createSavedTo" size="sm" variant="ghost">New saved meal</Button>
            </p>
            <div v-else class="flex flex-wrap items-center gap-2" role="group" aria-label="Saved meals">
                <Button
                    v-for="sm in savedMeals"
                    :key="sm.ID"
                    size="sm"
                    :variant="armed?.ID === sm.ID ? 'primary' : 'secondary'"
                    :aria-pressed="armed?.ID === sm.ID"
                    @click="toggleArmed(sm)"
                >
                    {{ sm.name }}
                    <span class="text-xs tabular-nums opacity-70"
                        >{{ formatInt(savedMealMacros(sm).calories) }} kcal</span
                    >
                </Button>
                <IconButton v-if="armed" label="Stop adding" size="sm" @click="armed = null">
                    <X class="size-4" />
                </IconButton>
            </div>
        </div>

        <ErrorState
            v-if="failed"
            title="Couldn't load this week"
            :error="failed.error"
            @retry="queries.forEach((q) => q.isError && q.refetch())"
        />
        <template v-else-if="loading">
            <div v-if="showSkeleton" class="flex flex-col gap-3" role="status" aria-busy="true">
                <span class="sr-only">Loading the week</span>
                <Skeleton v-for="i in 7" :key="i" shape="block" height="6rem" />
            </div>
        </template>
        <ol v-else class="m-0 flex list-none flex-col gap-3 p-0" :aria-label="`Week of ${rangeLabel}`">
            <li v-for="(day, i) in days" :key="day.key" class="min-w-0">
                <PlanDayCard
                    :day="day"
                    :previous="days[i - 1]"
                    :armed-name="armed?.name ?? null"
                    :busy="busyDays.has(day.offset)"
                    :busy-meal-ids="busyMeals"
                    @add="onAdd(day)"
                    @copy-to="onCopyTo(day)"
                    @copy-from-previous="days[i - 1] && copy(days[i - 1]!, [day.offset])"
                    @clear="clear(day)"
                    @move="(idx, dir) => move(day, idx, dir)"
                    @remove="(id) => remove(day, id)"
                />
            </li>
        </ol>

        <AddMealOverlay
            v-model:open="addOpen"
            :day="addDay"
            :saved-meals="savedMeals"
            :pending="savedPending"
            :error="savedError"
            :adding="addingSaved"
            @add="(sm) => addDay && addTo(addDay, sm)"
            @retry="refetchSaved()"
        />
        <CopyDayOverlay
            v-model:open="copyOpen"
            :source="copySource"
            :pending="copyDay.isPending.value"
            @copy="(to) => copySource && copy(copySource, to)"
        />
    </div>
</template>
