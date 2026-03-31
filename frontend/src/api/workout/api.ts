import type {
    WorkoutLog,
    Exercise,
    LoggedExercise,
    PlannedCardio,
    Cardio,
} from "~/types/workout";

export type ExerciseGroup = {
    planned?: Exercise;
    logged?: LoggedExercise;
    previous?: LoggedExercise;
};

export type WorkoutLogsPreviousResponse = {
    day: WorkoutLog;
    planned_exercises: ExerciseGroup[];
    planned_cardio: PlannedCardio | null;
    logged_cardio: Cardio | null;
};
