<script setup lang="ts">
import { ChevronDown, ChevronUp, Plus, SquarePen, Trash2 } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { previewSavedMealDelete } from "~/api/diet/api";
import type { SavedMealDeleteDependentsInfo } from "~/api/diet/api";
import type { PlannedMeal, SavedMeal } from "~/types/diet";
import {
    useAddPlannedFromSaved,
    useDeletePlannedMeal,
    useDietLogsToday,
    useReorderPlannedMeals,
} from "~/api/diet/queries";
import { useDeleteSavedMeal } from "~/pages/diet/logmeal/queries/useMealMutations";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { withDietDayOffsetQuery } from "~/utils/dateUtil";
import { EDIT_SAVED_TYPE } from "~/pages/diet/logmeal/logmealMode";
import {
    Card,
    EmptyState,
    ErrorState,
    IconButton,
    List,
    ListRow,
    Section,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { macroTotals, offsetToHeadingLabel } from "./planUtils";
import { useSavedMeals } from "./useSavedMeals";

const props = defineProps<{ dateOffset: number }>();

const router = useRouter();
const dateOffset = computed(() => props.dateOffset);

const {
    data: dayData,
    isPending: dayPending,
    error: dayError,
    refetch: refetchDay,
} = useDietLogsToday(dateOffset);
const showDaySkeleton = useDelayedFlag(dayPending);
const sortedPlannedMeals = computed(() =>
    [...(dayData.value?.day.plannedMeals ?? [])].sort(
        (a, b) =>
            (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID,
    ),
);

const {
    savedMeals,
    isPending: savedPending,
    error: savedError,
    refetch: refetchSaved,
} = useSavedMeals();
const showSavedSkeleton = useDelayedFlag(savedPending);

const deletePlanned = useDeletePlannedMeal(dateOffset);
const reorderPlanned = useReorderPlannedMeals(dateOffset);
const addFromSaved = useAddPlannedFromSaved(dateOffset);
const deleteSavedTemplate = useDeleteSavedMeal();
const busySavedId = ref<number | null>(null);
const plannedBusy = computed(
    () => deletePlanned.isPending.value || reorderPlanned.isPending.value,
);
const savedBusy = computed(
    () =>
        addFromSaved.isPending.value ||
        deleteSavedTemplate.isPending.value ||
        busySavedId.value != null,
);

async function removePlanned(pm: PlannedMeal) {
    try {
        await deletePlanned.mutateAsync(pm.ID);
        toast.push("Removed from planned", "success");
    } catch {
        toast.push("Could not remove planned meal", "error");
    }
}

async function movePlanned(index: number, dir: -1 | 1) {
    const ordered = sortedPlannedMeals.value.map((p) => p.ID);
    const j = index + dir;
    if (j < 0 || j >= ordered.length) return;
    const next = [...ordered];
    const t = next[index];
    next[index] = next[j];
    next[j] = t;
    try {
        await reorderPlanned.mutateAsync(next);
    } catch {
        toast.push("Could not update order", "error");
    }
}

async function addSavedToSelectedDay(sm: SavedMeal) {
    busySavedId.value = sm.ID;
    try {
        await addFromSaved.mutateAsync(sm.ID);
        toast.push(`Added "${sm.name}" to planned`, "success");
    } catch {
        toast.push("Could not add to planned", "error");
    } finally {
        busySavedId.value = null;
    }
}

async function deleteSavedMealTemplate(sm: SavedMeal) {
    busySavedId.value = sm.ID;
    try {
        let preview: SavedMealDeleteDependentsInfo;
        try {
            preview = await previewSavedMealDelete(sm.ID);
        } catch {
            toast.push("Could not load saved meal", "error");
            return;
        }
        const ok = await dialogManager.confirm({
            title:
                preview.reference_count > 0
                    ? "Saved meal is on your plan"
                    : "Delete saved meal",
            tone: "danger",
            message:
                preview.reference_count > 0
                    ? `This meal is referenced ${preview.reference_count} times in your plan. Deleting removes those planned meals. Delete “${sm.name}”?`
                    : `Delete “${sm.name}”?`,
            confirmText: "Delete",
            cancelText: "Cancel",
        });
        if (!ok) return;
        try {
            await deleteSavedTemplate.mutateAsync({
                savedMealId: sm.ID,
                force: true,
            });
            toast.push("Saved meal deleted", "success");
        } catch {
            toast.push("Could not delete saved meal", "error");
        }
    } finally {
        busySavedId.value = null;
    }
}

function openEditSavedMeal(sm: SavedMeal) {
    router.push({
        name: "diet-log",
        query: withDietDayOffsetQuery(props.dateOffset, {
            type: EDIT_SAVED_TYPE,
            id: String(sm.ID),
        }),
    });
}
</script>

<template>
    <div class="flex flex-col gap-6">
        <p class="m-0 text-sm text-fg-muted">
            Editing:
            <span class="font-medium text-fg">{{
                offsetToHeadingLabel(dateOffset)
            }}</span>
        </p>
        <ErrorState
            v-if="dayError"
            title="Couldn't load this day"
            :error="dayError"
            @retry="refetchDay()"
        />
        <template v-else>
            <Section title="Currently planned">
                <Card padding="none">
                    <SkeletonList
                        v-if="showDaySkeleton"
                        :rows="2"
                        :leading="false"
                    />
                    <EmptyState
                        v-else-if="!dayPending && !sortedPlannedMeals.length"
                        title="Nothing planned yet"
                        description="Add a saved meal below."
                    />
                    <List v-else-if="sortedPlannedMeals.length">
                        <ListRow
                            v-for="(pm, index) in sortedPlannedMeals"
                            :key="pm.ID"
                            :pending="
                                deletePlanned.isPending.value &&
                                deletePlanned.variables.value === pm.ID
                            "
                        >
                            {{ pm.meal.name }}
                            <template v-if="pm.meal.items?.length" #subtitle>
                                <SimpleMacros
                                    decimals="none"
                                    v-bind="macroTotals(pm.meal.items)"
                                />
                            </template>
                            <template #end>
                                <IconButton
                                    label="Move up"
                                    size="sm"
                                    :disabled="index === 0 || plannedBusy"
                                    @click="movePlanned(index, -1)"
                                >
                                    <ChevronUp class="size-5" />
                                </IconButton>
                                <IconButton
                                    label="Move down"
                                    size="sm"
                                    :disabled="
                                        index === sortedPlannedMeals.length - 1 ||
                                        plannedBusy
                                    "
                                    @click="movePlanned(index, 1)"
                                >
                                    <ChevronDown class="size-5" />
                                </IconButton>
                                <IconButton
                                    :label="`Remove ${pm.meal.name}`"
                                    variant="danger"
                                    size="sm"
                                    :disabled="plannedBusy"
                                    @click="removePlanned(pm)"
                                >
                                    <Trash2 class="size-5" />
                                </IconButton>
                            </template>
                        </ListRow>
                    </List>
                </Card>
            </Section>
            <Section title="Add from saved meals">
                <Card padding="none">
                    <ErrorState
                        v-if="savedError"
                        compact
                        title="Couldn't load saved meals"
                        :error="savedError"
                        @retry="refetchSaved()"
                    />
                    <SkeletonList
                        v-else-if="showSavedSkeleton"
                        :rows="3"
                        :leading="false"
                    />
                    <EmptyState
                        v-else-if="!savedPending && !savedMeals.length"
                        title="No saved meals yet"
                        description="Create one with New saved meal on the Diet page."
                    />
                    <List v-else-if="savedMeals.length">
                        <ListRow
                            v-for="sm in savedMeals"
                            :key="sm.ID"
                            :pending="busySavedId === sm.ID"
                            :disabled="savedBusy"
                            @click="addSavedToSelectedDay(sm)"
                        >
                            <template #leading><Plus class="size-4" /></template>
                            {{ sm.name }}
                            <template v-if="sm.items?.length" #subtitle>
                                <SimpleMacros
                                    decimals="none"
                                    v-bind="macroTotals(sm.items)"
                                />
                            </template>
                            <template #end>
                                <IconButton
                                    :label="`Edit ${sm.name}`"
                                    size="sm"
                                    :disabled="savedBusy"
                                    @click="openEditSavedMeal(sm)"
                                >
                                    <SquarePen class="size-5" />
                                </IconButton>
                                <IconButton
                                    :label="`Delete ${sm.name}`"
                                    variant="danger"
                                    size="sm"
                                    :disabled="savedBusy"
                                    @click="deleteSavedMealTemplate(sm)"
                                >
                                    <Trash2 class="size-5" />
                                </IconButton>
                            </template>
                        </ListRow>
                    </List>
                </Card>
            </Section>
        </template>
    </div>
</template>
