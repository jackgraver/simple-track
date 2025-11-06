<script setup lang="ts">
import type { Exercise, LoggedExercise, WorkoutLog } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
}>(`workout/previous?offset=0`);

const log = ref<ExerciseGroup[]>(
    // data?.value?.previous_exercises.map((e) => e.previous) ?? [],
    data.value?.previous_exercises ?? [],
);

const addSet = (exercise: LoggedExercise) => {
    if (!exercise) return;

    const weight = exercise?.sets[0]?.weight ?? 0;
    const reps = exercise?.sets[0]?.reps ?? 0;

    exercise.sets.push({
        logged_exercise_id: 0,
        reps: reps,
        weight: weight,
        ID: 0,
        created_at: "",
        updated_at: "",
    });
};

const removeSet = (
    exercise: LoggedExercise,
    set: LoggedExercise["sets"][0],
) => {
    if (!exercise) return;

    const index = exercise.sets.indexOf(set);
    exercise.sets.splice(index, 1);
};

const logExercise = async (
    exercise: LoggedExercise,
    type: "logged" | "previous",
): Promise<boolean> => {
    console.log("logging", exercise, type);
    const rawExercise = toRaw(exercise);
    rawExercise.sets = toRaw(rawExercise.sets).filter(
        (set) => !(set.reps === 0 && set.weight === 0),
    );
    rawExercise.workout_log_id =
        data.value?.day.ID ?? rawExercise.workout_log_id;

    const { response, error } = await useAPIPost<{
        exercise: LoggedExercise;
    }>(
        `workout/exercise/log`,
        "POST",
        {
            exercise: rawExercise,
            type: type,
        },
        {},
        false,
    );

    if (error) {
        console.error(error);
        toast.push(error.message, "error");
        return false;
    }

    return true;
};

const confirmLogs = async () => {
    const { response, error } = await useAPIPost<{
        all_logged: boolean;
    }>(`workout/exercise/all-logged`, "POST", {});

    if (error) {
        console.error(error);
        toast.push(error.message, "error");
        return;
    }
    if (response) {
        if (response.all_logged) {
            toast.push("All logged!", "success");
        } else {
            toast.push("Not all logged!", "error");
        }
    }
};
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div class="card-container">
            <LiveExerciseCard
                v-for="log in log"
                :data="log"
                :add-set="addSet"
                :log-exercise="logExercise"
                :remove-set="removeSet"
            />
        </div>

        <button @click="confirmLogs"><span>Finish</span></button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.card-container {
    display: flex;
    gap: 1rem;
}

@media (max-width: 767px) {
    .card-container {
        flex-direction: column;
    }
}

@media (min-width: 768px) {
    .card-container {
        flex-direction: row;
        flex-wrap: wrap;
    }
}

.container header {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    align-items: center;
    padding-bottom: 1rem;
}

.container header h1 {
    flex: 1;
}
</style>
