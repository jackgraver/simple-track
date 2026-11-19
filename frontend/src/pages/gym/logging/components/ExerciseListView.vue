<script setup lang="ts">
import type {
    Cardio,
    Exercise,
    LoggedExercise,
    PlannedCardio,
    MobilityLogged,
} from "~/types/workout";
import type { ComboboxOption } from "~/shared/ui";
import { ref, computed } from "vue";
import { ChevronRight, Dumbbell, Trash2 } from "lucide-vue-next";
import { Badge, Button, Combobox, Card, EmptyState, IconButton, List, ListRow } from "~/shared/ui";
import { useRouter } from "vue-router";
import { useExercisesPaginated } from "~/api/workout/queries";

type ExerciseGroup = {
    planned?: Exercise;
    logged?: LoggedExercise;
    previous?: LoggedExercise;
};

const props = defineProps<{
    workoutName: string;
    exercises: ExerciseGroup[];
    plannedCardio?: PlannedCardio | null;
    loggedCardio?: Cardio | null;
    loggedPreMobility?: MobilityLogged | null;
    loggedPostMobility?: MobilityLogged | null;
    removingId?: number | null;
}>();

const workoutName = computed(() => props.workoutName);
const emit = defineEmits<{
    (e: "select-exercise", index: number): void;
    (e: "add-exercise", exerciseId: number): void;
    (e: "remove-exercise", index: number): void;
    (e: "select-cardio"): void;
    (e: "select-pre-mobility"): void;
    (e: "select-post-mobility"): void;
}>();

const showCardioRow = computed(
    () => props.plannedCardio != null || props.loggedCardio != null,
);

const cardioName = computed(
    () => props.plannedCardio?.type ?? props.loggedCardio?.type ?? "Cardio",
);

const cardioIsLogged = computed(() => (props.loggedCardio?.minutes ?? 0) > 0);

const showPreMobilityRow = computed(() => {
    const l = props.loggedPreMobility;
    return l != null && l.items.length > 0;
});

const showPostMobilityRow = computed(() => {
    const l = props.loggedPostMobility;
    return l != null && l.items.length > 0;
});

const preMobilityComplete = computed(() => {
    const l = props.loggedPreMobility;
    if (!l?.items.length) return false;
    return l.items.every((name) => l.checked.includes(name));
});

const postMobilityComplete = computed(() => {
    const l = props.loggedPostMobility;
    if (!l?.items.length) return false;
    return l.items.every((name) => l.checked.includes(name));
});

const router = useRouter();

const exerciseSearch = ref("");
const {
    data: exercisesPages,
    fetchNextPage,
    hasNextPage,
    isFetching,
} = useExercisesPaginated(exerciseSearch);

const exerciseOptions = computed<ComboboxOption[]>(() => {
    const all = exercisesPages.value?.pages.flatMap((p) => p.exercises) ?? [];
    const existingIds = new Set(
        props.exercises
            .map((eg) => eg.planned?.ID ?? eg.logged?.exercise_id)
            .filter((id): id is number => id != null),
    );
    return all
        .filter((ex) => !existingIds.has(ex.ID))
        .map((ex) => ({ label: ex.name, value: ex.ID }));
});

const handleExerciseSelect = (option: ComboboxOption) => {
    emit("add-exercise", option.value as number);
};

const handleExerciseSearch = (query: string) => {
    exerciseSearch.value = query;
};

const handleLoadMore = () => {
    if (hasNextPage.value) fetchNextPage();
};

// Handle remove exercise
const removeExercise = (index: number, event: Event) => {
    event.stopPropagation();
    emit("remove-exercise", index);
};

const isRemoving = (g: ExerciseGroup) =>
    props.removingId != null &&
    (g.logged?.exercise_id ?? g.planned?.ID) === props.removingId;

const setSummary = (g: ExerciseGroup) =>
    (g.logged?.sets ?? []).map((set) => `${set.weight}lbs × ${set.reps}`).join(" • ");

const isEmptyList = computed(
    () =>
        props.exercises.length === 0 &&
        !showPreMobilityRow.value &&
        !showPostMobilityRow.value &&
        !showCardioRow.value,
);

