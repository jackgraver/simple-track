import type { QueryClient } from "@tanstack/vue-query";
import { AxiosError, AxiosHeaders } from "axios";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, type EffectScope } from "vue";

import { createQueryClient } from "~/api/queryClient";
import type { WorkoutLogsPreviousResponse } from "~/api/workout/api";
import type { Exercise, LoggedExercise, LoggedSet } from "~/types/workout";
import { workoutKeys } from "./keys";
import { findGroup, setKey, type LogSetVars } from "./setsCache";

const http = vi.hoisted(() => ({ apiGET: vi.fn(), apiPOST: vi.fn(), apiPATCH: vi.fn(), apiDELETE: vi.fn() }));
vi.mock("~/api/client", () => http);
const push = vi.hoisted(() => vi.fn());
vi.mock("~/composables/toast/useToast", () => ({ toast: { push } }));
const qcRef = vi.hoisted(() => ({ qc: null as unknown }));
vi.mock("@tanstack/vue-query", async (orig) => ({
    ...(await orig<typeof import("@tanstack/vue-query")>()),
    useQueryClient: () => qcRef.qc,
}));

import { retrySetWrite, useDeleteSet, useLogSet, useSaveExerciseNotes, useSetIntents, useUpdateSet, workoutQueries } from "./queries";

const DATE = "2026-11-22";
const BENCH = 7;
const KEY = workoutKeys.days.day(DATE);

const exercise: Exercise = { ID: BENCH, name: "Bench", rep_rollover: 10, cues: "", created_at: "", updated_at: "" };
const set = (ID: number, client_id: string | null, reps = 8): LoggedSet => ({
    ID,
    client_id,
    logged_exercise_id: 50,
    weight: 100,
    reps,
    weight_setup: "",
    created_at: "t",
    updated_at: "t",
});
const logged = (sets: LoggedSet[], extra: Partial<LoggedExercise> = {}): LoggedExercise => ({
    ID: 50,
    workout_log_id: 3,
    exercise_id: BENCH,
    exercise,
    sets,
    notes: "",
    created_at: "",
    updated_at: "",
    ...extra,
});
function day(benchLogged?: LoggedExercise): WorkoutLogsPreviousResponse {
    return {
        day: { ID: 3 } as WorkoutLogsPreviousResponse["day"],
        planned_exercises: [{ planned: exercise, logged: benchLogged }],
        planned_cardio: null,
        logged_cardio: null,
        planned_pre_mobility: null,
        logged_pre_mobility: null,
        planned_post_mobility: null,
        logged_post_mobility: null,
    };
}
const vars = (client_id: string, reps = 8): LogSetVars => ({
    client_id,
    date: DATE,
    exercise_id: BENCH,
    weight: 100,
    reps,
    weight_setup: "",
    created_at: "2026-11-22T10:00:00Z",
});

function httpError(status: number): AxiosError {
    const config = { headers: new AxiosHeaders() };
    return new AxiosError("fail", "ERR_BAD_REQUEST", config, null, { status, statusText: "", headers: {}, config, data: {} });
}

function deferred<T>() {
    let resolve!: (v: T) => void;
    let reject!: (e: unknown) => void;
    const promise = new Promise<T>((res, rej) => {
        resolve = res;
        reject = rej;
    });
    return { promise, resolve, reject };
}

const flush = () => vi.advanceTimersByTimeAsync(0);
const sets = () => findGroup(qc.getQueryData<WorkoutLogsPreviousResponse>(KEY), BENCH)?.logged?.sets ?? [];

let qc: QueryClient;
let scope: EffectScope;
beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    qc = createQueryClient(vi.fn());
    qcRef.qc = qc;
    scope = effectScope();
});
afterEach(() => {
    scope.stop();
    vi.useRealTimers();
});

