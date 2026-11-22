import type { QueryClient } from "@tanstack/vue-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, type EffectScope } from "vue";

import { createQueryClient } from "~/api/queryClient";
import type { WorkoutLogsPreviousResponse } from "~/api/workout/api";
import type { Exercise, LoggedExercise } from "~/types/workout";
import { workoutKeys } from "./keys";

const http = vi.hoisted(() => ({ apiGET: vi.fn(), apiPOST: vi.fn(), apiPATCH: vi.fn(), apiDELETE: vi.fn() }));
vi.mock("~/api/client", () => http);
const push = vi.hoisted(() => vi.fn());
vi.mock("~/composables/toast/useToast", () => ({ toast: { push } }));
const qcRef = vi.hoisted(() => ({ qc: null as unknown }));
vi.mock("@tanstack/vue-query", async (orig) => ({
    ...(await orig<typeof import("@tanstack/vue-query")>()),
    useQueryClient: () => qcRef.qc,
}));

import {
    useAddExerciseToWorkout,
    useRemoveExerciseFromWorkout,
    useSwitchWorkoutPlan,
    useUpsertCardio,
    useUpsertMobilityPre,
} from "./queries";

const DATE = "2026-11-22";
const KEY = workoutKeys.days.day(DATE);
const ex = (ID: number): Exercise => ({ ID, name: `E${ID}`, rep_rollover: 10, cues: "", created_at: "", updated_at: "" });
const lg = (exercise_id: number, ID = 90): LoggedExercise => ({
    ID,
    workout_log_id: 3,
    exercise_id,
    exercise: ex(exercise_id),
    sets: [],
    notes: "",
    created_at: "",
    updated_at: "",
});
const day = (over: Partial<WorkoutLogsPreviousResponse> = {}): WorkoutLogsPreviousResponse => ({
    day: { ID: 3 } as WorkoutLogsPreviousResponse["day"],
    planned_exercises: [],
    planned_cardio: null,
    logged_cardio: null,
    planned_pre_mobility: { title: "Pre", items: ["a", "b"] },
    logged_pre_mobility: null,
    planned_post_mobility: null,
    logged_post_mobility: null,
    ...over,
});
const get = () => qc.getQueryData<WorkoutLogsPreviousResponse>(KEY)!;
const flush = () => vi.advanceTimersByTimeAsync(0);
const flushRetries = () => vi.advanceTimersByTimeAsync(10_000);

function deferred<T>() {
    let resolve!: (v: T) => void;
    let reject!: (e: unknown) => void;
    const promise = new Promise<T>((res, rej) => {
        resolve = res;
        reject = rej;
    });
    return { promise, resolve, reject };
}

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

describe("useAddExerciseToWorkout", () => {
    it("shows the group at once, then swaps in the saved row", async () => {
        qc.setQueryData(KEY, day());
        const d = deferred<unknown>();
        http.apiPOST.mockReturnValue(d.promise);
        const m = scope.run(() => useAddExerciseToWorkout())!;
        m.mutate({ date: DATE, exerciseId: 5, exercise: ex(5) });
        await flush();
        expect(get().planned_exercises[0].logged).toMatchObject({ ID: 0, exercise_id: 5 });
        d.resolve({ exercise: lg(5, 91) });
        await flush();
        expect(get().planned_exercises[0].logged?.ID).toBe(91);
    });
    it("removes the group and toasts with Retry on failure", async () => {
        qc.setQueryData(KEY, day());
        http.apiPOST.mockRejectedValue(new Error("boom"));
        const m = scope.run(() => useAddExerciseToWorkout())!;
        m.mutate({ date: DATE, exerciseId: 5, exercise: ex(5) });
        await flush();
        expect(get().planned_exercises).toEqual([]);
        expect(push).toHaveBeenCalledWith(expect.any(String), "error", expect.objectContaining({ action: expect.anything() }));
    });
});

