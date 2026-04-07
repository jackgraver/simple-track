import { apiPATCH } from "~/api/client";
import type {
    WorkoutLog,
    Exercise,
    LoggedExercise,
    PlannedCardio,
    Cardio,
    MobilityRoutine,
    MobilityLogged,
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
