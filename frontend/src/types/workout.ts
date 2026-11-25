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
    workout_program_id?: number | null;
    day_of_week: number | null; // 0=Sunday, 1=Monday, ..., 6=Saturday, null=unassigned
    assigned_days?: number[];
    planned_cardio_type?: string;
    planned_cardio_minutes?: number;
    pre_mobility_items?: string[];
    post_mobility_items?: string[];
    exercises: Exercise[];
}

export interface WorkoutProgram extends BaseModel {
    name: string;
    is_active: boolean;
    plans: WorkoutPlan[];
}

export interface PlannedCardio {
    type: string;
    minutes: number;
}

export interface LoggedSet extends BaseModel {
    logged_exercise_id: number;
    reps: number;
    weight: number;
    /** Plate text (e.g. "45 + 2×10"), still written for plate-loaded sets; parsed for old sets. */
    weight_setup: string;
    /** Plates the set was lifted with (plate-loaded exercises). Null before migration 0008. */
    plate_setup?: PlateSetup | null;
    /** Client-generated uuid (POST /workout/sets). Null on sets logged before G1. */
    client_id?: string | null;
}

export type PlateMode = "per_side" | "total";

/** How a set was loaded: `plates` maps a plate's weight ("45", "2.5") to its count (per side in
 *  per_side mode). `base` is the bar/sled weight at the time. */
export interface PlateSetup {
    base: number;
    mode: PlateMode;
    plates: Record<string, number>;
}

export interface LoggedExercise extends BaseModel {
    workout_log_id: number;
    exercise_id: number;
    exercise: Exercise;
    sets: LoggedSet[];
    notes: string;
    percent_change?: number;
    log_date?: string;
}

export interface Exercise extends BaseModel {
    name: string;
    rep_rollover: number;
    cues: string;
    load_type?: ExerciseLoadType | LegacyExerciseLoadType;
    /** Bar/sled/starting weight; null means the load type's default (see domain/plates.ts). */
    base_weight?: number | null;
    /** Weight change per step; null means the load type's default. */
    weight_step?: number | null;
    /** Machine positions (seat, pad, pulley). Not editable yet; null/absent means none. */
    machine_settings?: MachineSetting[] | null;
}

export interface MachineSetting {
    label: string;
    value: string;
}

export type ExerciseLoadType =
    | "plate_loaded_per_side"
    | "plate_loaded_total"
    | "weight_stack"
    | "free_weights";

/** Load types from before migration 0008 (per side with base 45 / 0). Only seen in stale caches. */
export type LegacyExerciseLoadType = "plate_loaded_with_bar" | "plate_loaded_without_bar";

export const DEFAULT_EXERCISE_LOAD_TYPE: ExerciseLoadType =
    "plate_loaded_per_side";

export interface Cardio extends BaseModel {
    workout_log_id: number;
    minutes: number;
    type: string;
    notes: string;
}

export interface MobilityRoutine {
    title: string;
    items: string[];
}

export interface MobilityLogged extends MobilityRoutine {
    checked: string[];
}

export interface WorkoutLog extends BaseModel {
    date: string; // ISO string for time.Time
    workout_plan_id?: number | null;
    workout_plan?: WorkoutPlan | null;
    exercises: LoggedExercise[];
    cardio?: Cardio | null;
    pre_mobility_checked?: string[];
    post_mobility_checked?: string[];
}
