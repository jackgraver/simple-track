import { describe, expect, it } from "vitest";

import type { WorkoutLogsPreviousResponse } from "~/api/workout/api";
import type { Exercise, LoggedExercise, LoggedSet } from "~/types/workout";
import {
    appendSet,
    deriveSetRows,
    dropRemovedExercise,
    EMPTY_SET_INTENTS,
    findGroup,
    mergeLoggedExercise,
    patchSet,
    removeSetByKey,
    restoreSet,
    setKey,
    setNotes,
    setRef,
    withGroup,
    type LogSetVars,
    type SetIntentState,
    planSetRetry,
} from "./setsCache";

const DATE = "2026-11-22";
const BENCH = 7;
const SQUAT = 9;

const exercise = (ID: number, name: string): Exercise => ({ ID, name, rep_rollover: 10, cues: "", created_at: "", updated_at: "" });

function set(ID: number, client_id: string | null, weight = 100, reps = 8): LoggedSet {
    return { ID, client_id, logged_exercise_id: 50, weight, reps, weight_setup: "", created_at: "t", updated_at: "t" };
}

function logged(sets: LoggedSet[], extra: Partial<LoggedExercise> = {}): LoggedExercise {
    return { ID: 50, workout_log_id: 3, exercise_id: BENCH, exercise: exercise(BENCH, "Bench"), sets, notes: "", created_at: "", updated_at: "", ...extra };
}

function day(benchLogged?: LoggedExercise): WorkoutLogsPreviousResponse {
    return {
        day: { ID: 3 } as WorkoutLogsPreviousResponse["day"],
        planned_exercises: [
            { planned: exercise(BENCH, "Bench"), logged: benchLogged, previous: logged([set(10, null, 95, 8)], { ID: 40, notes: "last" }) },
            { planned: exercise(SQUAT, "Squat") },
        ],
        planned_cardio: null,
        logged_cardio: null,
        planned_pre_mobility: null,
        logged_pre_mobility: null,
        planned_post_mobility: null,
        logged_post_mobility: null,
    };
}

const logVars = (client_id: string, weight = 100, reps = 8, exercise_id = BENCH): LogSetVars => ({
    client_id,
    date: DATE,
    exercise_id,
    weight,
    reps,
    weight_setup: "",
    created_at: "2026-11-22T10:00:00Z",
});

const benchSets = (d: WorkoutLogsPreviousResponse) => findGroup(d, BENCH)?.logged?.sets ?? [];

describe("keys", () => {
    it("uses client_id when present, else the numeric id", () => {
        expect(setKey(set(5, "c1"))).toBe("c1");
        expect(setKey(set(5, null))).toBe("id:5");
        expect(setRef("id:5")).toBe("5");
        expect(setRef("c1")).toBe("c1");
    });
});

describe("withGroup", () => {
    it("returns the same day when nothing matches or nothing changed", () => {
        const d = day();
        expect(withGroup(d, 999, () => null)).toBe(d);
        expect(withGroup(d, BENCH, (g) => g)).toBe(d);
    });
    it("removes a group when fn returns null, without touching the input", () => {
        const d = day();
        const next = withGroup(d, SQUAT, () => null);
        expect(next.planned_exercises).toHaveLength(1);
        expect(d.planned_exercises).toHaveLength(2);
    });
});

describe("appendSet", () => {
    it("creates a local logged exercise for the first set", () => {
        const next = appendSet(day(), logVars("c1"));
        const g = findGroup(next, BENCH)!;
        expect(g.logged).toMatchObject({ ID: 0, exercise_id: BENCH, workout_log_id: 3, notes: "" });
        expect(g.logged!.exercise.name).toBe("Bench");
        expect(g.logged!.sets).toEqual([
            expect.objectContaining({ ID: 0, client_id: "c1", weight: 100, reps: 8, created_at: "2026-11-22T10:00:00Z" }),
        ]);
    });

    it("appends after saved sets and is idempotent (Retry re-applies)", () => {
        const once = appendSet(day(logged([set(1, "a")])), logVars("c1", 105, 6));
        const twice = appendSet(once, logVars("c1", 105, 6));
        expect(benchSets(twice).map(setKey)).toEqual(["a", "c1"]);
        expect(benchSets(twice)[1]).toMatchObject({ weight: 105, reps: 6 });
    });

    it("keeps two identical sets apart (3×10 @ 135 is common)", () => {
        const d = appendSet(appendSet(day(), logVars("c1")), logVars("c2"));
        expect(benchSets(d).map(setKey)).toEqual(["c1", "c2"]);
    });

    it("leaves other exercises and the input untouched", () => {
        const d = day();
        const next = appendSet(d, logVars("c1"));
        expect(findGroup(d, BENCH)!.logged).toBeUndefined();
        expect(findGroup(next, SQUAT)).toBe(findGroup(d, SQUAT));
    });

    it("is a no-op for an exercise not in the day", () => {
        const d = day();
        expect(appendSet(d, logVars("c1", 100, 8, 999))).toBe(d);
    });
});

