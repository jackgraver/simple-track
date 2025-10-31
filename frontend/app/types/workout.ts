import type { BaseModel } from "./diet";

export interface PlannedSet extends BaseModel {
    planned_exercise_id: number;
    reps: number;
    weight: number;
}

export interface PlannedExercise extends BaseModel {
    workout_plan_id: number;
    name: string;
    sets: PlannedSet[];
    threshold: number;
}

export interface WorkoutPlan extends BaseModel {
    name: string;
    exercises: Exercise[];
}

export interface LoggedSet extends BaseModel {
    logged_exercise_id: number;
    reps: number;
    weight: number;
}

export interface LoggedExercise extends BaseModel {
    workout_log_id: number;
    exercise_id: number;
    exercise: Exercise;
    sets: LoggedSet[];
    weight_setup: string;
    percent_change?: number;
}

export interface Exercise extends BaseModel {
    name: string;
    rep_rollover: number;
}

export interface Cardio extends BaseModel {
    workout_log_id: number;
    minutes: number;
    type: string;
}

export interface WorkoutLog extends BaseModel {
    date: string; // ISO string for time.Time
    workout_plan_id?: number | null;
    workout_plan?: WorkoutPlan | null;
    exercises: LoggedExercise[];
    cardio?: Cardio | null;
}
