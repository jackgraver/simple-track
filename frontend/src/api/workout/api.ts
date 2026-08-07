import { apiGET, apiPATCH, apiPOST } from "~/api/client";
import type {
    WorkoutLog,
    Exercise,
    LoggedExercise,
    PlannedCardio,
    Cardio,
    MobilityRoutine,
    MobilityLogged,
    WorkoutPlan,
    WorkoutProgram,
} from "~/types/workout";

export async function getWorkoutPrograms(): Promise<{ programs: WorkoutProgram[] }> {
    return apiGET<{ programs: WorkoutProgram[] }>("/workout/programs");
}

export async function createWorkoutProgram(name: string): Promise<{ program: WorkoutProgram }> {
    return apiPOST<{ program: WorkoutProgram }>("/workout/programs", { name });
}

export async function renameWorkoutProgram(
    id: number,
    name: string,
): Promise<{ program: WorkoutProgram }> {
    return apiPATCH<{ program: WorkoutProgram }>(`/workout/programs/${id}`, { name });
}

export async function activateWorkoutProgram(
    id: number,
): Promise<{ program: WorkoutProgram }> {
    return apiPOST<{ program: WorkoutProgram }>(`/workout/programs/${id}/activate`);
}

export async function createWorkoutPlan(
    programId: number,
    name: string,
    dayOfWeek: number | null,
): Promise<{ plan: WorkoutPlan }> {
    return apiPOST<{ plan: WorkoutPlan }>(`/workout/programs/${programId}/plans`, {
        name,
        ...(dayOfWeek === null ? {} : { day_of_week: dayOfWeek }),
    });
}

export type ExerciseGroup = {
    planned?: Exercise;
    logged?: LoggedExercise;
    previous?: LoggedExercise;
    max?: LoggedExercise;
};

export type WorkoutActivityRange = {
    start: string;
    end: string;
};

export type WorkoutActivityResponse = {
    active_dates: string[];
    range: WorkoutActivityRange;
    mode: string;
};

export type WorkoutLogsPreviousResponse = {
    day: WorkoutLog;
    planned_exercises: ExerciseGroup[];
    planned_cardio: PlannedCardio | null;
    logged_cardio: Cardio | null;
    planned_pre_mobility: MobilityRoutine | null;
    logged_pre_mobility: MobilityLogged | null;
    planned_post_mobility: MobilityRoutine | null;
    logged_post_mobility: MobilityLogged | null;
};

export async function switchWorkoutPlan(
    offset: number,
    planId: number | null,
): Promise<WorkoutLogsPreviousResponse> {
    return apiPATCH<WorkoutLogsPreviousResponse>(
        "/workout/logs/switch-plan",
        { plan_id: planId },
        { params: { offset } },
    );
}
