export type ExerciseLoggingFlowStep = "setup" | "reps" | "rest";

export type ExerciseLoggingFlowAction =
    | "startLogging"
    | "setLogged"
    | "nextSet"
    | "backToSetup";

export type StoredExerciseLoggingFlow = {
    step: ExerciseLoggingFlowStep;
    setNumber: number | null;
};

export function parseStoredExerciseLoggingFlow(
    value: string | null,
): StoredExerciseLoggingFlow {
    if (value === "setup" || value === "reps" || value === "rest") {
        return { step: value, setNumber: null };
    }
    if (!value) return { step: "setup", setNumber: null };
    try {
        const parsed = JSON.parse(value) as {
            step?: unknown;
            setNumber?: unknown;
        };
        const step =
            parsed.step === "setup" ||
            parsed.step === "reps" ||
            parsed.step === "rest"
                ? parsed.step
                : "setup";
        const setNumber =
            typeof parsed.setNumber === "number" &&
            Number.isInteger(parsed.setNumber) &&
            parsed.setNumber > 0
                ? parsed.setNumber
                : null;
        return { step, setNumber };
    } catch {
        return { step: "setup", setNumber: null };
    }
}

export function transitionExerciseLoggingFlow(
    step: ExerciseLoggingFlowStep,
    action: ExerciseLoggingFlowAction,
): ExerciseLoggingFlowStep {
    if (step === "setup" && action === "startLogging") return "reps";
    if (step === "reps" && action === "setLogged") return "rest";
    if (step === "rest" && action === "nextSet") return "reps";
    if ((step === "reps" || step === "rest") && action === "backToSetup") {
        return "setup";
    }
    return step;
}
