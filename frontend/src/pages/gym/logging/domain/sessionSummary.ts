import type { ExerciseGroup } from "~/api/workout/api";
import { setsForProgression } from "../exercise/domain/unilateral";

type SetLike = { weight: number; reps: number };

export type LiftGain = {
    key: string;
    name: string;
    /** Heaviest set today. */
    top: SetLike;
    /** Top-set weight change vs the previous session of this lift (null when there's no previous). */
    deltaWeight: number | null;
    /** Rep change at the same top weight (only meaningful when deltaWeight is 0). */
    deltaReps: number | null;
    /** Beat the best previous top set: heavier, or the same weight for more reps. */
    isPR: boolean;
};

export type SessionSummary = {
    lifts: LiftGain[];
    sets: number;
    reps: number;
    volume: number; // lb moved: sum of weight × reps
    /** Volume change vs the previous session, over the lifts that have one. */
    volumeDeltaPct: number | null;
    prCount: number;
};

const volumeOf = (sets: SetLike[]) => sets.reduce((sum, s) => sum + s.weight * s.reps, 0);

/** The heaviest set; ties go to the one with more reps. */
export function topSet(sets: SetLike[] | null | undefined): SetLike | null {
    if (!sets?.length) return null;
    return sets.reduce((best, s) =>
        s.weight > best.weight || (s.weight === best.weight && s.reps > best.reps) ? s : best,
    );
}

const beats = (a: SetLike, b: SetLike) => a.weight > b.weight || (a.weight === b.weight && a.reps > b.reps);

/** What today's logged lifts add up to, compared with last time (`previous`) and the best so far (`max`). */
export function buildSessionSummary(log: ExerciseGroup[]): SessionSummary {
    const lifts: LiftGain[] = [];
    let sets = 0;
    let reps = 0;
    let volume = 0;
    let comparedNow = 0;
    let comparedPrev = 0;

    log.forEach((g, i) => {
        // A unilateral set number (left + right) counts as one set, at its weaker side, the same as
        // progression does and as a bilateral set from before the toggle counts. So sets, reps and
        // volume compare evenly with a bilateral session.
        const today = setsForProgression(g.logged?.sets);
        const top = topSet(today);
        if (!top) return;
        sets += today.length;
        reps += today.reduce((n, s) => n + s.reps, 0);
        volume += volumeOf(today);

        const prevSets = setsForProgression(g.previous?.sets);
        const prevTop = topSet(prevSets);
        if (prevTop) {
            comparedNow += volumeOf(today);
            comparedPrev += volumeOf(prevSets);
        }
        const best = topSet(g.max?.sets);
        lifts.push({
            key: String(g.planned?.ID ?? g.logged?.exercise_id ?? i),
            name: g.planned?.name || g.logged?.exercise?.name || "Exercise",
            top,
            deltaWeight: prevTop ? top.weight - prevTop.weight : null,
            deltaReps: prevTop && prevTop.weight === top.weight ? top.reps - prevTop.reps : null,
            isPR: best ? beats(top, best) : false,
        });
    });

    return {
        lifts,
        sets,
        reps,
        volume,
        volumeDeltaPct: comparedPrev > 0 ? Math.round(((comparedNow - comparedPrev) / comparedPrev) * 100) : null,
        prCount: lifts.filter((l) => l.isPR).length,
    };
}
