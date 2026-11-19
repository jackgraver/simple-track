<script setup lang="ts">
import SearchList from "~/shared/SearchList.vue";
import LogMealGroupBlock from "./LogMealGroupBlock.vue";
import LogMealItemRow from "./LogMealItemRow.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
import {
    Button,
    Card,
    ErrorState,
    Field,
    Input,
    List,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { computed, nextTick, ref, watch } from "vue";
import { useMeal } from "./queries/useMeal";
import { useSavedMeal } from "./queries/useSavedMeal";
import { useDietLogsToday } from "~/api/diet/queries";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import { savedMealToMeal } from "~/utils/savedMealToMeal";
import {
    EDIT_VARIANT,
    PAGE_MODE,
    parseEditMealVariant,
} from "./logmealMode";
import SimpleMacros from "~/shared/SimpleMacros.vue";
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
</script>

<template>
    <div class="flex w-full flex-col gap-4 pb-above-tabbar lg:pb-10">
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
            class="grid grid-cols-1 items-start gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]"
        >
            <Card padding="none" as="article">
                <header class="flex flex-col gap-4 border-b border-line p-4">
                    <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                        <h1 class="m-0 text-xl font-semibold tracking-tight">
                            {{ pageTitle }}
                        </h1>
                        <SimpleMacros
                            :calories="totalMacros.calories"
                            :protein="totalMacros.protein"
                            :fat="totalMacros.fat"
                            :carbs="totalMacros.carbs"
                            :fiber="totalMacros.fiber"
                        />
                    </div>
                    <Field label="Meal name" required>
                        <Input v-model="meal.name" required />
                    </Field>
                </header>
                <section class="flex flex-col">
                    <div class="flex items-center justify-between gap-3 px-4 py-3">
                        <h2 class="m-0 text-base font-semibold">Meal items</h2>
                        <div class="flex flex-wrap justify-end gap-2">
                            <Button size="sm" @click="groupSelectedRows">
                                Group selected
                            </Button>
                            <Button size="sm" @click="ungroupSelectedRows">
                                Ungroup
                            </Button>
                        </div>
                    </div>
                    <p
                        v-if="!mealItemBlocks.length"
                        class="m-0 border-t border-line px-4 py-8 text-center text-sm text-fg-muted"
                    >
                        No items yet. Pick a food from Add foods.
                    </p>
                    <List v-else class="border-t border-line">
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
                <footer class="flex flex-col gap-4 border-t border-line p-4">
                    <div class="flex w-full flex-col gap-2 lg:flex-row lg:justify-end">
                        <Button
                            v-if="pageMode === PAGE_MODE.create"
                            variant="primary"
                            :loading="submitting"
                            @click="submit(saveSavedMealTemplate)"
                        >
                            Save meal
                        </Button>
                        <template v-if="pageMode === PAGE_MODE.log">
                            <Button
                                variant="secondary"
                                :disabled="submitting"
                                @click="submit(() => logMealToDay(true))"
                            >
                                Log and save meal
                            </Button>
                            <Button
                                variant="primary"
                                :loading="submitting"
                                @click="submit(() => logMealToDay(false))"
                            >
                                Log meal
                            </Button>
                        </template>
                        <Button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.saved
                            "
                            variant="primary"
                            :loading="submitting"
                            @click="submit(saveEditedSavedMeal)"
                        >
                            Save changes
                        </Button>
                        <Button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.logged
                            "
                            variant="primary"
                            :loading="submitting"
                            @click="submit(updateLoggedMeal)"
                        >
                            Update
                        </Button>
                        <Button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.planned
                            "
                            variant="primary"
                            :loading="submitting"
                            @click="submit(logEditedMeal)"
                        >
                            Log
                        </Button>
                    </div>
                    <MacroBars
                        v-if="showDietDayMacroBars"
                        :total-calories="macroBarsDayTotals.totalCalories"
                        :total-protein="macroBarsDayTotals.totalProtein"
                        :total-fat="macroBarsDayTotals.totalFat"
                        :total-carbs="macroBarsDayTotals.totalCarbs"
                        :planned-calories="dayLogs?.day.plan.calories ?? 0"
                        :planned-protein="dayLogs?.day.plan.protein ?? 0"
                        :planned-fat="dayLogs?.day.plan.fat ?? 0"
                        :planned-carbs="dayLogs?.day.plan.carbs ?? 0"
                    />
                </footer>
            </Card>
            <div class="flex flex-col gap-4 lg:sticky lg:top-4">
                <Card as="aside">
                    <div class="flex h-[26rem] flex-col gap-2">
                        <h2 class="m-0 text-lg font-semibold">Add foods</h2>
                        <SearchList
                            :route="'diet/meals/food/all'"
                            :onSelect="pickFoodOrComposite"
                            :onCreate="createFood"
                            :displayComponent="FoodDisplay"
                        />
                    </div>
                </Card>
                <Card as="aside">
                    <div class="flex h-[22rem] flex-col gap-2">
                        <h2 class="m-0 text-lg font-semibold">Saved meals</h2>
                        <SearchList
                            :key="meal.ID"
                            :route="'diet/meals/saved-meal/all'"
                            :on-select="setMeal"
                        />
                    </div>
                </Card>
            </div>
        </div>
    </div>
</template>