describe("useLogSet", () => {
    it("shows the set at once, posts the per-set body, and reconciles by client_id", async () => {
        qc.setQueryData(KEY, day());
        const d = deferred<unknown>();
        http.apiPOST.mockReturnValue(d.promise);
        const m = scope.run(() => useLogSet())!;
        m.mutate(vars("c1"));
        await flush();
        expect(sets()).toEqual([expect.objectContaining({ ID: 0, client_id: "c1" })]);
        expect(http.apiPOST).toHaveBeenCalledWith("/workout/sets", {
            client_id: "c1",
            date: DATE,
            exercise_id: BENCH,
            reps: 8,
            weight: 100,
            weight_setup: "",
        });
        d.resolve({ exercise: logged([set(11, "c1")], { ID: 77 }), set: set(11, "c1"), created: true, first_set_of_day: false });
        await flush();
        expect(sets()).toEqual([set(11, "c1")]);
        expect(findGroup(qc.getQueryData(KEY), BENCH)!.logged!.ID).toBe(77);
    });

    it("invalidates activity only for the day's first set", async () => {
        qc.setQueryData(KEY, day());
        const spy = vi.spyOn(qc, "invalidateQueries");
        http.apiPOST.mockResolvedValueOnce({ exercise: logged([set(11, "c1")]), set: set(11, "c1"), created: true, first_set_of_day: true });
        http.apiPOST.mockResolvedValueOnce({
            exercise: logged([set(11, "c1"), set(12, "c2")]),
            set: set(12, "c2"),
            created: true,
            first_set_of_day: false,
        });
        const m = scope.run(() => useLogSet())!;
        m.mutate(vars("c1"));
        await flush();
        expect(spy).toHaveBeenCalledWith({ queryKey: workoutKeys.activity.all });
        spy.mockClear();
        m.mutate(vars("c2"));
        await flush();
        expect(spy).not.toHaveBeenCalled();
    });

    it("two rapid sets: the first reconcile keeps the second, still pending, row", async () => {
        qc.setQueryData(KEY, day());
        const d1 = deferred<unknown>();
        const d2 = deferred<unknown>();
        http.apiPOST.mockReturnValueOnce(d1.promise).mockReturnValueOnce(d2.promise);
        const m = scope.run(() => useLogSet())!;
        m.mutate(vars("c1"));
        m.mutate(vars("c2"));
        await flush();
        expect(sets().map(setKey)).toEqual(["c1", "c2"]);
        // Same exercise scope: the second request waits for the first.
        expect(http.apiPOST).toHaveBeenCalledTimes(1);
        d1.resolve({ exercise: logged([set(11, "c1")]), set: set(11, "c1"), created: true, first_set_of_day: true });
        await flush();
        expect(sets().map((s) => [setKey(s), s.ID])).toEqual([
            ["c1", 11],
            ["c2", 0],
        ]);
        expect(http.apiPOST).toHaveBeenCalledTimes(2);
        d2.resolve({ exercise: logged([set(11, "c1"), set(12, "c2")]), set: set(12, "c2"), created: true, first_set_of_day: false });
        await flush();
        expect(sets().map((s) => s.ID)).toEqual([11, 12]);
    });

    it("failure keeps the row marked; Retry resends the same client_id once and clears it", async () => {
        qc.setQueryData(KEY, day());
        http.apiPOST.mockRejectedValue(new Error("offline"));
        const m = scope.run(() => useLogSet())!;
        const intents = scope.run(() => useSetIntents())!;
        const v = vars("c1");
        m.mutate(v);
        await vi.advanceTimersByTimeAsync(10_000); // retry: 2 with backoff
        expect(http.apiPOST).toHaveBeenCalledTimes(3);
        expect(sets().map(setKey)).toEqual(["c1"]);
        expect(intents.value.failedCreate.has("c1")).toBe(true);
        expect(push).not.toHaveBeenCalled(); // 'mark' mode: inline, no toast

        // A background refetch drops the unsaved row from the cache; the intents still carry it.
        qc.setQueryData(KEY, day());
        expect(intents.value.creates.map((c) => c.client_id)).toEqual(["c1"]);

        http.apiPOST.mockReset();
        http.apiPOST.mockResolvedValue({ exercise: logged([set(11, "c1")]), set: set(11, "c1"), created: true, first_set_of_day: true });
        m.retry(v);
        await flush();
        expect(http.apiPOST).toHaveBeenCalledTimes(1);
        expect(http.apiPOST.mock.calls[0][1]).toMatchObject({ client_id: "c1" });
        expect(sets()).toEqual([set(11, "c1")]);
        expect(intents.value.failedCreate.size).toBe(0);
        expect(intents.value.creates).toEqual([]);
    });
});

