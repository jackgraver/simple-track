<script setup lang="ts">
import type { Exercise, LoggedExercise } from "~/types/workout";
import { ref, computed, watch } from "vue";
import { Trash2 } from "lucide-vue-next";
import { useAPIGet } from "~/composables/useApiFetch";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const props = defineProps<{
    exercises: ExerciseGroup[];
}>();

const emit = defineEmits<{
    (e: "select-exercise", index: number): void;
    (e: "finish-workout"): void;
    (e: "add-exercise", exerciseId: number): void;
    (e: "remove-exercise", index: number): void;
}>();

// Autocomplete state
const searchQuery = ref("");
const showSuggestions = ref(false);
const allExercises = ref<Exercise[]>([]);
const filteredExercises = computed(() => {
    const query = searchQuery.value.toLowerCase();
    return allExercises.value.filter(ex => 
        ex.name.toLowerCase().includes(query) &&
        !props.exercises.some(eg => eg.planned?.ID === ex.ID)
    ).slice(0, 10);
});

// Load all exercises on mount
const { data: exercisesData } = useAPIGet<{ exercises: Exercise[] }>("workout/exercises/all");
watch(exercisesData, (newData) => {
    if (newData?.exercises) {
        allExercises.value = newData.exercises;
    }
}, { immediate: true });

// Handle exercise selection
const selectExercise = (exercise: Exercise) => {
    emit("add-exercise", exercise.ID);
    searchQuery.value = "";
    showSuggestions.value = false;
};

// Handle blur with delay to allow clicking suggestions
const handleBlur = () => {
    setTimeout(() => {
        showSuggestions.value = false;
    }, 200);
};

// Handle remove exercise
const removeExercise = (index: number, event: Event) => {
    event.stopPropagation();
    emit("remove-exercise", index);
};

// Get maximum weight from previous exercise
const getMaxWeight = (exerciseGroup: ExerciseGroup): number | null => {
    if (!exerciseGroup.previous || !exerciseGroup.previous.sets || exerciseGroup.previous.sets.length === 0) {
        return null;
    }
    return Math.max(...exerciseGroup.previous.sets.map(set => set.weight));
};

// Get last set from logged exercise
const getLastSet = (exerciseGroup: ExerciseGroup): { weight: number; reps: number } | null => {
    if (!exerciseGroup.logged || !exerciseGroup.logged.sets || exerciseGroup.logged.sets.length === 0) {
        return null;
    }
    const lastSet = exerciseGroup.logged.sets[exerciseGroup.logged.sets.length - 1];
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
    return !!exerciseGroup.logged && exerciseGroup.logged.sets && exerciseGroup.logged.sets.length > 0;
};
</script>

<template>
    <div class="list-view">
        <h2>Exercises</h2>
        <div class="add-exercise-container">
            <div class="autocomplete-wrapper">
                <input
                    v-model="searchQuery"
                    @focus="showSuggestions = true"
                    @blur="handleBlur"
                    type="text"
                    placeholder="Add exercise..."
                    class="exercise-search-input"
                />
                <ul v-if="showSuggestions && filteredExercises.length > 0" class="suggestions-list">
                    <li
                        v-for="exercise in filteredExercises"
                        :key="exercise.ID"
                        @mousedown.prevent="selectExercise(exercise)"
                        class="suggestion-item"
                    >
                        {{ exercise.name }}
                    </li>
                </ul>
            </div>
        </div>
        <ul class="exercise-list">
            <li
                v-for="(exerciseGroup, index) in exercises"
                :key="index"
                @click="emit('select-exercise', index)"
                :class="['exercise-item', { 'logged': isLogged(exerciseGroup) }]"
            >
                <div class="exercise-content">
                    <div class="exercise-title-section">
                        <span class="exercise-name">{{ exerciseGroup.planned?.name || exerciseGroup.logged?.exercise?.name }}</span>
                        <span v-if="isLogged(exerciseGroup) && getLastSet(exerciseGroup)" class="exercise-subtitle">
                              <p v-for="set in exerciseGroup.logged?.sets">{{ set.weight }}lbs × {{ set.reps }}</p>
                        </span>
                        <span v-else-if="!isLogged(exerciseGroup) && getMaxWeight(exerciseGroup) !== null" class="exercise-subtitle">
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
        </ul>
        <button @click="emit('finish-workout')" class="finish-button">
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

.list-view h2 {
    margin: 0 0 1rem 0;
}

.add-exercise-container {
    margin-bottom: 0.5rem;
}

.autocomplete-wrapper {
    position: relative;
    width: 100%;
}

.exercise-search-input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    box-sizing: border-box;
}

.exercise-search-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}

.suggestions-list {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    margin-top: 0.25rem;
    list-style: none;
    padding: 0;
    background: rgb(27, 27, 27);
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    max-height: 200px;
    overflow-y: auto;
    z-index: 1000;
}

.suggestion-item {
    padding: 0.75rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.suggestion-item:hover {
    background: rgb(40, 40, 40);
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
    transition: background-color 0.2s, opacity 0.2s;
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
    transition: background-color 0.2s, border-color 0.2s;
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

