<script setup lang="ts">
import type {
    Cardio,
    Exercise,
    LoggedExercise,
    PlannedCardio,
    MobilityLogged,
} from "~/types/workout";
import type { DropdownOption } from "~/shared/input/Dropdown.vue";
import Dropdown from "~/shared/input/Dropdown.vue";
import { ref, computed } from "vue";
import { Trash2 } from "lucide-vue-next";
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

const exerciseOptions = computed<DropdownOption[]>(() => {
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

const handleExerciseSelect = (option: DropdownOption) => {
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

// Get last set from logged exercise
const getLastSet = (
    exerciseGroup: ExerciseGroup,
): { weight: number; reps: number } | null => {
    if (
        !exerciseGroup.logged ||
        !exerciseGroup.logged.sets ||
        exerciseGroup.logged.sets.length === 0
    ) {
        return null;
    }
    const lastSet =
        exerciseGroup.logged.sets[exerciseGroup.logged.sets.length - 1];
    if (!lastSet) {
        return null;
    }
    return {
        weight: lastSet.weight,
        reps: lastSet.reps,
    };
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
    <div class="list-view">
        <header>
            <h1 v-if="workoutName">{{ workoutName }} Day</h1>
        </header>
        <div class="mb-2">
            <Dropdown
                :options="exerciseOptions"
                :on-select="handleExerciseSelect"
                :has-more="hasNextPage ?? false"
                :loading="isFetching"
                placeholder="Add exercise..."
                @load-more="handleLoadMore"
                @search="handleExerciseSearch"
            />
        </div>
        <ul class="exercise-list">
            <li
                v-if="showPreMobilityRow"
                key="pre-mobility"
                @click="emit('select-pre-mobility')"
                :class="['exercise-item', { logged: preMobilityComplete }]"
            >
                <div class="exercise-content">
                    <div class="exercise-title-section">
                        <span class="exercise-name">{{
                            loggedPreMobility?.title ?? "Pre-workout mobility"
                        }}</span>
                    </div>
                </div>
            </li>
            <li
                v-for="(exerciseGroup, index) in exercises"
                :key="
                    exerciseGroup.planned?.ID ??
                    exerciseGroup.logged?.exercise_id ??
                    exerciseGroup.logged?.ID ??
                    index
                "
                @click="emit('select-exercise', index)"
                :class="['exercise-item', { logged: isLogged(exerciseGroup) }]"
            >
                <div class="exercise-content">
                    <div class="exercise-title-section">
                        <span class="exercise-name">{{
                            exerciseGroup.planned?.name ||
                            exerciseGroup.logged?.exercise?.name
                        }}</span>
                        <span
                            v-if="
                                isLogged(exerciseGroup) &&
                                getLastSet(exerciseGroup)
                            "
                            class="exercise-subtitle"
                        >
                            <p v-for="set in exerciseGroup.logged?.sets">
                                {{ set.weight }}lbs × {{ set.reps }}
                            </p>
                        </span>
                        <span
                            v-else-if="
                                !isLogged(exerciseGroup) &&
                                getMaxWeight(exerciseGroup) !== null
                            "
                            class="exercise-subtitle"
                        >
                            Prev {{ getMaxWeight(exerciseGroup) }}lbs
                        </span>
                    </div>
                </div>
                <button
                    @click="removeExercise(index, $event)"
                    class="remove-button"
                    type="button"
                >
                    <Trash2 :size="18" />
                </button>
            </li>
            <li
                v-if="showCardioRow"
                key="cardio"
                @click="emit('select-cardio')"
                :class="['exercise-item', { logged: cardioIsLogged }]"
            >
                <div class="exercise-content">
                    <div class="exercise-title-section">
                        <span class="exercise-name">{{ cardioName }}</span>
                        <span v-if="cardioIsLogged" class="exercise-subtitle">
                            {{ loggedCardio?.minutes }} min
                        </span>
                    </div>
                </div>
            </li>
            <li
                v-if="showPostMobilityRow"
                key="post-mobility"
                @click="emit('select-post-mobility')"
                :class="['exercise-item', { logged: postMobilityComplete }]"
            >
                <div class="exercise-content">
                    <div class="exercise-title-section">
                        <span class="exercise-name">{{
                            loggedPostMobility?.title ?? "Post-workout mobility"
                        }}</span>
                    </div>
                </div>
            </li>
        </ul>
        <button @click="router.push('/')" type="button" class="finish-button">
            <span>Finish Workout</span>
        </button>
    </div>
</template>

<style scoped>
.list-view {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    max-width: 100%;
}

.exercise-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
}

.exercise-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    cursor: pointer;
    transition:
        background-color 0.2s,
        opacity 0.2s;
    width: 100%;
    box-sizing: border-box;
    gap: 1rem;
}

.exercise-item:hover {
    background: rgb(40, 40, 40);
}

.exercise-item.logged {
    opacity: 0.6;
    background: rgb(20, 20, 20);
    border: 1px solid rgb(19, 128, 42);
}

.exercise-item.logged:hover {
    background: rgb(30, 30, 30);
    opacity: 0.8;
}

.exercise-content {
    flex: 1;
    display: flex;
    align-items: center;
    min-width: 0;
}

.exercise-title-section {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
}

.exercise-name {
    font-weight: 500;
    font-size: 1.1rem;
}

.exercise-subtitle {
    color: rgb(150, 150, 150);
    font-size: 0.9rem;
    display: flex;
    flex-direction: row;
}

.exercise-subtitle p {
    margin: 0;
}

.exercise-subtitle p:not(:last-child)::after {
    content: "•";
    margin: 0 0.5rem;
    color: rgb(150, 150, 150);
}

.exercise-item.logged .exercise-subtitle {
    color: rgb(200, 200, 200);
    font-weight: 500;
}

.remove-button {
    width: 2rem;
    height: 2rem;
    border: 1px solid rgb(80, 40, 40);
    border-radius: 3px;
    background: rgb(40, 20, 20);
    color: rgb(200, 100, 100);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition:
        background-color 0.2s,
        border-color 0.2s;
    padding: 0;
    flex-shrink: 0;
}

.remove-button:hover {
    background: rgb(60, 30, 30);
    border-color: rgb(120, 60, 60);
}

.finish-button {
    margin-top: 1rem;
    padding: 0.75rem 1.5rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(80, 80, 40) !important;
    color: inherit;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.finish-button:hover {
    background: rgb(100, 100, 50) !important;
}
</style>