describe("useUpdateSet", () => {
    it("patches at once, PATCHes by client_id or numeric id, and keeps a later edit's values", async () => {
        qc.setQueryData(KEY, day(logged([set(1, null), set(2, "b")])));
        const d1 = deferred<unknown>();
        const d2 = deferred<unknown>();
        http.apiPATCH.mockReturnValueOnce(d1.promise).mockReturnValueOnce(d2.promise);
        const m = scope.run(() => useUpdateSet())!;
        m.mutate({ client_id: "id:1", date: DATE, exercise_id: BENCH, weight: 100, reps: 5, weight_setup: "" });
        m.mutate({ client_id: "b", date: DATE, exercise_id: BENCH, weight: 100, reps: 6, weight_setup: "" });
        await flush();
        expect(sets().map((s) => s.reps)).toEqual([5, 6]);
        expect(http.apiPATCH).toHaveBeenCalledWith("/workout/sets/1", { reps: 5, weight: 100, weight_setup: "" });
        // The first answer predates the second edit: set b must keep its local 6 reps.
        d1.resolve({ exercise: logged([set(1, null, 5), set(2, "b", 8)]), set: set(1, null, 5) });
        await flush();
        expect(sets().map((s) => s.reps)).toEqual([5, 6]);
        expect(http.apiPATCH).toHaveBeenLastCalledWith("/workout/sets/b", { reps: 6, weight: 100, weight_setup: "" });
        d2.resolve({ exercise: logged([set(1, null, 5), set(2, "b", 6)]), set: set(2, "b", 6) });
        await flush();
        expect(sets().map((s) => s.reps)).toEqual([5, 6]);
    });

    it("an edit of a set still being created waits for the create", async () => {
        qc.setQueryData(KEY, day());
        const create = deferred<unknown>();
        http.apiPOST.mockReturnValue(create.promise);
        http.apiPATCH.mockResolvedValue({ exercise: logged([set(11, "c1", 3)]), set: set(11, "c1", 3) });
        const log = scope.run(() => useLogSet())!;
        const edit = scope.run(() => useUpdateSet())!;
        log.mutate(vars("c1"));
        edit.mutate({ client_id: "c1", date: DATE, exercise_id: BENCH, weight: 100, reps: 3, weight_setup: "" });
        await flush();
        expect(sets()[0].reps).toBe(3);
        expect(http.apiPATCH).not.toHaveBeenCalled();
        create.resolve({ exercise: logged([set(11, "c1", 8)]), set: set(11, "c1", 8), created: true, first_set_of_day: false });
        await flush();
        expect(http.apiPATCH).toHaveBeenCalledWith("/workout/sets/c1", expect.anything());
        expect(sets()).toEqual([set(11, "c1", 3)]);
    });
});

describe("useDeleteSet", () => {
    it("removes at once and drops the logged exercise when the server removed it", async () => {
        qc.setQueryData(KEY, day(logged([set(1, "a")])));
        const d = deferred<unknown>();
        http.apiDELETE.mockReturnValue(d.promise);
        const m = scope.run(() => useDeleteSet())!;
        const intents = scope.run(() => useSetIntents())!;
        m.mutate({ client_id: "a", date: DATE, exercise_id: BENCH });
        await flush();
        expect(sets()).toEqual([]);
        expect(intents.value.deleting.has("a")).toBe(true);
        expect(http.apiDELETE).toHaveBeenCalledWith("/workout/sets/a");
        d.resolve({ success: true, deleted: true, logged_exercise_id: 50, exercise_removed: true });
        await flush();
        expect(findGroup(qc.getQueryData(KEY), BENCH)!.logged).toBeUndefined();
    });

    it("rolls back to the old index with a Retry toast on failure", async () => {
        qc.setQueryData(KEY, day(logged([set(1, "a"), set(2, "b"), set(3, "c")])));
        http.apiDELETE.mockRejectedValue(new Error("offline"));
        const m = scope.run(() => useDeleteSet())!;
        m.mutate({ client_id: "b", date: DATE, exercise_id: BENCH });
        await vi.advanceTimersByTimeAsync(10_000);
        expect(sets().map(setKey)).toEqual(["a", "b", "c"]);
        expect(push).toHaveBeenCalledWith("Couldn’t delete the set", "error", expect.objectContaining({ action: expect.anything() }));
    });

    it("reverts only its own row while a create on another exercise of the day is pending", async () => {
        const SQUAT = 9;
        const d0 = day(logged([set(1, "a"), set(2, "b")]));
        d0.planned_exercises.push({ planned: { ...exercise, ID: SQUAT, name: "Squat" } });
        qc.setQueryData(KEY, d0);
        const create = deferred<unknown>();
        http.apiPOST.mockReturnValue(create.promise);
        http.apiDELETE.mockRejectedValue(new Error("offline"));
        const log = scope.run(() => useLogSet())!;
        const del = scope.run(() => useDeleteSet())!;
        log.mutate({ ...vars("s1"), exercise_id: SQUAT });
        del.mutate({ client_id: "a", date: DATE, exercise_id: BENCH });
        await vi.advanceTimersByTimeAsync(10_000);
        const day1 = qc.getQueryData<WorkoutLogsPreviousResponse>(KEY);
        // A snapshot restore would have wiped the squat row; the revert puts back only "a".
        expect(sets().map(setKey)).toEqual(["a", "b"]);
        expect(findGroup(day1, SQUAT)!.logged!.sets.map(setKey)).toEqual(["s1"]);
        // Settle it: scope chains are module-level and would block later tests.
        create.reject(new Error("done"));
        await vi.advanceTimersByTimeAsync(10_000);
    });
});

