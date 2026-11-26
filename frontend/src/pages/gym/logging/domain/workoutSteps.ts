import type { ExerciseGroup } from "~/api/workout/api";
import { formatPair, formatWeight, hasSides, pairSets, type SidedSet } from "../exercise/domain/unilateral";
import type { Cardio, MobilityLogged, MobilityRoutine, PlannedCardio } from "~/types/workout";

/** One thing to do in a gym session, in the order it's done: warmup, lifts, cardio, cooldown. */
export type WorkoutStep = {
    key: string;
    kind: "warmup" | "lift" | "cardio" | "cooldown";
    name: string;
    done: boolean;
    /** Short context line: what was logged, or what last time looked like. */
    detail: string | null;
    /** Index into the sorted exercise log (lifts only), for navigation and removal. */
    exerciseIndex?: number;
};

export type WorkoutStepsInput = {
    log: ExerciseGroup[];
    plannedCardio?: PlannedCardio | null;
    loggedCardio?: Cardio | null;
    plannedPreMobility?: MobilityRoutine | null;
    loggedPreMobility?: MobilityLogged | null;
    plannedPostMobility?: MobilityRoutine | null;
    loggedPostMobility?: MobilityLogged | null;
};

type SetLike = SidedSet;

/**
 * "135 lb × 8, 8, 7 · 140 lb × 6": consecutive sets at the same weight share one weight. Weight 0 is
 * bodyweight. Unilateral sets read per pair: "30 lb × L10 R9, L10 R10", and a pair with a different
 * weight on each side spells both out ("L 30 lb × 10 R 25 lb × 9").
 */
export function formatSets(sets: SetLike[] | null | undefined): string {
    if (!sets?.length) return "";
    const items = hasSides(sets)
        ? pairSets(sets).map(formatPair)
        : sets.map((s) => ({ weight: s.weight as number | null, text: String(s.reps) }));
    const groups: { weight: number | null; reps: string[] }[] = [];
    for (const item of items) {
        const last = groups[groups.length - 1];
        if (last && item.weight != null && last.weight === item.weight) last.reps.push(item.text);
        else groups.push({ weight: item.weight, reps: [item.text] });
    }
    return groups
        .map((g) => (g.weight == null ? g.reps.join(", ") : `${formatWeight(g.weight)} × ${g.reps.join(", ")}`))
        .join(" · ");
}

function mobilityStep(
    kind: "warmup" | "cooldown",
    logged: MobilityLogged | null | undefined,
    planned: MobilityRoutine | null | undefined,
): WorkoutStep | null {
    const items = logged?.items.length ? logged.items : (planned?.items ?? []);
    if (!items.length) return null;
    const checked = logged?.checked.filter((name) => items.includes(name)).length ?? 0;
    const fallback = kind === "warmup" ? "Dynamic warmup" : "Static stretching";
    return {
        key: kind,
        kind,
        name: logged?.title || planned?.title || fallback,
        done: checked === items.length,
        detail: checked > 0 && checked < items.length ? `${checked} of ${items.length} done` : `${items.length} items`,
    };
}

export function buildWorkoutSteps(input: WorkoutStepsInput): WorkoutStep[] {
    const steps: WorkoutStep[] = [];

    const warmup = mobilityStep("warmup", input.loggedPreMobility, input.plannedPreMobility);
    if (warmup) steps.push(warmup);

    input.log.forEach((g, index) => {
        const sets = g.logged?.sets ?? [];
        const done = sets.length > 0;
        const previous = formatSets(g.previous?.sets);
        steps.push({
            key: `lift-${g.planned?.ID ?? g.logged?.exercise_id ?? g.logged?.ID ?? index}`,
            kind: "lift",
            name: g.planned?.name || g.logged?.exercise?.name || "Exercise",
            done,
            detail: done ? formatSets(sets) : previous ? `Last ${previous}` : null,
            exerciseIndex: index,
        });
    });

    if (input.plannedCardio || input.loggedCardio) {
        const minutes = input.loggedCardio?.minutes ?? 0;
        const type = input.plannedCardio?.type || input.loggedCardio?.type || "Cardio";
        steps.push({
            key: "cardio",
            kind: "cardio",
            name: type,
            done: minutes > 0,
            detail: minutes > 0 ? `${minutes} min` : input.plannedCardio ? `${input.plannedCardio.minutes} min planned` : null,
        });
    }

    const cooldown = mobilityStep("cooldown", input.loggedPostMobility, input.plannedPostMobility);
    if (cooldown) steps.push(cooldown);

    return steps;
}

/** The first step that isn't done, or -1 when everything is. */
export function nextStepIndex(steps: WorkoutStep[]): number {
    return steps.findIndex((s) => !s.done);
}

export const stepKindLabel: Record<WorkoutStep["kind"], string> = {
    warmup: "Warmup",
    lift: "Lift",
    cardio: "Cardio",
    cooldown: "Cooldown",
};
