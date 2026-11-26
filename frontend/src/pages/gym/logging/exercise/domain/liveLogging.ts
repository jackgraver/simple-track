import type { MachineSetting } from "~/types/workout";
import { sessionOutcome } from "./weightProgression";

type SetLike = { weight: number; reps: number };

/** Rest clock stages. It always counts up; red <1:00, orange <1:30, yellow <2:00, green after. */
export type RestStage = "idle" | "wait" | "near" | "soon" | "ready";

export const REST_NEAR_MS = 60_000;
export const REST_SOON_MS = 90_000;
export const REST_READY_MS = 120_000;

/** Unilateral, between sides: yellow from 0:15, green from 0:25. */
export const SWITCH_SOON_MS = 15_000;
export const SWITCH_READY_MS = 25_000;

/** `switching`: between the two sides of a unilateral set, red < 0:15, yellow < 0:25, green after. */
export function restStage(elapsedMs: number, active: boolean, switching = false): RestStage {
    if (!active) return "idle";
    if (switching) return elapsedMs >= SWITCH_READY_MS ? "ready" : elapsedMs >= SWITCH_SOON_MS ? "soon" : "wait";
    if (elapsedMs >= REST_READY_MS) return "ready";
    if (elapsedMs >= REST_SOON_MS) return "soon";
    if (elapsedMs >= REST_NEAR_MS) return "near";
    return "wait";
}

/** Plans don't store a set count yet: use last session's, else this. */
export const DEFAULT_PLANNED_SETS = 2;

export function plannedSetCount(previousSets: readonly SetLike[] | null | undefined): number {
    return previousSets?.length ? previousSets.length : DEFAULT_PLANNED_SETS;
}

/** Sets done against the plan: "2/4", or just "5" once extra sets pass it ("5/4" reads as an error). */
export function setsTally(done: number, planned: number): string {
    return done > planned ? String(done) : `${done}/${planned}`;
}

/** Today's weight against where last session left off (its final set), with the rollover tally that explains it. */
export type OverloadStatus =
    | { kind: "none" }
    | {
          kind: "up" | "same" | "down";
          /** Last session's final set weight: where today starts before any increase. */
          previousWeight: number;
          /** currentWeight - previousWeight. */
          delta: number;
          /** Sets last time that reached the rollover (null when no rollover is set). */
          hit: number | null;
          total: number;
          rollover: number | null;
          /** Every set last time was at one weight and hit the rollover, i.e. the increase was earned. */
          earned: boolean;
          /** Last time dropped weight partway: previousWeight is the dropped weight. */
          dropped: boolean;
          /** Last time's sets weren't all at one weight. */
          mixed: boolean;
      };

export function overloadStatus(
    previousSets: readonly SetLike[] | null | undefined,
    repRollover: number | null | undefined,
    currentWeight: number,
): OverloadStatus {
    const outcome = sessionOutcome(previousSets, repRollover);
    if (!outcome || !(currentWeight > 0)) return { kind: "none" };
    const delta = currentWeight - outcome.startWeight;
    return {
        kind: delta > 0 ? "up" : delta < 0 ? "down" : "same",
        previousWeight: outcome.startWeight,
        delta,
        hit: outcome.hit,
        total: outcome.total,
        rollover: outcome.rollover,
        earned: outcome.earned,
        dropped: outcome.dropped,
        mixed: !outcome.sameWeight,
    };
}

/** "+5", "−10" or "" for the weight badge. */
export function formatWeightDelta(delta: number): string {
    if (!delta) return "";
    return `${delta > 0 ? "+" : "−"}${Math.abs(delta)}`;
}

/** Live check of today's sets: where does the next session start, and does it go up? */
export type ProgressVerdict = {
    hit: number;
    total: number;
    rollover: number;
    /** Every set at one weight and every set hit the rollover. */
    onTrack: boolean;
    /** Today's final set weight: where the next session starts (before any increase). */
    nextWeight: number;
    dropped: boolean;
    mixed: boolean;
};

export function progressVerdict(
    loggedSets: readonly SetLike[],
    repRollover: number | null | undefined,
): ProgressVerdict | null {
    const outcome = sessionOutcome(loggedSets, repRollover);
    if (!outcome || outcome.rollover == null || outcome.hit == null) return null;
    return {
        hit: outcome.hit,
        total: outcome.total,
        rollover: outcome.rollover,
        onTrack: outcome.earned,
        nextWeight: outcome.startWeight,
        dropped: outcome.dropped,
        mixed: !outcome.sameWeight,
    };
}

/** The verdict line: live while logging, or `finished` on the "Up next" screen. */
export function verdictText(v: ProgressVerdict, finished = false): string {
    if (v.dropped) return `Dropped to ${v.nextWeight} lb · start at ${v.nextWeight} next time`;
    if (v.mixed) return `Mixed weights · start at ${v.nextWeight} next time`;
    const tally = `${v.hit}/${v.total} sets hit ${v.rollover}`;
    if (v.onTrack) return `${tally} · ${finished ? "goes up next time" : "on track to go up"}`;
    return `${tally} · ${finished ? "same weight next time" : `stay at ${v.nextWeight} next time`}`;
}

/** Only settings with a label; values may be blank (shown as a placeholder). */
export function visibleMachineSettings(settings: readonly MachineSetting[] | null | undefined): MachineSetting[] {
    return (settings ?? []).filter((s) => s && typeof s.label === "string" && s.label.trim() !== "");
}

/** One sentence on why today's weight is what it is. */
export function overloadReason(status: OverloadStatus): string {
    if (status.kind === "none") return "No history yet. Pick a starting weight.";
    const { previousWeight: prev, delta, hit, total, rollover } = status;
    if (status.kind === "up") {
        return status.earned
            ? `Last time ${total}/${total} sets hit ${rollover} at ${prev} lb, so +${delta} today.`
            : `+${delta} lb on last time (${prev} lb).`;
    }
    if (status.kind === "down") return `${formatWeightDelta(delta)} lb on last time (${prev} lb).`;
    const toGoUp = rollover == null ? "" : ` Keep every set at ${prev} lb and hit ${rollover} to go up.`;
    if (status.dropped) return `Dropped to ${prev} lb last time, so starting there.${toGoUp}`;
    if (status.mixed) return `Mixed weights last time, ending at ${prev} lb.${toGoUp}`;
    if (hit == null || rollover == null) return `Same as last time (${prev} lb).`;
    return status.earned
        ? `Last time ${hit}/${total} sets hit ${rollover} at ${prev} lb. You earned an increase today.`
        : `Last time ${hit}/${total} sets hit ${rollover}. Hit ${rollover} on every set to go up.`;
}
