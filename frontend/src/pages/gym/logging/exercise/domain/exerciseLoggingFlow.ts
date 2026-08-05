export type ExerciseLoggingFlowStep = "setup" | "reps" | "rest";

export type ExerciseLoggingFlowAction =
    | "startLogging"
    | "setLogged"
    | "nextSet"
    | "backToSetup";

export function transitionExerciseLoggingFlow(
    step: ExerciseLoggingFlowStep,
    action: ExerciseLoggingFlowAction,
): ExerciseLoggingFlowStep {
    if (step === "setup" && action === "startLogging") return "reps";
    if (step === "reps" && action === "setLogged") return "rest";
    if (step === "rest" && action === "nextSet") return "reps";
    if (
        (step === "reps" || step === "rest") &&
        action === "backToSetup"
    ) {
        return "setup";
    }
    return step;
}