describe("useSaveExerciseNotes", () => {
    it("waits for the set create, then full-replaces with the server's sets", async () => {
        qc.setQueryData(KEY, day());
        const create = deferred<unknown>();
        http.apiPOST.mockReturnValueOnce(create.promise);
        http.apiPOST.mockResolvedValueOnce({ exercise: logged([set(11, "c1")], { ID: 77, notes: "felt good" }) });
        http.apiGET.mockResolvedValue(day(logged([set(11, "c1")], { ID: 77 })));
        const log = scope.run(() => useLogSet())!;
        const notes = scope.run(() => useSaveExerciseNotes())!;
        log.mutate(vars("c1"));
        notes.mutate({ date: DATE, exercise_id: BENCH, notes: "felt good" });
        await flush();
        expect(http.apiPOST).toHaveBeenCalledTimes(1);
        create.resolve({ exercise: logged([set(11, "c1")], { ID: 77 }), set: set(11, "c1"), created: true, first_set_of_day: false });
        await flush();
        expect(http.apiPOST).toHaveBeenCalledTimes(2);
        const [url, body] = http.apiPOST.mock.calls[1];
        expect(url).toBe("/workout/exercises/log");
        expect(body.type).toBe("logged");
        expect(body.exercise).toMatchObject({ ID: 77, notes: "felt good" });
        expect(body.exercise.sets).toEqual([expect.objectContaining({ ID: 11 })]);
        expect(body.exercise.sets[0]).not.toHaveProperty("client_id");
        expect(findGroup(qc.getQueryData(KEY), BENCH)!.logged!.notes).toBe("felt good");
    });

    it("excludes sets that are not saved yet from the full replace", async () => {
        qc.setQueryData(KEY, day(logged([set(1, "a"), set(0, "local")])));
        http.apiGET.mockResolvedValue(day(logged([set(1, "a")])));
        http.apiPOST.mockResolvedValue({ exercise: logged([set(1, "a")], { notes: "n" }) });
        const notes = scope.run(() => useSaveExerciseNotes())!;
        notes.mutate({ date: DATE, exercise_id: BENCH, notes: "n" });
        await flush();
        expect(http.apiPOST.mock.calls[0][1].exercise.sets.map((s: LoggedSet) => s.ID)).toEqual([1]);
        // The unsaved local row survives the reconcile.
        expect(sets().map(setKey)).toEqual(["a", "local"]);
    });

    it("rev-g2 #4: sends the server's sets, not the cache's, so the replace never deletes or edits a set", async () => {
        // Cache: "a" carries a failed edit (reps 99); "lost" is a create whose answer was lost (ID 0)
        // but which the server committed; "other" was logged from another device.
        qc.setQueryData(KEY, day(logged([set(1, "a", 99), set(0, "lost")])));
        http.apiGET.mockResolvedValue(day(logged([set(1, "a", 8), set(2, "lost"), set(3, "other")])));
        http.apiPOST.mockResolvedValue({ exercise: logged([set(1, "a", 8), set(2, "lost"), set(3, "other")], { notes: "n" }) });
        const notes = scope.run(() => useSaveExerciseNotes())!;
        notes.mutate({ date: DATE, exercise_id: BENCH, notes: "n" });
        await flush();
        expect(http.apiGET).toHaveBeenCalledWith("/workout/logs/previous", { params: { offset: expect.any(Number), date: DATE } });
        const sent = http.apiPOST.mock.calls[0][1].exercise.sets as LoggedSet[];
        expect(sent.map((s) => [s.ID, s.reps])).toEqual([
            [1, 8],
            [2, 8],
            [3, 8],
        ]);
    });

    it("rev-g2 #4: a notes save queued before a delete keeps the set; the delete then removes the parent", async () => {
        qc.setQueryData(KEY, day(logged([set(1, "a")])));
        const notesPost = deferred<unknown>();
        http.apiGET.mockResolvedValue(day(logged([set(1, "a")])));
        http.apiPOST.mockReturnValue(notesPost.promise);
        http.apiDELETE.mockResolvedValue({ success: true, deleted: true, logged_exercise_id: 50, exercise_removed: true });
        const notes = scope.run(() => useSaveExerciseNotes())!;
        const del = scope.run(() => useDeleteSet())!;
        notes.mutate({ date: DATE, exercise_id: BENCH, notes: "n" });
        del.mutate({ client_id: "a", date: DATE, exercise_id: BENCH });
        await flush();
        expect(http.apiPOST.mock.calls[0][1].exercise.sets.map((s: LoggedSet) => s.ID)).toEqual([1]);
        expect(http.apiDELETE).not.toHaveBeenCalled();
        notesPost.resolve({ exercise: logged([set(1, "a")], { notes: "n" }) });
        await flush();
        expect(http.apiDELETE).toHaveBeenCalledTimes(1);
        expect(findGroup(qc.getQueryData(KEY), BENCH)?.logged).toBeUndefined();
    });
});

