/**
 * Pure cache patches for the rest of gym logging (optimistic-ui G3): cardio, mobility and
 * add/remove exercise. They operate on the `workoutKeys.days.day(date)` entry
 * (`GET /workout/logs/previous`). Per-set patches live in `setsCache.ts`.
 */
import type { ExerciseGroup, WorkoutLogsPreviousResponse } from "~/api/workout/api";
import { groupMatches } from "~/api/workout/setsCache";
import type { Cardio, Exercise, LoggedExercise, MobilityLogged } from "~/types/workout";

type Day = WorkoutLogsPreviousResponse;

export type MobilitySlot = "pre" | "post";
export type MobilityVars = { date: string; slot: MobilitySlot; checked: string[] };
export type CardioVars = { date: string; minutes: number; type?: string; notes?: string };
export type AddExerciseVars = { date: string; exerciseId: number; exercise?: Exercise };
export type RemoveExerciseVars = { date: string; exerciseId: number };

const loggedKey = (slot: MobilitySlot) => (slot === "pre" ? "logged_pre_mobility" : "logged_post_mobility");
const plannedKey = (slot: MobilitySlot) => (slot === "pre" ? "planned_pre_mobility" : "planned_post_mobility");

/** Tick state of one mobility slot. Without a logged row yet, it starts from the planned routine. */
export function setMobilityChecked(day: Day, slot: MobilitySlot, checked: string[]): Day {
    const base: MobilityLogged | null = day[loggedKey(slot)] ?? (day[plannedKey(slot)] ? { ...day[plannedKey(slot)]!, checked: [] } : null);
    if (!base) return day;
    return { ...day, [loggedKey(slot)]: { ...base, checked } };
}

/** Rollback when siblings also patched the day: put back only this slot's logged value. */
export function restoreMobility(cur: Day, snapshot: Day, slot: MobilitySlot): Day {
    return { ...cur, [loggedKey(slot)]: snapshot[loggedKey(slot)] };
}

/** Server answer for a mobility upsert: the slot's logged view. */
export function applyMobilityResponse(day: Day, slot: MobilitySlot, mobility: MobilityLogged): Day {
    return { ...day, [loggedKey(slot)]: mobility };
}

/** Cardio upsert. Empty/omitted `type` keeps the existing (or planned) type, like the server. */
export function setCardio(day: Day, v: CardioVars): Day {
    const prev = day.logged_cardio;
    const at = new Date().toISOString();
    const next: Cardio = {
        ID: prev?.ID ?? 0,
        workout_log_id: prev?.workout_log_id ?? day.day?.ID ?? 0,
        minutes: v.minutes,
        type: v.type ? v.type : (prev?.type ?? day.planned_cardio?.type ?? ""),
        notes: v.notes !== undefined ? v.notes : (prev?.notes ?? ""),
        created_at: prev?.created_at ?? at,
        updated_at: prev?.updated_at ?? at,
    };
    return { ...day, logged_cardio: next };
}

export function restoreCardio(cur: Day, snapshot: Day): Day {
    return { ...cur, logged_cardio: snapshot.logged_cardio };
}

export function applyCardioResponse(day: Day, cardio: Cardio): Day {
    return { ...day, logged_cardio: cardio };
}

function emptyLoggedFor(day: Day, exercise: Exercise): LoggedExercise {
    const at = new Date().toISOString();
    return {
        ID: 0,
        workout_log_id: day.day?.ID ?? 0,
        exercise_id: exercise.ID,
        exercise,
        sets: [],
        notes: "",
        created_at: at,
        updated_at: at,
    };
}

/** Add an exercise to the day. No-op without the exercise (nothing to render) or when it is already there. */
export function addExerciseGroup(day: Day, v: AddExerciseVars): Day {
    const groups = day.planned_exercises ?? [];
    if (!v.exercise) return day;
    const existing = groups.find((g) => groupMatches(g, v.exerciseId));
    if (existing?.logged) return day;
    if (existing) {
        return {
            ...day,
            planned_exercises: groups.map((g) => (g === existing ? { ...g, logged: emptyLoggedFor(day, v.exercise!) } : g)),
        };
    }
    return { ...day, planned_exercises: [...groups, { logged: emptyLoggedFor(day, v.exercise) }] };
}

/** Server answer for add: the saved logged row replaces the temporary one (previous/max arrive on refetch). */
export function applyAddedExercise(day: Day, exerciseId: number, logged: LoggedExercise): Day {
    const groups = day.planned_exercises ?? [];
    const i = groups.findIndex((g) => groupMatches(g, exerciseId));
    if (i < 0) return { ...day, planned_exercises: [...groups, { logged }] };
    const planned_exercises = [...groups];
    planned_exercises[i] = { ...groups[i], logged };
    return { ...day, planned_exercises };
}

/**
 * Remove an exercise's log. A group that is also in the plan stays as planned (the server rebuilds it
 * that way); an ad-hoc one disappears. Groups with nothing logged are left alone.
 */
export function removeExerciseGroup(day: Day, exerciseId: number): Day {
    const groups = day.planned_exercises ?? [];
    const i = groups.findIndex((g) => groupMatches(g, exerciseId));
    if (i < 0 || !groups[i].logged) return day;
    const planned_exercises = [...groups];
    if (groups[i].planned) {
        const { logged: _drop, ...rest } = groups[i];
        planned_exercises[i] = rest;
    } else {
        planned_exercises.splice(i, 1);
    }
    return { ...day, planned_exercises };
}

/** Inverse of `removeExerciseGroup` from the snapshot: restore the logged row, or re-insert the group in place. */
export function restoreExerciseGroup(cur: Day, snapshot: Day, exerciseId: number): Day {
    const snapGroups = snapshot.planned_exercises ?? [];
    const si = snapGroups.findIndex((g) => groupMatches(g, exerciseId));
    if (si < 0) return cur;
    const groups = cur.planned_exercises ?? [];
    const i = groups.findIndex((g) => groupMatches(g, exerciseId));
    const planned_exercises = [...groups];
    if (i >= 0) {
        planned_exercises[i] = { ...groups[i], logged: snapGroups[si].logged };
    } else {
        planned_exercises.splice(Math.min(si, planned_exercises.length), 0, snapGroups[si] as ExerciseGroup);
    }
    return { ...cur, planned_exercises };
}
