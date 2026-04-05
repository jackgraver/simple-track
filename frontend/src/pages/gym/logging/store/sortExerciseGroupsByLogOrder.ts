import type { LoggedExercise, LoggedSet } from "~/types/workout";
import type { ExerciseGroup } from "./useWorkoutStore";

function parseTimeMs(obj: unknown): number | null {
    if (!obj || typeof obj !== "object") return null;
    const o = obj as Record<string, unknown>;
    const raw = o.created_at ?? o.CreatedAt;
    if (typeof raw !== "string" || !raw) return null;
    const ms = Date.parse(raw);
    return Number.isNaN(ms) ? null : ms;
}

function earliestSetTimeMs(sets: LoggedSet[] | undefined): number | null {
    if (!sets?.length) return null;
    let min: number | null = null;
    for (const s of sets) {
        const t = parseTimeMs(s);
        if (t != null && (min === null || t < min)) min = t;
    }
    return min;
}

/** First submit/log activity for this exercise in today's session (logged sets, else logged row time). */
function logOrderTimeMs(group: ExerciseGroup): number | null {
    const logged = group.logged as LoggedExercise | undefined;
    if (!logged) return null;
    const fromSets = earliestSetTimeMs(logged.sets);
    if (fromSets != null) return fromSets;
    return parseTimeMs(logged);
}

/**
 * Exercises with any logged activity first, ordered by earliest log/submit time;
 * exercises not yet touched keep their relative order at the end.
 */
export function sortExerciseGroupsByLogOrder(groups: ExerciseGroup[]): ExerciseGroup[] {
    const indexed = groups.map((g, i) => ({ g, i }));
    indexed.sort((a, b) => {
        const ta = logOrderTimeMs(a.g);
        const tb = logOrderTimeMs(b.g);
        const aHas = ta !== null;
        const bHas = tb !== null;
        if (aHas && bHas) {
            if (ta !== tb) return ta - tb;
            return a.i - b.i;
        }
        if (aHas && !bHas) return -1;
        if (!aHas && bHas) return 1;
        return a.i - b.i;
    });
    return indexed.map((x) => x.g);
}