// Get maximum weight from previous exercise
const getMaxWeight = (exerciseGroup: ExerciseGroup): number | null => {
    if (
        !exerciseGroup.previous ||
        !exerciseGroup.previous.sets ||
        exerciseGroup.previous.sets.length === 0
    ) {
        return null;
    }
    return Math.max(...exerciseGroup.previous.sets.map((set) => set.weight));
};

// Check if exercise is logged
const isLogged = (exerciseGroup: ExerciseGroup): boolean => {
    return (
        !!exerciseGroup.logged &&
        exerciseGroup.logged.sets &&
        exerciseGroup.logged.sets.length > 0
    );
};
</script>

<template>
    <div class="flex w-full flex-col gap-4">
        <h1 v-if="workoutName" class="m-0 text-xl font-semibold tracking-tight text-fg">
            {{ workoutName }} Day
        </h1>
        <Combobox
            :options="exerciseOptions"
            :has-more="hasNextPage ?? false"
            :loading="isFetching"
            placeholder="Add exercise..."
            @select="handleExerciseSelect"
            @load-more="handleLoadMore"
            @search="handleExerciseSearch"
        />
        <Card padding="none">
            <List>
                <ListRow
                    v-if="showPreMobilityRow"
                    @click="emit('select-pre-mobility')"
                >
                    {{ loggedPreMobility?.title ?? "Dynamic warmup" }}
                    <template #trailing>
                        <Badge v-if="preMobilityComplete" tone="success">Done</Badge>
                        <ChevronRight v-else aria-hidden="true" class="size-4" />
                    </template>
                </ListRow>
                <ListRow
                    v-for="(exerciseGroup, index) in exercises"
                    :key="
                        exerciseGroup.planned?.ID ??
                        exerciseGroup.logged?.exercise_id ??
                        exerciseGroup.logged?.ID ??
                        index
                    "
                    :pending="isRemoving(exerciseGroup)"
                    @click="emit('select-exercise', index)"
                >
                    {{
                        exerciseGroup.planned?.name ||
                        exerciseGroup.logged?.exercise?.name
                    }}
                    <template #subtitle>
                        <template v-if="isLogged(exerciseGroup)">
                            {{ setSummary(exerciseGroup) }}
                        </template>
                        <template v-else-if="getMaxWeight(exerciseGroup) !== null">
                            Prev {{ getMaxWeight(exerciseGroup) }}lbs
                        </template>
                    </template>
                    <template #trailing>
                        <Badge v-if="isLogged(exerciseGroup)" tone="success">Done</Badge>
                    </template>
                    <template #end>
                        <IconButton
                            label="Remove exercise"
                            variant="danger"
                            :disabled="isRemoving(exerciseGroup)"
                            @click="removeExercise(index, $event)"
                        >
                            <Trash2 />
                        </IconButton>
                    </template>
                </ListRow>
                <ListRow v-if="showCardioRow" @click="emit('select-cardio')">
                    {{ cardioName }}
                    <template v-if="cardioIsLogged" #subtitle>{{ loggedCardio?.minutes }} min</template>
                    <template #trailing>
                        <Badge v-if="cardioIsLogged" tone="success">Done</Badge>
                        <ChevronRight v-else aria-hidden="true" class="size-4" />
                    </template>
                </ListRow>
                <ListRow
                    v-if="showPostMobilityRow"
                    @click="emit('select-post-mobility')"
                >
                    {{ loggedPostMobility?.title ?? "Static stretching" }}
                    <template #trailing>
                        <Badge v-if="postMobilityComplete" tone="success">Done</Badge>
                        <ChevronRight v-else aria-hidden="true" class="size-4" />
                    </template>
                </ListRow>
            </List>
            <EmptyState
                v-if="isEmptyList"
                :icon="Dumbbell"
                title="Nothing planned for today"
                description="Add an exercise above to start logging."
            />
        </Card>
        <Button variant="primary" size="lg" block @click="router.push('/')">
            Finish workout
        </Button>
    </div>
</template>
