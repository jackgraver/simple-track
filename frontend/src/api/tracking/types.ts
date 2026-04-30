export type BodyWeightLog = {
    ID: number;
    date: string;
    weight_lbs: number;
};

export type StepLog = {
    ID: number;
    date: string;
    steps: number;
};

export type MissedTracking = {
    date: string;
    weight: boolean;
    steps: boolean;
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
};
