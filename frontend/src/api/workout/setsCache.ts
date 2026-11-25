/**
 * Pure cache patches for per-set gym logging (optimistic-ui G2). They operate on the
 * `workoutKeys.days.day(date)` entry (`GET /workout/logs/previous`) and are used by the set hooks in
 * `queries.ts`. Nothing here touches the network or the QueryClient.
 *
 * Sets are identified by a "key": their `client_id` when they have one (every set logged since G1),
 * else `id:<ID>` for older sets. The key is also the `client_id` field of the set mutations' variables,
 * so `usePendingIds` and the helper's `retry`/`dismiss` work for both kinds.
 */
import type { ExerciseGroup, WorkoutLogsPreviousResponse } from "~/api/workout/api";
import type { LoggedExercise, LoggedSet, PlateSetup } from "~/types/workout";

type Day = WorkoutLogsPreviousResponse;

/** `plate_setup`: the plates for a plate-loaded exercise; null/absent otherwise. */
export type SetFields = { reps: number; weight: number; weight_setup: string; plate_setup?: PlateSetup | null };

/** The editable fields of `v`, so every patch copies the same ones. */
export function setFieldsOf(v: SetFields): Required<SetFields> {
    return { reps: v.reps, weight: v.weight, weight_setup: v.weight_setup, plate_setup: v.plate_setup ?? null };
}

/** `POST /workout/sets`. `created_at` is fixed at tap time so re-applying the patch is idempotent. */
export type LogSetVars = SetFields & { client_id: string; date: string; exercise_id: number; created_at: string };
/** `PATCH /workout/sets/:ref`. `client_id` is the set's key (see `setKey`). */
export type UpdateSetVars = SetFields & { client_id: string; date: string; exercise_id: number };
/** `DELETE /workout/sets/:ref`. `client_id` is the set's key. */
export type DeleteSetVars = { client_id: string; date: string; exercise_id: number };
export type SaveNotesVars = { date: string; exercise_id: number; notes: string };

export type LogSetResponse = { exercise: LoggedExercise; set: LoggedSet; created: boolean; first_set_of_day: boolean };
export type UpdateSetResponse = { exercise: LoggedExercise; set: LoggedSet };
export type DeleteSetResponse = {
    success: boolean;
    deleted: boolean;
    logged_exercise_id: number;
    exercise_removed: boolean;
};

const ID_PREFIX = "id:";

/** Stable identity of a set across optimistic and saved states. */
export function setKey(set: Pick<LoggedSet, "ID" | "client_id">): string {
    return set.client_id ? set.client_id : `${ID_PREFIX}${set.ID}`;
}

/** The `:id` path segment for PATCH/DELETE `/workout/sets/:id` (numeric id or client_id). */
export function setRef(key: string): string {
    return key.startsWith(ID_PREFIX) ? key.slice(ID_PREFIX.length) : key;
}

/** Not yet saved on the server (optimistic row). */
export function isUnsavedSet(set: Pick<LoggedSet, "ID">): boolean {
    return !(set.ID > 0);
}

export function groupMatches(group: ExerciseGroup, exerciseId: number): boolean {
    return group.planned?.ID === exerciseId || group.logged?.exercise_id === exerciseId;
}

export function findGroup(day: Day | undefined, exerciseId: number): ExerciseGroup | undefined {
    return day?.planned_exercises?.find((g) => groupMatches(g, exerciseId));
}

/**
 * Replace the group for `exerciseId` with `fn(group)` (`null` removes it). Returns `day` itself when
 * no group matches or nothing changed, so `setQueryData` is a no-op then.
 */
export function withGroup(day: Day, exerciseId: number, fn: (g: ExerciseGroup) => ExerciseGroup | null): Day {
    const groups = day.planned_exercises ?? [];
    const index = groups.findIndex((g) => groupMatches(g, exerciseId));
    if (index < 0) return day;
    const next = fn(groups[index]);
    if (next === groups[index]) return day;
    const planned_exercises = [...groups];
    if (next === null) planned_exercises.splice(index, 1);
    else planned_exercises[index] = next;
    return { ...day, planned_exercises };
}

