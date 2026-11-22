<script setup lang="ts">
import { Pencil } from "lucide-vue-next";
import { computed } from "vue";
import MealRow from "../home/MealRow.vue";
import { plannedRows } from "../home/dietDay";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";
import { EDIT_LOGGED_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { Button, Card, EmptyState, List, Section, SkeletonList, useDelayedFlag } from "~/shared/ui";
import { useDietLogsToday } from "~/api/diet/queries";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logPlannedMeal, logMeal } = useDietDayMealHandlers(() => props.dateOffset);
const { isPending } = useDietLogsToday(() => props.dateOffset);
const showSkeleton = useDelayedFlag(isPending);

const editRoute = computed(() => ({
    name: "diet-edit-planned",
    query: { offset: String(props.dateOffset) },
}));

// A logged planned meal leaves this list at once (the cache patch, like the server, moves it) and
// shows under Logged, pending until saved. The filter covers a day log that points at a still
// planned meal (older data).
const rows = computed(() =>
    plannedRows(data.value?.day.plannedMeals ?? [], data.value?.day.loggedMeals ?? []).filter((r) => !r.logged),
);
const anythingLogged = computed(() => (data.value?.day.loggedMeals?.length ?? 0) > 0);
const title = computed(() => (rows.value.length ? `Planned · ${rows.value.length} to go` : "Planned"));
</script>

<template>
    <Section :title="title">
        <template #actions>
            <Button variant="ghost" size="sm" :to="editRoute">
                <template #icon-left><Pencil class="size-4" /></template>
                Edit
            </Button>
        </template>
        <Card padding="none">
            <SkeletonList v-if="isPending" :rows="showSkeleton ? 3 : 0" :leading="true" />
            <EmptyState
                v-else-if="rows.length === 0 && anythingLogged"
                title="Nothing left to log"
                description="Planned meals move to Logged as you log them."
            >
                <template #action>
                    <Button variant="secondary" size="sm" :to="editRoute">Plan more</Button>
                </template>
            </EmptyState>
            <EmptyState
                v-else-if="rows.length === 0"
                title="Nothing planned"
                description="Plan meals for this day and tick them off here as you eat."
            >
                <template #action>
                    <Button variant="secondary" size="sm" :to="editRoute">Plan meals</Button>
                </template>
            </EmptyState>
            <List v-else>
                <MealRow
                    v-for="r in rows"
                    :key="r.planned.ID"
                    :meal="r.planned.meal"
                    kind="planned"
                    :can-log="dateOffset === 0"
                    :on-log="logPlannedMeal"
                    :on-edit="(m) => logMeal(m, EDIT_LOGGED_TYPE)"
                />
            </List>
        </Card>
    </Section>
</template>
