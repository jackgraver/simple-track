import { apiClient } from '~/utils/axios';
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

export async function getWorkoutLogsPrevious(offset: number): Promise<WorkoutLogsPreviousResponse> {
    const response = await apiClient.get<WorkoutLogsPreviousResponse>(
        '/workout/logs/previous',
        {
            params: { offset },
        }
    );
    return response.data;
}

