import { computed, ref } from "vue";
import type { SetSide } from "~/types/workout";

type StoredGlobalTimer = {
    startedAt: number;
    exerciseName: string;
    /** Unilateral: the first side was just logged and this side is next (the short switch wait). */
    switchTo?: SetSide | null;
};

const STORAGE_KEY = "gym-rest-timer:global";
const TICK_MS = 250;
const MAX_REST_MS = 5 * 60 * 1000;

const startedAt = ref<number | null>(null);
const timerExerciseName = ref("");
const switchTo = ref<SetSide | null>(null);
const elapsedMs = ref(0);

let intervalId: number | null = null;

function persist() {
    if (startedAt.value === null) {
        window.localStorage.removeItem(STORAGE_KEY);
        return;
    }
    const payload: StoredGlobalTimer = {
        startedAt: startedAt.value,
        exerciseName: timerExerciseName.value,
        switchTo: switchTo.value,
    };
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
}

function clearTick() {
    if (intervalId !== null) {
        window.clearInterval(intervalId);
        intervalId = null;
    }
}

function tick() {
    if (startedAt.value === null) {
        elapsedMs.value = 0;
        return;
    }
    const elapsed = Math.max(0, Date.now() - startedAt.value);
    if (elapsed >= MAX_REST_MS) {
        clear();
        return;
    }
    elapsedMs.value = elapsed;
}

function ensureTick() {
    if (intervalId !== null) return;
    intervalId = window.setInterval(tick, TICK_MS);
}

/** Restart the clock from now. `switchSide`: a unilateral first side was logged, that side is next. */
function start(name: string, switchSide: SetSide | null = null) {
    startedAt.value = Date.now();
    timerExerciseName.value = name;
    switchTo.value = switchSide;
    persist();
    tick();
    ensureTick();
}

/**
 * The second side of a unilateral set was logged: keep counting from the first side (the rest is
 * measured from it), just leave the switch wait. Starts the clock if it isn't running.
 */
function endSwitch(name: string) {
    if (startedAt.value === null) {
        start(name);
        return;
    }
    switchTo.value = null;
    persist();
}

function clear() {
    startedAt.value = null;
    timerExerciseName.value = "";
    switchTo.value = null;
    elapsedMs.value = 0;
    clearTick();
    persist();
}

function restore() {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return;
    try {
        const parsed = JSON.parse(raw) as Partial<StoredGlobalTimer>;
        if (
            typeof parsed.startedAt !== "number" ||
            !Number.isFinite(parsed.startedAt)
        ) {
            window.localStorage.removeItem(STORAGE_KEY);
            return;
        }
        if (parsed.startedAt > Date.now()) {
            window.localStorage.removeItem(STORAGE_KEY);
            return;
        }
        startedAt.value = parsed.startedAt;
        timerExerciseName.value = parsed.exerciseName ?? "";
        switchTo.value = parsed.switchTo === "left" || parsed.switchTo === "right" ? parsed.switchTo : null;
        tick();
        ensureTick();
    } catch {
        window.localStorage.removeItem(STORAGE_KEY);
    }
}

restore();

const isActive = computed(() => startedAt.value !== null);

const displayText = computed(() => {
    if (elapsedMs.value <= 0) return "0:00";
    const totalSeconds = Math.floor(elapsedMs.value / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
});

export function useGlobalRestTimer() {
    return {
        elapsedMs,
        isActive,
        displayText,
        exerciseName: timerExerciseName,
        /** Set while waiting to switch sides on a unilateral exercise: the side that's next. */
        switchTo,
        start,
        endSwitch,
        clear,
    };
}
