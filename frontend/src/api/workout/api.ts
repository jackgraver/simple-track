import type { WorkoutLog, Exercise, LoggedExercise } from '~/types/workout';

export type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

export type WorkoutLogsPreviousResponse = {
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
};