function emptyLogged(day: Day, group: ExerciseGroup, exerciseId: number, at: string): LoggedExercise {
    const exercise = group.planned ?? group.previous?.exercise ?? group.max?.exercise;
    return {
        ID: 0,
        workout_log_id: day.day?.ID ?? 0,
        exercise_id: exerciseId,
        exercise: exercise as LoggedExercise["exercise"],
        sets: [],
        notes: "",
        created_at: at,
        updated_at: at,
    };
}

/** A logged exercise left with no sets that only ever existed locally goes away with its last set. */
function dropIfEmptyLocal(group: ExerciseGroup, logged: LoggedExercise): ExerciseGroup {
    if (logged.sets.length === 0 && isUnsavedSet(logged)) {
        const { logged: _drop, ...rest } = group;
        return rest;
    }
    return { ...group, logged };
}

// ── Patches ────────────────────────────────────────────────────────────────────────────────────

/** Log set (apply). Upsert by client_id, so re-applying on Retry doesn't duplicate the row. */
export function appendSet(day: Day, v: LogSetVars): Day {
    return withGroup(day, v.exercise_id, (g) => {
        const logged = g.logged ?? emptyLogged(day, g, v.exercise_id, v.created_at);
        const fields = setFieldsOf(v);
        const existing = logged.sets.findIndex((s) => s.client_id === v.client_id);
        const sets = [...logged.sets];
        if (existing >= 0) {
            sets[existing] = { ...sets[existing], ...fields };
        } else {
            sets.push({
                ID: 0,
                logged_exercise_id: logged.ID,
                client_id: v.client_id,
                ...fields,
                created_at: v.created_at,
                updated_at: v.created_at,
            });
        }
        return { ...g, logged: { ...logged, sets } };
    });
}

/** Log set (revert) and delete set (apply): remove by key. */
export function removeSetByKey(day: Day, v: { exercise_id: number; client_id: string }): Day {
    return withGroup(day, v.exercise_id, (g) => {
        if (!g.logged) return g;
        const sets = g.logged.sets.filter((s) => setKey(s) !== v.client_id);
        if (sets.length === g.logged.sets.length) return g;
        return dropIfEmptyLocal(g, { ...g.logged, sets });
    });
}

/** Edit set (apply). Idempotent. */
export function patchSet(day: Day, v: UpdateSetVars): Day {
    return withGroup(day, v.exercise_id, (g) => {
        if (!g.logged || !g.logged.sets.some((s) => setKey(s) === v.client_id)) return g;
        const sets = g.logged.sets.map((s) =>
            setKey(s) === v.client_id ? { ...s, ...setFieldsOf(v) } : s,
        );
        return { ...g, logged: { ...g.logged, sets } };
    });
}

/**
 * Delete set (revert): put the snapshot's set back at its old index, unless it is already there.
 * Recreates `logged` from the snapshot when the delete had removed it.
 */
export function restoreSet(cur: Day, v: DeleteSetVars, snapshot: Day): Day {
    const before = findGroup(snapshot, v.exercise_id)?.logged;
    const index = before?.sets.findIndex((s) => setKey(s) === v.client_id) ?? -1;
    if (!before || index < 0) return cur;
    const set = before.sets[index];
    return withGroup(cur, v.exercise_id, (g) => {
        const logged = g.logged ?? { ...before, sets: [] };
        if (logged.sets.some((s) => setKey(s) === v.client_id)) return g;
        const sets = [...logged.sets];
        sets.splice(Math.min(index, sets.length), 0, set);
        return { ...g, logged: { ...logged, sets } };
    });
}