describe("removeSetByKey", () => {
    it("reverts a first set by dropping the local parent too", () => {
        const added = appendSet(day(), logVars("c1"));
        const reverted = removeSetByKey(added, { exercise_id: BENCH, client_id: "c1" });
        expect(findGroup(reverted, BENCH)!.logged).toBeUndefined();
    });

    it("keeps a saved parent (the server decides when it goes)", () => {
        const d = removeSetByKey(day(logged([set(1, "a")])), { exercise_id: BENCH, client_id: "a" });
        expect(findGroup(d, BENCH)!.logged).toMatchObject({ ID: 50, sets: [] });
    });

    it("removes legacy sets by id key and only that set", () => {
        const d = removeSetByKey(day(logged([set(1, null), set(2, "b")])), { exercise_id: BENCH, client_id: "id:1" });
        expect(benchSets(d).map(setKey)).toEqual(["b"]);
    });

    it("returns the same day when the set is not there", () => {
        const d = day(logged([set(1, "a")]));
        expect(removeSetByKey(d, { exercise_id: BENCH, client_id: "zzz" })).toBe(d);
    });
});

describe("patchSet", () => {
    it("changes only the targeted set's fields", () => {
        const d = patchSet(day(logged([set(1, "a"), set(2, "b")])), {
            client_id: "b",
            date: DATE,
            exercise_id: BENCH,
            weight: 110,
            reps: 5,
            weight_setup: "45+10",
        });
        expect(benchSets(d)[0]).toMatchObject({ weight: 100, reps: 8 });
        expect(benchSets(d)[1]).toMatchObject({ ID: 2, client_id: "b", weight: 110, reps: 5, weight_setup: "45+10" });
    });
});

describe("restoreSet (delete revert)", () => {
    it("puts the set back at its old index", () => {
        const snapshot = day(logged([set(1, "a"), set(2, "b"), set(3, "c")]));
        const deleted = removeSetByKey(snapshot, { exercise_id: BENCH, client_id: "b" });
        const back = restoreSet(deleted, { client_id: "b", date: DATE, exercise_id: BENCH }, snapshot);
        expect(benchSets(back).map(setKey)).toEqual(["a", "b", "c"]);
    });

    it("keeps sets added after the snapshot and doesn't duplicate", () => {
        const snapshot = day(logged([set(1, "a"), set(2, "b")]));
        let cur = removeSetByKey(snapshot, { exercise_id: BENCH, client_id: "b" });
        cur = appendSet(cur, logVars("c3"));
        const back = restoreSet(cur, { client_id: "b", date: DATE, exercise_id: BENCH }, snapshot);
        expect(benchSets(back).map(setKey)).toEqual(["a", "b", "c3"]);
        expect(restoreSet(back, { client_id: "b", date: DATE, exercise_id: BENCH }, snapshot)).toBe(back);
    });

    it("recreates the logged exercise when it was dropped", () => {
        const snapshot = day(logged([set(1, "a")]));
        const dropped = dropRemovedExercise(removeSetByKey(snapshot, { exercise_id: BENCH, client_id: "a" }), BENCH);
        expect(findGroup(dropped, BENCH)!.logged).toBeUndefined();
        const back = restoreSet(dropped, { client_id: "a", date: DATE, exercise_id: BENCH }, snapshot);
        expect(findGroup(back, BENCH)!.logged).toMatchObject({ ID: 50, sets: [expect.objectContaining({ client_id: "a" })] });
    });
});

describe("setNotes", () => {
    it("patches an existing logged exercise only", () => {
        expect(findGroup(setNotes(day(logged([])), { exercise_id: BENCH, notes: "x" }), BENCH)!.logged!.notes).toBe("x");
        const d = day();
        expect(setNotes(d, { exercise_id: BENCH, notes: "x" })).toBe(d);
    });
});