describe("useRemoveExerciseFromWorkout", () => {
    it("hides the group at once, restores it on failure", async () => {
        qc.setQueryData(KEY, day({ planned_exercises: [{ logged: lg(5) }] }));
        const d = deferred<unknown>();
        http.apiDELETE.mockReturnValue(d.promise);
        const m = scope.run(() => useRemoveExerciseFromWorkout())!;
        m.mutate({ date: DATE, exerciseId: 5 });
        await flush();
        expect(get().planned_exercises).toEqual([]);
        d.reject(new Error("boom"));
        await flushRetries();
        expect(get().planned_exercises).toEqual([{ logged: lg(5) }]);
    });
    it("treats a 404 as success", async () => {
        qc.setQueryData(KEY, day({ planned_exercises: [{ logged: lg(5) }] }));
        const err = Object.assign(new Error("nf"), { isAxiosError: true, response: { status: 404 } });
        http.apiDELETE.mockRejectedValue(err);
        const m = scope.run(() => useRemoveExerciseFromWorkout())!;
        m.mutate({ date: DATE, exerciseId: 5 });
        await flush();
        expect(push).not.toHaveBeenCalled();
        expect(get().planned_exercises).toEqual([]);
    });
});

describe("mobility and cardio", () => {
    it("ticks mobility at once and rolls back on failure", async () => {
        qc.setQueryData(KEY, day());
        http.apiPOST.mockRejectedValue(new Error("boom"));
        const m = scope.run(() => useUpsertMobilityPre())!;
        m.mutate({ date: DATE, checked: ["a"] });
        await flushRetries();
        expect(http.apiPOST).toHaveBeenCalledWith("/workout/logs/mobility/pre", { checked: ["a"] }, { params: { offset: 0, date: DATE } });
        expect(get().logged_pre_mobility).toBeNull();
    });
    it("keeps the optimistic tick and takes the response when it lands", async () => {
        qc.setQueryData(KEY, day());
        http.apiPOST.mockResolvedValue({ mobility: { title: "Pre", items: ["a", "b"], checked: ["a"] } });
        const m = scope.run(() => useUpsertMobilityPre())!;
        m.mutate({ date: DATE, checked: ["a"] });
        await flush();
        expect(get().logged_pre_mobility?.checked).toEqual(["a"]);
    });
    it("shows cardio at once and sends the same body as before", async () => {
        qc.setQueryData(KEY, day({ planned_cardio: { type: "Run", minutes: 20 } }));
        const d = deferred<unknown>();
        http.apiPOST.mockReturnValue(d.promise);
        const m = scope.run(() => useUpsertCardio())!;
        m.mutate({ date: DATE, minutes: 25, notes: "n" });
        await flush();
        expect(get().logged_cardio).toMatchObject({ minutes: 25, type: "Run" });
        expect(http.apiPOST).toHaveBeenCalledWith("/workout/logs/cardio", { minutes: 25, notes: "n" }, { params: { offset: 0, date: DATE } });
        d.resolve({ cardio: { ID: 9, workout_log_id: 3, minutes: 25, type: "Run", notes: "n", created_at: "", updated_at: "" } });
        await flush();
        expect(get().logged_cardio?.ID).toBe(9);
    });
});

describe("useSwitchWorkoutPlan", () => {
    it("writes the response into the day and does not refetch it", async () => {
        qc.setQueryData(KEY, day());
        const spy = vi.spyOn(qc, "invalidateQueries");
        const res = day({ planned_exercises: [{ planned: ex(1) }] });
        http.apiPATCH.mockResolvedValue(res);
        const m = scope.run(() => useSwitchWorkoutPlan())!;
        await m.mutateAsync({ date: DATE, planId: 4 });
        expect(get()).toEqual(res);
        expect(spy).not.toHaveBeenCalledWith(expect.objectContaining({ queryKey: KEY }));
    });
});
