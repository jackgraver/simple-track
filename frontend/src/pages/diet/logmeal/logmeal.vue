<script setup lang="ts">
import "~/pages/diet/macro-colors.css";
import SearchList from "~/shared/SearchList.vue";
import LogMealGroupBlock from "./LogMealGroupBlock.vue";
import LogMealItemRow from "./LogMealItemRow.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
import { mealItemsListGridClass } from "./mealItemsListGrid";
import Input from "~/shared/input/Input.vue";
import { computed, nextTick, watch } from "vue";
import { useMeal } from "./queries/useMeal";
import { useSavedMeal } from "./queries/useSavedMeal";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
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

const { data: mealData, error: mealError } = useMeal(mealId);
const { data: savedMealData, error: savedMealError } =
    useSavedMeal(savedMealEditId);

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
    amountPlusMinus,
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
    <div class="diet-module flex flex-col items-center pt-4">
        <div
            v-if="editMissingId"
            class="flex h-[60dvh] items-center justify-center p-8 text-cfRed"
        >
            <span>Missing meal id for this edit.</span>
        </div>
        <div
            v-else-if="mealLoadError && id !== 0"
            class="flex h-[60dvh] items-center justify-center p-8 text-cfRed"
        >
            <span
                >Error loading meal:
                {{ mealLoadError?.message || "Unknown error" }}</span
            >
        </div>
        <div
            v-else-if="meal"
            class="flex w-full flex-col gap-4 md:grid md:h-[calc(100dvh-7rem)] md:grid-cols-[2fr_1fr] md:grid-rows-2 pb-16 md:pb-0"
        >
            <article
                class="flex min-h-[min(28rem,55dvh)] flex-col overflow-hidden rounded-lg bg-firstBg md:row-span-2 md:min-h-0"
            >
                <header
                    class="shrink-0 border-b border-secondBg p-4 text-textPrimary"
                >
                    <div
                        class="mb-4 flex flex-row items-center justify-between"
                    >
                        <h1 class="m-0 text-xl font-semibold">
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
                    <div class="flex min-w-0 flex-1 flex-col gap-2">
                        <Input
                            label="Meal Name"
                            type="text"
                            v-model="meal.name"
                            required
                        />
                    </div>
                </header>
                <section
                    class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden text-textPrimary"
                >
                    <div
                        class="flex shrink-0 items-center justify-between gap-3 px-4 py-2.5"
                    >
                        <h3
                            class="m-0 min-w-0 text-base font-medium leading-tight"
                        >
                            Meal Items
                        </h3>
                        <div
                            class="flex shrink-0 flex-wrap items-center justify-end gap-2"
                        >
                            <button
                                type="button"
                                class="rounded border border-secondBg bg-secondBg px-3 py-1.5 text-sm hover:bg-thirdBg"
                                @click="groupSelectedRows"
                            >
                                Group selected
                            </button>
                            <button
                                type="button"
                                class="rounded border border-secondBg bg-secondBg px-3 py-1.5 text-sm hover:bg-thirdBg"
                                @click="ungroupSelectedRows"
                            >
                                Ungroup
                            </button>
                        </div>
                    </div>
                    <div
                        class="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto px-4 pb-1"
                    >
                        <div
                            :class="[
                                mealItemsListGridClass,
                                'mb-1 border-b border-secondBg pb-2 text-xs font-medium text-textSecondary sm:grid',
                            ]"
                        >
                            <span><span class="sr-only">Select</span></span>
                            <span class="min-w-0">Item</span>
                            <span class="min-w-0 text-center">Qty</span>
                            <span class="hidden md:block min-w-0 text-right"
                                >Macros</span
                            >
                            <span
                                class="flex h-9 w-9 shrink-0 justify-self-end"
                                aria-hidden="true"
                            ></span>
                        </div>
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
                                    @amount-plus-minus="amountPlusMinus"
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
                                @amount-plus-minus="amountPlusMinus"
                                @set-item-amount="setMealItemAmount"
                                @swap-variant="swapVariant"
                                @remove-item="removeFood"
                            />
                        </template>
                    </div>
                </section>
                <footer
                    class="flex shrink-0 flex-col gap-4 border-t border-secondBg p-4"
                >
                    <div class="flex w-full flex-row gap-3">
                        <button
                            v-if="pageMode === PAGE_MODE.create"
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="saveSavedMealTemplate"
                        >
                            Save Meal
                        </button>
                        <button
                            v-if="pageMode === PAGE_MODE.log"
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="logMealToDay(false)"
                        >
                            Log Meal
                        </button>
                        <button
                            v-if="pageMode === PAGE_MODE.log"
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="logMealToDay(true)"
                        >
                            Log and Save Meal
                        </button>
                        <button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.saved
                            "
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="saveEditedSavedMeal"
                        >
                            Save changes
                        </button>
                        <button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.logged
                            "
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="updateLoggedMeal"
                        >
                            Update
                        </button>
                        <button
                            v-if="
                                pageMode === PAGE_MODE.edit &&
                                editVariant === EDIT_VARIANT.planned
                            "
                            class="flex-1 cursor-pointer rounded bg-secondBg px-5 py-2.5 text-sm text-textPrimary hover:bg-thirdBg"
                            type="button"
                            @click="logEditedMeal"
                        >
                            Log
                        </button>
                    </div>
                    <MacroBars
                        v-if="showDietDayMacroBars"
                        :totalCalories="macroBarsDayTotals.totalCalories"
                        :totalProtein="macroBarsDayTotals.totalProtein"
                        :totalFat="macroBarsDayTotals.totalFat"
                        :totalCarbs="macroBarsDayTotals.totalCarbs"
                        :planned-calories="dayLogs?.day.plan.calories ?? 0"
                        :planned-protein="dayLogs?.day.plan.protein ?? 0"
                        :planned-fat="dayLogs?.day.plan.fat ?? 0"
                        :planned-carbs="dayLogs?.day.plan.carbs ?? 0"
                    />
                </footer>
            </article>
            <aside
                class="flex min-h-[min(20rem,45dvh)] flex-col overflow-hidden rounded-lg bg-firstBg p-4 text-textPrimary md:min-h-0"
            >
                <h2 class="mt-0 text-lg font-semibold">Add Foods</h2>
                <SearchList
                    :route="'diet/meals/food/all'"
                    :onSelect="pickFoodOrComposite"
                    :onCreate="createFood"
                    :displayComponent="FoodDisplay"
                />
            </aside>
            <aside
                class="flex min-h-[min(20rem,45dvh)] flex-col overflow-hidden rounded-lg bg-firstBg p-4 text-textPrimary md:min-h-0"
            >
                <h2 class="mt-0 text-lg font-semibold">Select Saved Meal</h2>
                <SearchList
                    :key="meal.ID"
                    :route="'diet/meals/saved-meal/all'"
                    :on-select="setMeal"
                />
            </aside>
        </div>
    </div>
</template>