describe("mergeLoggedExercise (reconcile)", () => {
    it("swaps the optimistic first set and parent for the server's", () => {
        const optimistic = appendSet(day(), logVars("c1"));
        const saved = logged([set(11, "c1")], { ID: 77 });
        const merged = mergeLoggedExercise(optimistic, BENCH, saved);
        expect(findGroup(merged, BENCH)!.logged).toMatchObject({ ID: 77 });
        expect(benchSets(merged)).toEqual([set(11, "c1")]);
    });

    it("keeps sibling sets that are still unsaved, after the saved ones", () => {
        let d = appendSet(day(), logVars("c1"));
        d = appendSet(d, logVars("c2"));
        const merged = mergeLoggedExercise(d, BENCH, logged([set(11, "c1")], { ID: 77 }));
        expect(benchSets(merged).map((s) => [setKey(s), s.ID])).toEqual([
            ["c1", 11],
            ["c2", 0],
        ]);
    });

    it("never duplicates when the server already has the sibling", () => {
        let d = appendSet(day(), logVars("c1"));
        d = appendSet(d, logVars("c2"));
        const merged = mergeLoggedExercise(d, BENCH, logged([set(11, "c1"), set(12, "c2")]));
        expect(benchSets(merged).map(setKey)).toEqual(["c1", "c2"]);
    });

    it("keeps local fields for a set with an edit in flight", () => {
        const d = patchSet(day(logged([set(1, "a")])), { client_id: "a", date: DATE, exercise_id: BENCH, weight: 120, reps: 3, weight_setup: "" });
        const merged = mergeLoggedExercise(d, BENCH, logged([set(1, "a", 100, 8)]), {
            keepFields: new Set(["a"]),
            deleting: new Set(),
            keepNotes: false,
        });
        expect(benchSets(merged)[0]).toMatchObject({ ID: 1, weight: 120, reps: 3 });
    });

    it("doesn't bring back a set with a delete in flight", () => {
        const d = removeSetByKey(day(logged([set(1, "a"), set(2, "b")])), { exercise_id: BENCH, client_id: "a" });
        const merged = mergeLoggedExercise(d, BENCH, logged([set(1, "a"), set(2, "b")]), {
            keepFields: new Set(),
            deleting: new Set(["a"]),
            keepNotes: false,
        });
        expect(benchSets(merged).map(setKey)).toEqual(["b"]);
    });

    it("drops saved sets the server no longer has, and matches legacy sets by id", () => {
        const d = day(logged([set(1, null), set(2, "b")]));
        const merged = mergeLoggedExercise(d, BENCH, logged([set(1, null)]));
        expect(benchSets(merged).map(setKey)).toEqual(["id:1"]);
    });

    it("keeps local notes while a notes save is in flight, else takes the server's", () => {
        const d = setNotes(day(logged([set(1, "a")])), { exercise_id: BENCH, notes: "mine" });
        const server = logged([set(1, "a")], { notes: "old" });
        expect(findGroup(mergeLoggedExercise(d, BENCH, server), BENCH)!.logged!.notes).toBe("old");
        expect(
            findGroup(mergeLoggedExercise(d, BENCH, server, { keepFields: new Set(), deleting: new Set(), keepNotes: true }), BENCH)!
                .logged!.notes,
        ).toBe("mine");
    });

    it("keeps the cached exercise when the response lacks it", () => {
        const d = day(logged([set(1, "a")]));
        const merged = mergeLoggedExercise(d, BENCH, logged([set(1, "a")], { exercise: { ID: 0 } as Exercise }));
        expect(findGroup(merged, BENCH)!.logged!.exercise.name).toBe("Bench");
    });
});

describe("dropRemovedExercise", () => {
    it("drops logged from a planned group", () => {
        const d = dropRemovedExercise(day(logged([])), BENCH);
        expect(findGroup(d, BENCH)!.logged).toBeUndefined();
        expect(findGroup(d, BENCH)!.planned!.ID).toBe(BENCH);
    });

    it("removes an unplanned group entirely", () => {
        const d = day();
        d.planned_exercises.push({ logged: logged([], { exercise_id: 42 }) });
        expect(findGroup(dropRemovedExercise(d, 42), 42)).toBeUndefined();
    });

    it("keeps a local parent for unsaved sets (their create upserts it again)", () => {
        const d = appendSet(day(logged([])), logVars("c9"));
        const next = dropRemovedExercise(d, BENCH);
        expect(findGroup(next, BENCH)!.logged).toMatchObject({ ID: 0, sets: [expect.objectContaining({ client_id: "c9" })] });
    });
});

