import { describe, expect, it } from "vitest";
import {
    transitionExerciseLoggingFlow,
    type ExerciseLoggingFlowStep,
} from "./exerciseLoggingFlow";

describe("transitionExerciseLoggingFlow", () => {
    it("starts on setup and advances through the logging flow", () => {
        let step: ExerciseLoggingFlowStep = "setup";
        step = transitionExerciseLoggingFlow(step, "startLogging");
        expect(step).toBe("reps");
        step = transitionExerciseLoggingFlow(step, "setLogged");
        expect(step).toBe("rest");
        step = transitionExerciseLoggingFlow(step, "nextSet");
        expect(step).toBe("reps");
    });

    it("returns to setup from reps or rest", () => {
        expect(transitionExerciseLoggingFlow("reps", "backToSetup")).toBe(
            "setup",
        );
        expect(transitionExerciseLoggingFlow("rest", "backToSetup")).toBe(
            "setup",
        );
    });

    it("does not advance for actions that do not belong to the current screen", () => {
        expect(transitionExerciseLoggingFlow("setup", "setLogged")).toBe("setup");
        expect(transitionExerciseLoggingFlow("reps", "nextSet")).toBe("reps");
        expect(transitionExerciseLoggingFlow("rest", "startLogging")).toBe("rest");
        expect(transitionExerciseLoggingFlow("setup", "backToSetup")).toBe(
            "setup",
        );
    });
});
