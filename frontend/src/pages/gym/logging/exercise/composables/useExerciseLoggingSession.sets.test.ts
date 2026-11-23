/**
 * The logging session against the real set hooks and a real QueryClient (HTTP mocked): repros for
 * the G2 review (rev-g2 #1-#3), where the bug only shows in how the session drives the hooks.
 */
import type { QueryClient } from "@tanstack/vue-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { computed, effectScope, nextTick, ref, shallowRef, type EffectScope } from "vue";
import type { Router } from "vue-router";

import { createQueryClient } from "~/api/queryClient";
import type { WorkoutLogsPreviousResponse } from "~/api/workout/api";
import { workoutKeys } from "~/api/workout/keys";
import { findGroup } from "~/api/workout/setsCache";
import type { Exercise, LoggedExercise, LoggedSet } from "~/types/workout";

const http = vi.hoisted(() => ({ apiGET: vi.fn(), apiPOST: vi.fn(), apiPATCH: vi.fn(), apiDELETE: vi.fn() }));
vi.mock("~/api/client", () => http);
const push = vi.hoisted(() => vi.fn());
vi.mock("~/composables/toast/useToast", () => ({ toast: { push } }));
const qcRef = vi.hoisted(() => ({ qc: null as unknown }));
vi.mock("@tanstack/vue-query", async (orig) => ({
    ...(await orig<typeof import("@tanstack/vue-query")>()),
    useQueryClient: () => qcRef.qc,
}));
vi.mock("~/composables/useGlobalRestTimer", () => ({ useGlobalRestTimer: () => ({ start: vi.fn() }) }));
vi.mock("./useGymWeightPrefs", async () => {
    const { ref } = await import("vue");
    return { useGymWeightPrefs: () => ({ weightIncrement: ref(2.5), setWeightIncrement: vi.fn() }) };
});
vi.mock("~/composables/useWebStorageJsonSync", () => ({
    useWebStorageJsonSync: () => ({ restore: () => false, setSaveEnabled: vi.fn(), clear: vi.fn() }),
}));

import { useExerciseLoggingSession } from "./useExerciseLoggingSession";

const DATE = "2026-11-22";
const BENCH = 7;
const KEY = workoutKeys.days.day(DATE);

const exercise: Exercise = { ID: BENCH, name: "Bench", rep_rollover: 10, cues: "", created_at: "", updated_at: "" };
const set = (ID: number, client_id: string | null, reps = 8): LoggedSet => ({
    ID,
    client_id,
    logged_exercise_id: 77,
    weight: 100,
    reps,
    weight_setup: "",
    created_at: "t",
    updated_at: "t",
});
const logged = (sets: LoggedSet[], extra: Partial<LoggedExercise> = {}): LoggedExercise => ({
    ID: 77,
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
const deleteOk = { success: true, deleted: false, logged_exercise_id: 77, exercise_removed: false };

let qc: QueryClient;
let scope: EffectScope;
let unsubscribe: () => void;

function openSession() {
    // The page reads the group from the day query; mirror the cache into a ref the same way.
    const group = shallowRef(findGroup(qc.getQueryData(KEY), BENCH) ?? null);
    unsubscribe = qc.getQueryCache().subscribe(() => {
        group.value = findGroup(qc.getQueryData(KEY), BENCH) ?? null;
    });
    return scope.run(() =>
        useExerciseLoggingSession({
            exerciseGroup: computed(() => group.value),
            pending: ref(false),
            offset: computed(() => 0),
            router: { push: vi.fn() } as unknown as Router,
        }),
    )!;
}

beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 10, 22, 12, 0, 0));
    vi.clearAllMocks();
    http.apiGET.mockReset();
    http.apiPOST.mockReset();
    http.apiDELETE.mockReset();
    vi.stubGlobal("window", { sessionStorage: {}, localStorage: {} });
    qc = createQueryClient(vi.fn());
    qcRef.qc = qc;
    scope = effectScope();
});
afterEach(() => {
    unsubscribe?.();
    scope.stop();
    vi.unstubAllGlobals();
    vi.useRealTimers();
});

describe("deleting sets whose create failed", () => {
    it("#1: a pending set deleted before its create fails stays deleted", async () => {
        qc.setQueryData(KEY, day(logged([set(11, "s0")])));
        http.apiPOST.mockRejectedValue(new Error("offline"));
        http.apiDELETE.mockResolvedValue(deleteOk);
        const s = openSession();

        expect(s.addNextSet()).toBe(true);
        await vi.advanceTimersByTimeAsync(0);
        expect(s.loggedSets.map((r) => r.status)).toEqual(["success", "pending"]);
        s.deleteSet(1); // tapped while the POST is still spinning

        await vi.advanceTimersByTimeAsync(10_000); // create fails (with retries), then the DELETE runs
        expect(http.apiDELETE).toHaveBeenCalledTimes(1);
        expect(s.loggedSets.map((r) => r.key)).toEqual(["s0"]);
        expect(s.finishLogging()).toBe(true);
    });

    it("#2: a failed delete of a never-saved row leaves it marked \"Didn't save\"", async () => {
        qc.setQueryData(KEY, day(logged([set(11, "s0")])));
        http.apiPOST.mockRejectedValue(new Error("offline"));
        http.apiDELETE.mockRejectedValue(new Error("offline"));
        const s = openSession();

        s.addNextSet();
        await vi.advanceTimersByTimeAsync(10_000);
        expect(s.loggedSets[1]).toMatchObject({ status: "error", failed: "create" });

        s.deleteSet(1);
        await vi.advanceTimersByTimeAsync(0);
        expect(s.loggedSets.map((r) => r.key)).toEqual(["s0"]); // hidden while the delete runs
        await vi.advanceTimersByTimeAsync(10_000);

        expect(s.loggedSets[1]).toMatchObject({ status: "error", failed: "create", id: null });
        expect(s.finishLogging()).toBe(false);
    });

    it("a successful delete of a failed row clears it for good", async () => {
        qc.setQueryData(KEY, day(logged([set(11, "s0")])));
        http.apiPOST.mockRejectedValue(new Error("offline"));
        http.apiDELETE.mockResolvedValue(deleteOk);
        const s = openSession();

        s.addNextSet();
        await vi.advanceTimersByTimeAsync(10_000);
        s.deleteSet(1);
        await vi.advanceTimersByTimeAsync(0);
        expect(s.loggedSets.map((r) => r.key)).toEqual(["s0"]);
        qc.setQueryData(KEY, day(logged([set(11, "s0")]))); // a later refetch
        expect(s.loggedSets.map((r) => r.key)).toEqual(["s0"]);
    });
});

