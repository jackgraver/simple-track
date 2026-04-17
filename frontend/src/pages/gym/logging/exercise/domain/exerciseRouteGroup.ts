import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { ExerciseGroup } from "../../store/useWorkoutStore";

export function parseExerciseIdParam(
    route: Pick<RouteLocationNormalizedLoaded, "params">,
): number | null {
    const raw = route.params.id;
    const str = Array.isArray(raw) ? raw[0] : raw;
    const n = Number.parseInt(String(str ?? ""), 10);
    return Number.isFinite(n) ? n : null;
}

export function findExerciseGroupByExerciseId(
    log: readonly ExerciseGroup[],
    exerciseId: number,
): ExerciseGroup | null {
    return (
        log.find(
            (eg) =>
                eg.planned?.ID === exerciseId ||
                eg.logged?.exercise_id === exerciseId,
        ) ?? null
    );
}
