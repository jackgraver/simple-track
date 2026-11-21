<script setup lang="ts">
import { ChevronDown, ChevronUp } from "lucide-vue-next";
import { computed, ref, watch } from "vue";
import { useQueryClient } from "@tanstack/vue-query";
import { addPlannedMealFromSaved } from "~/api/diet/api";
import { dietKeys } from "~/api/diet/keys";
import type { SavedMeal } from "~/types/diet";
import { toast } from "~/composables/toast/useToast";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import {
    Button,
    Card,
    Checkbox,
    EmptyState,
    ErrorState,
    Field,
    IconButton,
    Input,
    List,
    ListRow,
    Section,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { enumerateOffsetsInclusive, macroTotals } from "./planUtils";
import { useSavedMeals } from "./useSavedMeals";

const queryClient = useQueryClient();
const {
    savedMeals,
    isPending: savedPending,
    error: savedError,
    refetch: refetchSaved,
} = useSavedMeals();
const showSkeleton = useDelayedFlag(savedPending);

const bulkFrom = ref("");
const bulkTo = ref("");
const bulkSavedOrder = ref<number[]>([]);
const bulkSaving = ref(false);

// A saved meal deleted elsewhere on the page must not stay in the selection.
watch(savedMeals, (list) => {
    const ids = new Set(list.map((s) => s.ID));
    if (bulkSavedOrder.value.some((id) => !ids.has(id))) {
        bulkSavedOrder.value = bulkSavedOrder.value.filter((id) => ids.has(id));
    }
});

const bulkSavedMealsOrdered = computed((): SavedMeal[] => {
    const byId = new Map(savedMeals.value.map((s) => [s.ID, s]));
    return bulkSavedOrder.value
        .map((id) => byId.get(id))
        .filter((s): s is SavedMeal => s != null);
});

function toggleBulkSaved(id: number) {
    const cur = bulkSavedOrder.value;
    bulkSavedOrder.value = cur.includes(id)
        ? cur.filter((x) => x !== id)
        : [...cur, id];
}

function moveBulkSaved(index: number, dir: -1 | 1) {
    const j = index + dir;
    const cur = bulkSavedOrder.value;
    if (j < 0 || j >= cur.length) return;
    const next = [...cur];
    const t = next[index];
    next[index] = next[j];
    next[j] = t;
    bulkSavedOrder.value = next;
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
    const mealIds = [...bulkSavedOrder.value];
    if (mealIds.length === 0) {
        toast.push("Select at least one saved meal", "error");
        return;
    }
    bulkSaving.value = true;
    try {
        // Days are independent, so run them in parallel; meals within a day
        // stay sequential to preserve the chosen order.
        const results = await Promise.allSettled(
            offsets.map(async (off) => {
                for (const mealId of mealIds) {
                    await addPlannedMealFromSaved(mealId, off);
                }
            }),
        );
        const failedDays = results.filter((r) => r.status === "rejected").length;
        if (failedDays === 0) {
            toast.push(
                `Added ${mealIds.length} meal(s) across ${offsets.length} day(s)`,
                "success",
            );
        } else {
            toast.push(
                `Could not apply range (${failedDays} of ${offsets.length} day(s) failed)`,
                "error",
            );
        }
    } finally {
        await queryClient.invalidateQueries({ queryKey: dietKeys.days.all });
        bulkSaving.value = false;
    }
}
</script>

<template>
    <Section title="Plan a date range">
        <Card>
            <div class="flex flex-col gap-4">
                <p class="m-0 text-sm text-fg-muted">
                    Pick dates and saved meals. The order below is first to last
                    on each day.
                </p>
                <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
                    <Field label="Start">
                        <Input v-model="bulkFrom" type="date" />
                    </Field>
                    <Field label="End">
                        <Input v-model="bulkTo" type="date" />
                    </Field>
                </div>
                <div class="flex flex-col gap-2">
                    <span class="text-xs font-medium text-fg-muted"
                        >Saved meals to add on each day</span
                    >
                    <ErrorState
                        v-if="savedError"
                        compact
                        title="Couldn't load saved meals"
                        :error="savedError"
                        @retry="refetchSaved()"
                    />
                    <SkeletonList
                        v-else-if="showSkeleton"
                        :rows="3"
                        :leading="false"
                    />
                    <EmptyState
                        v-else-if="!savedPending && !savedMeals.length"
                        title="No saved meals"
                    />
                    <ul
                        v-else
                        class="m-0 flex list-none flex-col divide-y divide-line rounded-card border border-line p-0"
                    >
                        <li
                            v-for="sm in savedMeals"
                            :key="'bulk-' + sm.ID"
                            class="px-3 lg:px-4"
                        >
                            <Checkbox
                                :model-value="bulkSavedOrder.includes(sm.ID)"
                                :label="sm.name"
                                @update:model-value="toggleBulkSaved(sm.ID)"
                            >
                                <span class="flex flex-col gap-0.5 py-1">
                                    <span class="font-medium">{{ sm.name }}</span>
                                    <SimpleMacros
                                        v-if="sm.items?.length"
                                        decimals="none"
                                        v-bind="macroTotals(sm.items)"
                                    />
                                </span>
                            </Checkbox>
                        </li>
                    </ul>
                </div>
                <div v-if="bulkSavedOrder.length > 0" class="flex flex-col gap-2">
                    <span class="text-xs font-medium text-fg-muted"
                        >Order on each day</span
                    >
                    <List class="rounded-card border border-line">
                        <ListRow
                            v-for="(sm, index) in bulkSavedMealsOrdered"
                            :key="'bulk-order-' + sm.ID"
                        >
                            {{ sm.name }}
                            <template #end>
                                <IconButton
                                    label="Move up in bulk order"
                                    size="sm"
                                    :disabled="index === 0 || bulkSaving"
                                    @click="moveBulkSaved(index, -1)"
                                >
                                    <ChevronUp class="size-5" />
                                </IconButton>
                                <IconButton
                                    label="Move down in bulk order"
                                    size="sm"
                                    :disabled="
                                        index === bulkSavedMealsOrdered.length - 1 ||
                                        bulkSaving
                                    "
                                    @click="moveBulkSaved(index, 1)"
                                >
                                    <ChevronDown class="size-5" />
                                </IconButton>
                            </template>
                        </ListRow>
                    </List>
                </div>
                <div>
                    <Button
                        variant="primary"
                        :loading="bulkSaving"
                        :disabled="savedPending || savedMeals.length === 0"
                        @click="saveBulkRange"
                    >
                        Save to range
                    </Button>
                </div>
            </div>
        </Card>
    </Section>
</template>