describe("notes", () => {
    it("#3: notes typed before the first set are saved after that set's create", async () => {
        qc.setQueryData(KEY, day());
        const s = openSession();
        s.updateNotes("seat 4");
        await vi.advanceTimersByTimeAsync(0);
        expect(http.apiPOST).not.toHaveBeenCalled(); // nothing to attach them to yet

        http.apiPOST.mockImplementation(async (url: string, body: { client_id?: string; exercise?: LoggedExercise }) => {
            if (url === "/workout/sets") {
                const saved = set(11, body.client_id!);
                return { exercise: logged([saved]), set: saved, created: true, first_set_of_day: true };
            }
            return { exercise: logged(body.exercise!.sets, { notes: body.exercise!.notes }) };
        });
        http.apiGET.mockImplementation(async () => {
            const g = findGroup(qc.getQueryData(KEY), BENCH)!.logged!;
            return day(logged(g.sets.filter((x) => x.ID > 0)));
        });
        s.commitRepsFromInput(8);
        expect(s.addNextSet()).toBe(true);
        s.goBackToList();
        await vi.advanceTimersByTimeAsync(0);

        const urls = http.apiPOST.mock.calls.map((c) => c[0]);
        expect(urls).toEqual(["/workout/sets", "/workout/exercises/log"]);
        expect(http.apiPOST.mock.calls[1][1].exercise).toMatchObject({ ID: 77, notes: "seat 4" });
        expect(findGroup(qc.getQueryData(KEY), BENCH)!.logged!.notes).toBe("seat 4");
    });
});

describe("reps default for the next set", () => {
    it("follows the previous set of this session, not last session's reps", async () => {
        const lastSession = logged([set(1, "p0", 12), set(2, "p1", 12), set(3, "p2", 12)]);
        qc.setQueryData(KEY, {
            ...day(),
            planned_exercises: [{ planned: exercise, logged: undefined, previous: lastSession }],
        });
        http.apiPOST.mockRejectedValue(new Error("offline"));
        const s = openSession();

        s.commitRepsFromInput(8);
        expect(s.addNextSet()).toBe(true);
        expect(s.currentReps).toBe(8); // not 12 from last session's set 2
        s.commitRepsFromInput(7);
        expect(s.addNextSet()).toBe(true);
        expect(s.currentReps).toBe(7);
    });

    it("a weight change after the first set keeps this session's reps", () => {
        const lastSession = logged([set(1, "p0", 12), { ...set(2, "p1", 12), weight: 110 }]);
        qc.setQueryData(KEY, {
            ...day(logged([set(11, "s0", 7)])),
            planned_exercises: [{ planned: exercise, logged: logged([set(11, "s0", 7)]), previous: lastSession }],
        });
        const s = openSession();
        expect(s.currentReps).toBe(7);
        s.commitWeightFromInput(110); // last session did 12 at 110 for set 2
        expect(s.currentReps).toBe(7);
    });

    it("saving a fix to an earlier set goes back to the last set's weight and reps", () => {
        qc.setQueryData(KEY, day(logged([{ ...set(11, "s0", 10), weight: 90 }, set(12, "s1", 8)])));
        http.apiPATCH.mockReturnValue(new Promise(() => {}));
        const s = openSession();
        s.editSet(0);
        expect(s.currentWeight).toBe(90);
        s.commitRepsFromInput(9);
        expect(s.addNextSet()).toBe(true);
        expect(s.currentWeight).toBe(100);
        expect(s.currentReps).toBe(8);
    });
});

describe("an exercise opened before its history is cached", () => {
    it("fills the draft from last session once the history arrives", async () => {
        qc.setQueryData(KEY, day(logged([])));
        const s = openSession();
        expect(s.currentWeight).toBe(0);

        const withHistory = day(logged([]));
        withHistory.planned_exercises![0]!.previous = logged([set(1, "p0", 6), set(2, "p1", 6)]);
        qc.setQueryData(KEY, withHistory);
        await nextTick();
        expect(s.currentWeight).toBe(100);
        expect(s.currentReps).toBe(6);
    });

    it("leaves the draft alone once the lifter has changed it", async () => {
        qc.setQueryData(KEY, day(logged([])));
        const s = openSession();
        s.commitRepsFromInput(5);

        const withHistory = day(logged([]));
        withHistory.planned_exercises![0]!.previous = logged([set(1, "p0", 6)]);
        qc.setQueryData(KEY, withHistory);
        await nextTick();
        expect(s.currentWeight).toBe(0);
        expect(s.currentReps).toBe(5);
    });
});
