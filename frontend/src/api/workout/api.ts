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

export async function getWorkoutLogsPrevious(
    offset: number = 0
): Promise<WorkoutLogsPreviousResponse> {
    const response = await apiClient.get<WorkoutLogsPreviousResponse>(
        '/workout/logs/previous',
        {
            params: { offset },
        }
    );
    return response.data;
}

export async function getWorkoutLogsToday(): Promise<WorkoutLog> {
    const response = await apiClient.get<WorkoutLog>('/workout/logs/today');
    return response.data;
}

export async function logExercise(
    exercise: LoggedExercise,
    type: 'logged' | 'previous'
): Promise<LoggedExercise> {
    const response = await apiClient.post<{ exercise: LoggedExercise }>(
        '/workout/exercises/log',
        {
            exercise,
            type,
        }
    );
    return response.data.exercise;
}

export async function addExerciseToWorkout(
    exerciseId: number,
    offset: number = 0
): Promise<LoggedExercise> {
    const response = await apiClient.post<{ exercise: LoggedExercise }>(
        '/workout/exercises/add',
        {
            exercise_id: exerciseId,
        },
        {
            params: { offset },
        }
    );
    return response.data.exercise;
}

export async function removeExerciseFromWorkout(
    exerciseId: number,
    offset: number = 0
): Promise<void> {
    await apiClient.delete('/workout/exercises/remove', {
        data: {
            exercise_id: exerciseId,
        },
        params: {
            offset,
        },
    });
}

export async function deleteLoggedSet(setId: number): Promise<void> {
    await apiClient.delete(`/workout/exercises/sets/${setId}`);
}

export async function getAllExercises(): Promise<Exercise[]> {
    const response = await apiClient.get<{ exercises: Exercise[] }>(
        '/workout/exercises/all'
    );
    return response.data.exercises;
}