/** Notes (apply/revert). Only patches an existing logged exercise. */
export function setNotes(day: Day, v: { exercise_id: number; notes: string }): Day {
    return withGroup(day, v.exercise_id, (g) => {
        if (!g.logged || g.logged.notes === v.notes) return g;
        return { ...g, logged: { ...g.logged, notes: v.notes } };
    });
}

/** Writes other mutations still have in flight (or failed and marked) for this day. */
export type LocalSetIntents = {
    /** Set keys whose fields are being written locally (pending/failed create or edit): keep local values. */
    keepFields: ReadonlySet<string>;
    /** Set keys with a delete in flight: never re-add them from a server answer. */
    deleting: ReadonlySet<string>;
    /** A notes save for this exercise is in flight: keep the local notes. */
    keepNotes: boolean;
};

export const NO_INTENTS: LocalSetIntents = { keepFields: new Set(), deleting: new Set(), keepNotes: false };

/**
 * Reconcile: merge the server's full `LoggedExercise` into the group, by set key, without wiping
 * anything other mutations still own:
 * - saved sets come from the server, in its order (by id), except keys being deleted;
 * - a saved set with a pending/failed edit keeps its local fields (the edit lands later);
 * - local sets the server doesn't have yet (pending or failed creates) stay, after the saved ones.
 * Sets with a server id that the server no longer returns are dropped (deleted elsewhere).
 */
export function mergeLoggedExercise(
    day: Day,
    exerciseId: number,
    saved: LoggedExercise,
    intents: LocalSetIntents = NO_INTENTS,
): Day {
    return withGroup(day, exerciseId, (g) => {
        const current = g.logged;
        const local = new Map((current?.sets ?? []).map((s) => [setKey(s), s]));
        const savedSets = saved.sets ?? [];
        const savedKeys = new Set<string>();
        const sets: LoggedSet[] = [];
        for (const s of savedSets) {
            const key = setKey(s);
            savedKeys.add(key);
            // A legacy set (no client_id) is also known locally under its numeric id.
            savedKeys.add(`${ID_PREFIX}${s.ID}`);
            if (intents.deleting.has(key)) continue;
            const mine = local.get(key);
            sets.push(
                mine && intents.keepFields.has(key)
                    ? { ...s, ...setFieldsOf(mine) }
                    : s,
            );
        }
        for (const s of current?.sets ?? []) {
            if (isUnsavedSet(s) && !savedKeys.has(setKey(s)) && !intents.deleting.has(setKey(s))) sets.push(s);
        }
        const exercise = saved.exercise?.ID ? saved.exercise : (current?.exercise ?? saved.exercise);
        const notes = intents.keepNotes && current ? current.notes : saved.notes;
        return { ...g, logged: { ...saved, exercise, notes, sets } };
    });
}

/**
 * Delete reconcile when the server removed the parent with its last set (`exercise_removed`).
 * Unsaved local sets keep a local parent (their create will upsert it again); otherwise the logged
 * exercise goes, and so does the group when it wasn't planned (it only existed because of the log).
 */
export function dropRemovedExercise(day: Day, exerciseId: number): Day {
    return withGroup(day, exerciseId, (g) => {
        if (!g.logged) return g;
        const unsaved = g.logged.sets.filter(isUnsavedSet);
        if (unsaved.length > 0) return { ...g, logged: { ...g.logged, ID: 0, sets: unsaved } };
        if (!g.planned) return null;
        const { logged: _drop, ...rest } = g;
        return rest;
    });
}

// ── View rows ──────────────────────────────────────────────────────────────────────────────────

export type SetRowStatus = "pending" | "success" | "error";

/** One row of the "today" list: a set with its derived save state. */
export type SetRow = {
    /** `setKey` of the set. */
    key: string;
    weight: number;
    reps: number;
    weight_setup: string;
    plate_setup: PlateSetup | null;
    status: SetRowStatus;
    /** Server id, null while unsaved. */
    id: number | null;
    /** Which write failed, so Retry re-runs the right one. */
    failed: "create" | "update" | null;
};

