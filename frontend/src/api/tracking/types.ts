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
