<script setup lang="ts">
import { computed } from "vue";
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";
import { EmptyState, Section, SkeletonList, useDelayedFlag } from "~/shared/ui";
import { useDietLogsToday } from "~/api/diet/queries";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal } =
    useDietDayMealHandlers(() => props.dateOffset);
const { isPending } = useDietLogsToday(() => props.dateOffset);
const showSkeleton = useDelayedFlag(isPending);

const loggedMeals = computed(() => data.value?.day.loggedMeals ?? []);
</script>

<template>
    <Section :title="`Logged (${loggedMeals.length})`">
        <SkeletonList v-if="showSkeleton" :rows="2" :leading="false" />
        <template v-else-if="data">
            <EmptyState
                v-if="loggedMeals.length === 0"
                title="Nothing logged yet"
                description="Log a planned meal, or use Log meal or Quick log."
            />
            <MealCard
                v-for="log in loggedMeals"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                type="logged"
            />
        </template>
    </Section>
</template>
