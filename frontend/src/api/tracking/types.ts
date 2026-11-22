export type ActivityLevel =
    | 'sedentary'
    | 'lightly_active'
    | 'moderately_active'
    | 'very_active'
    | 'extra_active';

export type UserProfile = {
    ID: number;
    height_in: number;
    age: number;
    sex: 'male' | 'female';
    activity_level: ActivityLevel;
};

export type BodyWeightLog = {
    ID: number;
    date: string;
    weight_lbs: number;
    /** Renpho body-composition metrics; null for manual entries. */
    measured_at?: string | null;
    bmi?: number | null;
    body_fat_pct?: number | null;
    body_fat_mass_lbs?: number | null;
    muscle_pct?: number | null;
    muscle_mass_lbs?: number | null;
    skeletal_muscle_pct?: number | null;
    skeletal_muscle_mass_lbs?: number | null;
    bone_pct?: number | null;
    bone_mass_lbs?: number | null;
    protein_pct?: number | null;
    protein_mass_lbs?: number | null;
    body_water_pct?: number | null;
    body_water_mass_lbs?: number | null;
    fat_free_mass_lbs?: number | null;
    subcutaneous_fat_pct?: number | null;
    visceral_fat?: number | null;
    bmr_kcal?: number | null;
    metabolic_age?: number | null;
    whr?: number | null;
    optimal_weight_lbs?: number | null;
    weight_level?: string | null;
    body_type?: string | null;
    target_optimal_weight_lbs?: number | null;
    target_optimal_muscle_mass_lbs?: number | null;
    target_optimal_fat_mass_lbs?: number | null;
    remarks?: string | null;
};

export type WeightImportResult = { created: number; updated: number };

export type StepLog = {
    ID: number;
    date: string;
    steps: number;
};

export type GroceryItem = {
    ID: number;
    name: string;
    completed_at: string | null;
    CreatedAt?: string;
    /** Set on rows created through the optimistic add; stable across the temp -> server swap. */
    client_id?: string;
};

export type ParsedGroceryLine = {
    input: string;
    name: string;
    matched: boolean;
    on_list: boolean;
};

export type DrinkSizePreset = {
    ID: number;
    name: string;
    amount_oz: number;
};

export type WaterLog = {
    ID: number;
    date: string;
    amount_oz: number;
    preset_id: number | null;
    preset?: DrinkSizePreset | null;
    CreatedAt?: string;
};
