<script setup lang="ts">
import { Plus } from "lucide-vue-next";
import { computed } from "vue";
import type { SavedMeal } from "~/types/diet";
import { withDietDayOffsetQuery } from "~/utils/dateUtil";
import { CREATE_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { formatInt } from "~/pages/diet/home/dietDay";
import {
    Button,
    EmptyState,
    ErrorState,
    List,
    ListRow,
    Overlay,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { kcalLine, savedMealMacros, shortDayLabel, type PlanDay } from "./domain/planWeek";

const open = defineModel<boolean>("open", { required: true });
const props = defineProps<{
    day: PlanDay | null;
    savedMeals: SavedMeal[];
    pending: boolean;
    error: unknown;
    /** Saved meal ids being added right now. */
    adding: Set<number>;
}>();
const emit = defineEmits<{ add: [sm: SavedMeal]; retry: [] }>();

const showSkeleton = useDelayedFlag(() => props.pending);

const rows = computed(() =>
    props.savedMeals.map((sm) => {
        const m = savedMealMacros(sm);
        return { sm, macros: m, after: (props.day?.planned.calories ?? 0) + m.calories };
    }),
);

const title = computed(() => (props.day ? `Add to ${shortDayLabel(props.day.date)}` : "Add meal"));
const description = computed(() =>
    props.day ? `${kcalLine(props.day.planned.calories, props.day.target.calories)} planned` : undefined,
);
const createTo = computed(() => ({
    name: "diet-log",
    query: withDietDayOffsetQuery(props.day?.offset ?? 0, { type: CREATE_TYPE }),
}));
</script>

<template>
    <Overlay v-model:open="open" :title="title" :description="description">
        <ErrorState v-if="error" compact title="Couldn't load saved meals" :error="error" @retry="emit('retry')" />
        <SkeletonList v-else-if="showSkeleton" :rows="3" :leading="false" />
        <EmptyState
            v-else-if="!pending && !savedMeals.length"
            title="No saved meals yet"
            description="Save a meal once, then add it to any day here."
        >
            <template #action>
                <Button :to="createTo">New saved meal</Button>
            </template>
        </EmptyState>
        <List v-else-if="savedMeals.length" class="-mx-4 lg:-mx-5">
            <ListRow
                v-for="r in rows"
                :key="r.sm.ID"
                :pending="adding.has(r.sm.ID)"
                @click="emit('add', r.sm)"
            >
                <template #leading><Plus class="size-4" /></template>
                {{ r.sm.name }}
                <template #subtitle>
                    {{ formatInt(r.macros.calories) }} kcal · P {{ formatInt(r.macros.protein) }} · C
                    {{ formatInt(r.macros.carbs) }} · F {{ formatInt(r.macros.fat) }}
                </template>
                <template v-if="day && day.target.calories > 0" #trailing>
                    <span
                        :class="r.after > day.target.calories * 1.05 ? 'text-warn' : ''"
                        :aria-label="`Day would be ${formatInt(r.after)} kcal`"
                        >→ {{ formatInt(r.after) }}</span
                    >
                </template>
            </ListRow>
        </List>
        <template #footer>
            <Button variant="primary" @click="open = false">Done</Button>
        </template>
    </Overlay>
</template>
