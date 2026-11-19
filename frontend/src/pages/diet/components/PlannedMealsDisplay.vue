<script setup lang="ts">
import { Pencil } from "lucide-vue-next";
import { computed } from "vue";
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";
import {
    Button,
    EmptyState,
    Section,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { useDietLogsToday } from "~/api/diet/queries";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal } =
    useDietDayMealHandlers(() => props.dateOffset);
const { isPending } = useDietLogsToday(() => props.dateOffset);
const showSkeleton = useDelayedFlag(isPending);

const editRoute = computed(() => ({
    name: "diet-edit-planned",
    query: { offset: String(props.dateOffset) },
}));

const sortedPlannedMeals = computed(() => {
    const list = data.value?.day.plannedMeals ?? [];
    return [...list].sort(
        (a, b) =>
            (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID,
    );
});
</script>

<template>
    <Section :title="`Planned (${sortedPlannedMeals.length})`">
        <template #actions>
            <Button variant="ghost" size="sm" :to="editRoute">
                <template #icon-left><Pencil class="size-4" /></template>
                Edit
            </Button>
        </template>
        <SkeletonList v-if="showSkeleton" :rows="2" :leading="false" />
        <template v-else-if="data">
            <EmptyState
                v-if="sortedPlannedMeals.length === 0"
                title="Nothing planned"
                description="Plan meals for this day with Edit."
            />
            <MealCard
                v-for="log in sortedPlannedMeals"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                type="planned"
            />
        </template>
    </Section>
</template>
