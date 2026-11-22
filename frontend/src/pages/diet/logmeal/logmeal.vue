<script setup lang="ts">
import SearchList from "~/shared/SearchList.vue";
import LogMealGroupBlock from "./LogMealGroupBlock.vue";
import LogMealItemRow from "./LogMealItemRow.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
import {
    Button,
    Card,
    EmptyState,
    ErrorState,
    IconButton,
    Input,
    List,
    Menu,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { computed, nextTick, ref, watch } from "vue";
import { useMeal, useSavedMeal } from "~/api/diet/queries";
import { useDietLogsToday } from "~/api/diet/queries";
import SavedMealDisplay from "./components/SavedMealDisplay.vue";
import { savedMealToMeal } from "~/utils/savedMealToMeal";
import {
    EDIT_VARIANT,
    PAGE_MODE,
    parseEditMealVariant,
} from "./logmealMode";
import LogMealSummary from "./LogMealSummary.vue";
import { MoreVertical, UtensilsCrossed } from "lucide-vue-next";
import { defaultMealName } from "./domain/mealCard";
import { useLogMealMode } from "./useLogMealMode";
import {
    cloneMeal,
    macrosForMeal,
    useEditableMeal,
} from "./useEditableMeal";
import { useLogMealDraftStorage } from "./useLogMealDraftStorage";
import { useLogMealSubmit } from "./useLogMealSubmit";

const {
    dayOffset,
    queryType,
    pageMode,
    editVariant,
    id,
    mealLogDayId,
    mealId,
    savedMealEditId,
    editMissingId,
    pageTitle,
    showDietDayMacroBars,
} = useLogMealMode();

const {
    data: mealData,
    error: mealError,
    isLoading: mealLoading,
    refetch: refetchMeal,
} = useMeal(mealId);
const {
    data: savedMealData,
    error: savedMealError,
    isLoading: savedMealLoading,
    refetch: refetchSavedMeal,
} = useSavedMeal(savedMealEditId);

const editLoading = computed(
    () => mealLoading.value || savedMealLoading.value,
);
const showEditSkeleton = useDelayedFlag(editLoading);
function retryLoad() {
    void (editVariant.value === EDIT_VARIANT.saved
        ? refetchSavedMeal()
        : refetchMeal());
}

const mealLoadError = computed(() => {
    if (
        pageMode.value === PAGE_MODE.edit &&
        editVariant.value === EDIT_VARIANT.saved
    ) {
        return savedMealError.value;
    }
    return mealError.value;
});

const { data: dayLogs } = useDietLogsToday(dayOffset);

const {
    meal,
    baselineMealMacros,
    selectedForGroup,
    mealItemBlocks,
    totalMacros,
    resetMeal,
    setMealItemAmount,
    isGroupExpanded,
    toggleGroupCollapse,
    toggleSelectRow,
    setGroupLabel,
    groupSelectedRows,
    ungroupSelectedRows,
    removeGroupLines,
    pickFoodOrComposite,
    createFood,
    removeFood,
    swapVariant,
    setMeal,
} = useEditableMeal();

const draftStorage = useLogMealDraftStorage({
    meal,
    pageMode,
    routeMealId: id,
});

const {
    logMealToDay,
    saveSavedMealTemplate,
    logEditedMeal,
    updateLoggedMeal,
    saveEditedSavedMeal,
} = useLogMealSubmit({
    meal,
    dayOffset,
    editVariant,
    routeMealId: id,
    mealLogDayId,
    draftStorage,
});

// The submit functions toast their own failures; this only drives the pending state.
const submitting = ref(false);
async function submit(fn: () => Promise<void>) {
    if (submitting.value) return;
    submitting.value = true;
    try {
        await fn();
    } finally {
        submitting.value = false;
    }
}

let logMealDraftHydrateGeneration = 0;

watch(
    [pageMode, id, mealData, savedMealData, queryType],
    ([mode, newId, newMealData, newSavedData, qType]) => {
        const variant = parseEditMealVariant(qType);
        draftStorage.setSaveEnabled(false);
        logMealDraftHydrateGeneration++;
        if (
            (mode === PAGE_MODE.log || mode === PAGE_MODE.create) &&
            newId === 0
        ) {
            resetMeal();
            baselineMealMacros.value = null;
            const hydrateGen = logMealDraftHydrateGeneration;
            void nextTick(() => {
                if (hydrateGen !== logMealDraftHydrateGeneration) return;
                draftStorage.restore();
                if (mode === PAGE_MODE.log && !meal.value.name.trim()) {
                    meal.value.name = defaultMealName(new Date().getHours());
                }
                draftStorage.setSaveEnabled(true);
            });
            return;
        }
        if (mode === PAGE_MODE.edit && newId !== 0) {
            if (variant === EDIT_VARIANT.saved) {
                if (newSavedData) {
                    meal.value = cloneMeal(savedMealToMeal(newSavedData));
                } else {
                    resetMeal();
                }
                baselineMealMacros.value = null;
                return;
            }
            if (newMealData?.meal) {
                meal.value = cloneMeal(newMealData.meal);
                baselineMealMacros.value =
                    variant === EDIT_VARIANT.logged
                        ? macrosForMeal(meal.value)
                        : null;
            } else {
                resetMeal();
                baselineMealMacros.value = null;
            }
        }
    },
    { immediate: true },
);

const macroBarsDayTotals = computed(() => {
    const t = dayLogs.value;
    const tm = totalMacros.value;
    const dayCal = t?.totalCalories ?? 0;
    const dayProtein = t?.totalProtein ?? 0;
    const dayFiber = t?.totalFiber ?? 0;
    const dayCarbs = t?.totalCarbs ?? 0;
    const dayFat = t?.totalFat ?? 0;

    if (
        pageMode.value === PAGE_MODE.edit &&
        editVariant.value === EDIT_VARIANT.logged &&
        baselineMealMacros.value
    ) {
        const b = baselineMealMacros.value;
        return {
            totalCalories: dayCal - b.calories + tm.calories,
            totalProtein: dayProtein - b.protein + tm.protein,
            totalFiber: dayFiber - b.fiber + tm.fiber,
            totalCarbs: dayCarbs - b.carbs + tm.carbs,
            totalFat: dayFat - b.fat + tm.fat,
        };
    }
    return {
        totalCalories: dayCal + tm.calories,
        totalProtein: dayProtein + tm.protein,
        totalFiber: dayFiber + tm.fiber,
        totalCarbs: dayCarbs + tm.carbs,
        totalFat: dayFat + tm.fat,
    };
});

const daySummaryTotals = computed(() => {
    const t = macroBarsDayTotals.value;
    return {
        calories: t.totalCalories,
        protein: t.totalProtein,
        carbs: t.totalCarbs,
        fat: t.totalFat,
        fiber: t.totalFiber,
    };
});

const mealItemCount = computed(() => meal.value.items.length);

// Grouping lives behind Select mode so rows stay clean by default.
const selecting = ref(false);
const selectedCount = computed(
    () => Object.values(selectedForGroup.value).filter(Boolean).length,
);
function toggleSelecting() {
    selecting.value = !selecting.value;
    if (!selecting.value) selectedForGroup.value = {};
}
watch(mealItemCount, (n) => {
    if (n < 2 && selecting.value) toggleSelecting();
});
// Group/ungroup clear the selection on success; leave Select mode then.
function groupAndExit() {
    groupSelectedRows();
    if (selectedCount.value === 0) selecting.value = false;
}
function ungroupAndExit() {
    ungroupSelectedRows();
    if (selectedCount.value === 0) selecting.value = false;
}

const logMenuItems = [{ key: "log-save", label: "Log and save meal" }];

// One primary action per mode ("Log and save" lives in the menu beside it).
const primaryAction = computed(() => {
    if (pageMode.value === PAGE_MODE.create) return { label: "Save meal", run: saveSavedMealTemplate };
    if (pageMode.value === PAGE_MODE.log) return { label: "Log meal", run: () => logMealToDay(false) };
    if (editVariant.value === EDIT_VARIANT.saved) return { label: "Save changes", run: saveEditedSavedMeal };
    if (editVariant.value === EDIT_VARIANT.logged) return { label: "Update", run: updateLoggedMeal };
    if (editVariant.value === EDIT_VARIANT.planned) return { label: "Log", run: logEditedMeal };
    return null;
});
</script>

<template>
    <div class="flex w-full flex-col gap-4 pb-above-tabbar lg:pb-4">
        <p v-if="editMissingId" class="m-0 text-sm text-danger" role="alert">
            Missing meal id for this edit.
        </p>
        <ErrorState
            v-else-if="mealLoadError && id !== 0"
            title="Couldn't load this meal"
            :error="mealLoadError"
            @retry="retryLoad"
        />
        <SkeletonList
            v-else-if="showEditSkeleton"
            :rows="4"
            :leading="false"
        />
        <div
            v-else-if="meal && !editLoading"
            class="grid grid-cols-1 gap-4 lg:h-[calc(100dvh-8rem)] lg:min-h-[36rem] lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)] lg:grid-rows-2"
        >
            <!-- Bento at lg: the meal card spans both rows and scrolls its item list; the two pickers split the side column. -->
            <Card padding="none" as="article" fill class="lg:row-span-2">
                <header class="flex shrink-0 flex-col gap-4 border-b border-line p-4">
                    <div class="flex flex-col gap-2">
                        <h1 class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                            {{ pageTitle }}
                        </h1>
                        <div class="flex items-center gap-2">
                            <div class="min-w-0 flex-1">
                                <Input
                                    v-model="meal.name"
                                    aria-label="Meal name"
                                    placeholder="Name this meal"
                                    required
                                />
                            </div>
                            <Button
                                v-if="mealItemCount >= 2"
                                variant="ghost"
                                :aria-pressed="selecting"
                                @click="toggleSelecting"
                            >
                                {{ selecting ? "Done" : "Select" }}
                            </Button>
                        </div>
                    </div>
                    <LogMealSummary
                        :meal="totalMacros"
                        :day-after="showDietDayMacroBars ? daySummaryTotals : undefined"
                        :plan="dayLogs?.day.plan"
                    />
                </header>
                <section class="flex min-h-0 flex-1 flex-col" aria-labelledby="meal-items-heading">
                    <div class="flex min-h-touch shrink-0 items-center justify-between gap-3 px-4 py-2 lg:min-h-row">
                        <h2 id="meal-items-heading" class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                            Items <span class="tabular-nums">{{ mealItemCount }}</span>
                        </h2>
                        <div v-if="selecting" class="flex items-center gap-2">
                            <span class="text-sm text-fg-muted tabular-nums" aria-live="polite">{{ selectedCount }} selected</span>
                            <Button size="sm" :disabled="selectedCount < 2" @click="groupAndExit">Group</Button>
                            <Button size="sm" :disabled="selectedCount < 1" @click="ungroupAndExit">Ungroup</Button>
                        </div>
                    </div>
                    <EmptyState
                        v-if="!mealItemBlocks.length"
                        class="border-t border-line"
                        :icon="UtensilsCrossed"
                        title="No foods yet"
                        description="Search Add foods, or start from a saved meal."
                    />
                    <List v-else class="min-h-0 flex-1 overflow-y-auto border-t border-line">
                        <template
                            v-for="(block, bi) in mealItemBlocks"
                            :key="'b-' + bi"
                        >
                            <template v-if="block.kind === 'ungrouped'">
                                <LogMealItemRow
                                    v-for="{ item, index: i } in block.rows"
                                    :key="`u-${i}`"
                                    :item="item"
                                    :row-index="i"
                                    :selected="!!selectedForGroup[i]"
                                    :selecting="selecting"
                                    :meal-calories="totalMacros.calories"
                                    @toggle-select="toggleSelectRow"
                                    @set-item-amount="setMealItemAmount"
                                    @swap-variant="swapVariant"
                                    @remove="removeFood"
                                />
                            </template>
                            <LogMealGroupBlock
                                v-else
                                :block="block"
                                :expanded="isGroupExpanded(block.groupId)"
                                :selected-for-group="selectedForGroup"
                                :selecting="selecting"
                                :meal-calories="totalMacros.calories"
                                @toggle-collapse="toggleGroupCollapse"
                                @set-group-label="setGroupLabel"
                                @remove-group="removeGroupLines"
                                @toggle-select="toggleSelectRow"
                                @set-item-amount="setMealItemAmount"
                                @swap-variant="swapVariant"
                                @remove-item="removeFood"
                            />
                        </template>
                    </List>
                </section>
                <footer
                    class="sticky bottom-above-tabbar z-raised flex shrink-0 items-center gap-2 rounded-b-card border-t border-line bg-surface p-4 lg:static lg:justify-end"
                >
                    <Menu v-if="pageMode === PAGE_MODE.log" :items="logMenuItems" @select="submit(() => logMealToDay(true))">
                        <IconButton label="More log options" variant="secondary" :disabled="submitting">
                            <MoreVertical class="size-5" />
                        </IconButton>
                    </Menu>
                    <Button
                        v-if="primaryAction"
                        class="flex-1 lg:flex-none"
                        variant="primary"
                        :loading="submitting"
                        @click="submit(primaryAction.run)"
                    >
                        {{ primaryAction.label }}
                    </Button>
                </footer>
            </Card>
            <Card as="aside" fill>
                <div class="flex h-[26rem] flex-col gap-2 lg:h-auto lg:min-h-0 lg:flex-1">
                    <h2 class="m-0 text-lg font-semibold">Add foods</h2>
                    <SearchList
                        :route="'diet/meals/food/all'"
                        :onSelect="pickFoodOrComposite"
                        :onCreate="createFood"
                        :displayComponent="FoodDisplay"
                    />
                </div>
            </Card>
            <Card as="aside" fill>
                <div class="flex h-[22rem] flex-col gap-2 lg:h-auto lg:min-h-0 lg:flex-1">
                    <h2 class="m-0 text-lg font-semibold">Saved meals</h2>
                    <SearchList
                        :key="meal.ID"
                        :route="'diet/meals/saved-meal/all'"
                        :on-select="setMeal"
                        :display-component="SavedMealDisplay"
                    />
                </div>
            </Card>
        </div>
    </div>
</template>