describe("deriveSetRows", () => {
    const target = { date: DATE, exercise_id: BENCH };
    const intents = (over: Partial<SetIntentState>): SetIntentState => ({ ...EMPTY_SET_INTENTS, ...over });

    it("maps cached sets to saved rows", () => {
        expect(deriveSetRows([set(1, "a"), set(0, "c1")], EMPTY_SET_INTENTS, target)).toEqual([
            { key: "a", weight: 100, reps: 8, weight_setup: "", plate_setup: null, status: "success", id: 1, failed: null },
            { key: "c1", weight: 100, reps: 8, weight_setup: "", plate_setup: null, status: "success", id: null, failed: null },
        ]);
    });

    it("marks pending and failed rows", () => {
        const rows = deriveSetRows(
            [set(1, "a"), set(0, "c1"), set(2, "b")],
            intents({ pending: new Set(["c1"]), failedUpdate: new Set(["b"]) }),
            target,
        );
        expect(rows.map((r) => [r.key, r.status, r.failed])).toEqual([
            ["a", "success", null],
            ["c1", "pending", null],
            ["b", "error", "update"],
        ]);
    });

    it("re-adds a failed create that a refetch dropped from the cache (F6 gotcha)", () => {
        const rows = deriveSetRows(
            [set(1, "a")],
            intents({ failedCreate: new Set(["c1"]), creates: [logVars("c1", 110, 4)] }),
            target,
        );
        expect(rows).toEqual([
            expect.objectContaining({ key: "a", status: "success" }),
            { key: "c1", weight: 110, reps: 4, weight_setup: "", plate_setup: null, status: "error", id: null, failed: "create" },
        ]);
    });

    it("doesn't duplicate a create that is already cached", () => {
        const rows = deriveSetRows([set(0, "c1")], intents({ creates: [logVars("c1")], pending: new Set(["c1"]) }), target);
        expect(rows).toHaveLength(1);
    });

    it("ignores writes for other days and exercises", () => {
        const other = [{ ...logVars("x"), date: "2026-11-21" }, logVars("y", 100, 8, SQUAT)];
        expect(deriveSetRows([], intents({ creates: other }), target)).toEqual([]);
    });

    it("overlays edits in flight and hides deletes in flight", () => {
        const rows = deriveSetRows(
            [set(1, "a"), set(2, "b")],
            intents({
                pending: new Set(["a", "b"]),
                deleting: new Set(["b"]),
                updates: [{ client_id: "a", date: DATE, exercise_id: BENCH, weight: 130, reps: 2, weight_setup: "" }],
            }),
            target,
        );
        expect(rows).toEqual([expect.objectContaining({ key: "a", weight: 130, reps: 2, status: "pending" })]);
    });
});

describe("planSetRetry", () => {
    const create = { client_id: "c1", date: "2026-11-22", exercise_id: 5, created_at: "t", reps: 8, weight: 100, weight_setup: "" };
    const update = { client_id: "c1", date: "2026-11-22", exercise_id: 5, reps: 10, weight: 105, weight_setup: "" };

    it("nothing failed: nothing to send", () => {
        expect(planSetRetry({})).toBeNull();
    });

    it("a failed edit is re-sent, with new values when editing", () => {
        expect(planSetRetry({ update })).toEqual({ kind: "update", vars: update });
        expect(planSetRetry({ update }, { reps: 6, weight: 90, weight_setup: "x" })).toEqual({
            kind: "update",
            vars: { ...update, reps: 6, weight: 90, weight_setup: "x" },
        });
    });

    it("a failed create wins over the edit queued behind it and takes its values", () => {
        expect(planSetRetry({ create, update })).toEqual({
            kind: "create",
            vars: { ...create, reps: 10, weight: 105, plate_setup: null },
            dismissUpdate: update,
        });
        const edited = planSetRetry({ create }, { reps: 3, weight: 50, weight_setup: "" });
        expect(edited).toEqual({ kind: "create", vars: { ...create, reps: 3, weight: 50 } });
    });
});