/** Pending/failed set writes, from the mutation cache (`usePendingIds` / `useMutationState`). */
export type SetIntentState = {
    /** Keys with any create/edit/delete in flight. */
    pending: ReadonlySet<string>;
    /** Keys with a delete in flight (hidden from the list). */
    deleting: ReadonlySet<string>;
    failedCreate: ReadonlySet<string>;
    failedUpdate: ReadonlySet<string>;
    /** Variables of pending or failed creates, in call order. */
    creates: readonly LogSetVars[];
    /** Variables of pending or failed edits, in call order. */
    updates: readonly UpdateSetVars[];
};

export const EMPTY_SET_INTENTS: SetIntentState = {
    pending: new Set(),
    deleting: new Set(),
    failedCreate: new Set(),
    failedUpdate: new Set(),
    creates: [],
    updates: [],
};

/**
 * The rows shown for one exercise on one day: the cached sets, plus unsaved writes a refetch may have
 * dropped from the cache (the F6 gotcha: a refetch replaces the entry with server data, which never
 * had a failed set). Pending/failed edits overlay their values, pending deletes are hidden.
 */
export function deriveSetRows(
    sets: readonly LoggedSet[] | undefined,
    intents: SetIntentState,
    target: { date: string; exercise_id: number },
): SetRow[] {
    const mine = <T extends { date: string; exercise_id: number }>(v: T) =>
        v.date === target.date && v.exercise_id === target.exercise_id;

    const rows: SetRow[] = (sets ?? []).map((s) => ({
        key: setKey(s),
        weight: s.weight,
        reps: s.reps,
        weight_setup: s.weight_setup || "",
        plate_setup: s.plate_setup ?? null,
        status: "success",
        id: s.ID > 0 ? s.ID : null,
        failed: null,
    }));
    const known = new Set(rows.map((r) => r.key));
    for (const v of intents.creates) {
        if (!mine(v) || known.has(v.client_id)) continue;
        known.add(v.client_id);
        rows.push({
            key: v.client_id,
            ...setFieldsOf(v),
            status: "success",
            id: null,
            failed: null,
        });
    }
    for (const v of intents.updates) {
        if (!mine(v)) continue;
        const row = rows.find((r) => r.key === v.client_id);
        if (row) Object.assign(row, setFieldsOf(v));
    }
    return rows
        .filter((r) => !intents.deleting.has(r.key))
        .map((r) => {
            if (intents.pending.has(r.key)) return { ...r, status: "pending" as const };
            const failed = intents.failedCreate.has(r.key)
                ? ("create" as const)
                : intents.failedUpdate.has(r.key)
                  ? ("update" as const)
                  : null;
            return failed ? { ...r, status: "error" as const, failed } : r;
        });
}

// ── Retry / edit of unsaved rows ───────────────────────────────────────────────────────────────

/** The latest failed create and edit of one set (from the mutation cache). */
export type FailedSetWrites = { create?: LogSetVars; update?: UpdateSetVars };

export type SetWritePlan =
    | { kind: "create"; vars: LogSetVars; dismissUpdate?: UpdateSetVars }
    | { kind: "update"; vars: UpdateSetVars }
    | null;

/**
 * What Retry (no `fields`) or an edit of a failed row (`fields`) sends. A failed create wins: an
 * edit queued behind it can only fail too (the set doesn't exist), so the create is re-sent with
 * the newest values (the failed edit's, then `fields`) and the failed edit is dismissed.
 */
export function planSetRetry(failed: FailedSetWrites, fields?: SetFields): SetWritePlan {
    if (failed.create) {
        const fromEdit = failed.update
            ? setFieldsOf(failed.update)
            : {};
        return {
            kind: "create",
            vars: { ...failed.create, ...fromEdit, ...fields },
            ...(failed.update && { dismissUpdate: failed.update }),
        };
    }
    if (failed.update) return { kind: "update", vars: { ...failed.update, ...fields } };
    return null;
}