describe("rev-g2 fixes", () => {
    it("#5: a replayed create (created: false) refreshes the activity heatmap", async () => {
        qc.setQueryData(KEY, day());
        const spy = vi.spyOn(qc, "invalidateQueries");
        http.apiPOST.mockResolvedValue({ exercise: logged([set(11, "c1")]), set: set(11, "c1"), created: false, first_set_of_day: false });
        const m = scope.run(() => useLogSet())!;
        m.mutate(vars("c1"));
        await flush();
        expect(spy).toHaveBeenCalledWith({ queryKey: workoutKeys.activity.all });
    });

    it("#6: a 4xx is not retried, so later writes in the scope don't wait on backoff", async () => {
        qc.setQueryData(KEY, day(logged([set(1, "a"), set(2, "b")])));
        http.apiPATCH.mockRejectedValueOnce(httpError(404));
        http.apiPATCH.mockResolvedValueOnce({ exercise: logged([set(1, "a"), set(2, "b", 9)]), set: set(2, "b", 9) });
        const m = scope.run(() => useUpdateSet())!;
        const intents = scope.run(() => useSetIntents())!;
        m.mutate({ client_id: "a", date: DATE, exercise_id: BENCH, weight: 100, reps: 5, weight_setup: "" });
        m.mutate({ client_id: "b", date: DATE, exercise_id: BENCH, weight: 100, reps: 9, weight_setup: "" });
        await flush();
        expect(http.apiPATCH).toHaveBeenCalledTimes(2); // no backoff before the second edit
        expect(intents.value.failedUpdate.has("a")).toBe(true);
    });

    it("#6: transient errors still retry twice", async () => {
        expect(retrySetWrite(0, new Error("offline"))).toBe(true);
        expect(retrySetWrite(1, httpError(503))).toBe(true);
        expect(retrySetWrite(2, httpError(503))).toBe(false);
        for (const status of [400, 404, 409]) expect(retrySetWrite(0, httpError(status))).toBe(false);
    });

    it("#7: the day query names the day by date, as the set POST does", async () => {
        http.apiGET.mockResolvedValue(day());
        await qc.fetchQuery(workoutQueries.day(DATE));
        expect(http.apiGET).toHaveBeenCalledWith("/workout/logs/previous", { params: expect.objectContaining({ date: DATE }) });
    });
});
