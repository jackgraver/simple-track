<script setup lang="ts">
import { computed } from "vue";
import MealRow from "../home/MealRow.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";
import { EDIT_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { Card, EmptyState, List, Section, SkeletonList, useDelayedFlag } from "~/shared/ui";
import { useDietLogsToday } from "~/api/diet/queries";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logMeal, deleteLoggedMeal } = useDietDayMealHandlers(() => props.dateOffset);
const { isPending } = useDietLogsToday(() => props.dateOffset);
const showSkeleton = useDelayedFlag(isPending);

const loggedMeals = computed(() => data.value?.day.loggedMeals ?? []);
</script>

<template>
    <Section :title="loggedMeals.length ? `Logged · ${loggedMeals.length}` : 'Logged'">
        <Card padding="none">
            <SkeletonList v-if="isPending" :rows="showSkeleton ? 2 : 0" :leading="true" />
            <EmptyState
                v-else-if="loggedMeals.length === 0"
                title="Nothing logged yet"
                description="Log a planned meal above, or anything else with Log meal."
            />
            <List v-else>
                <MealRow
                    v-for="log in loggedMeals"
                    :key="log.ID"
                    :meal="log.meal"
                    kind="logged"
                    :on-edit="(m) => logMeal(m, EDIT_TYPE)"
                    :on-delete="deleteLoggedMeal"
                />
            </List>
        </Card>
    </Section>
</template>
