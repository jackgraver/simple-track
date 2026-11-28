import { computed, ref, watch } from "vue";
import { currentUserId, userKey } from "~/utils/userStorage";
import type { SetSide } from "~/types/workout";

type StoredGlobalTimer = {
    startedAt: number;
    exerciseName: string;
    /** Unilateral: the first side was just logged and this side is next (the short switch wait). */
    switchTo?: SetSide | null;
};

const STORAGE_NAME = "gym-rest-timer:global";
const TICK_MS = 250;
const MAX_REST_MS = 5 * 60 * 1000;

const startedAt = ref<number | null>(null);
const timerExerciseName = ref("");
const switchTo = ref<SetSide | null>(null);
const elapsedMs = ref(0);

let intervalId: number | null = null;

/** Null until the signed-in user is known: nothing is read or written then. */
const storageKey = () => userKey(STORAGE_NAME);

function persist() {
    const key = storageKey();
    if (key === null) return;
    if (startedAt.value === null) {
        window.localStorage.removeItem(key);
        return;
    }
    const payload: StoredGlobalTimer = {
        startedAt: startedAt.value,
        exerciseName: timerExerciseName.value,
        switchTo: switchTo.value,
    };
    window.localStorage.setItem(key, JSON.stringify(payload));
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
 * measured from it), just leave the switch wait. Restarts the clock instead when it isn't waiting on
 * this exercise's switch (stopped, restarted since, or another exercise's).
 */
function endSwitch(name: string) {
    if (startedAt.value === null || switchTo.value === null || timerExerciseName.value !== name) {
        start(name);
        return;
    }
    switchTo.value = null;
    persist();
}

/** The first side this switch wait was for was deleted: leave the switch wait, keep counting. */
function cancelSwitch(name: string) {
    if (switchTo.value === null || timerExerciseName.value !== name) return;
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
    const key = storageKey();
    if (key === null) return;
    const raw = window.localStorage.getItem(key);
    if (!raw) return;
    try {
        const parsed = JSON.parse(raw) as Partial<StoredGlobalTimer>;
        if (
            typeof parsed.startedAt !== "number" ||
            !Number.isFinite(parsed.startedAt)
        ) {
            window.localStorage.removeItem(key);
            return;
        }
        if (parsed.startedAt > Date.now()) {
            window.localStorage.removeItem(key);
            return;
        }
        startedAt.value = parsed.startedAt;
        timerExerciseName.value = parsed.exerciseName ?? "";
        switchTo.value = parsed.switchTo === "left" || parsed.switchTo === "right" ? parsed.switchTo : null;
        tick();
        ensureTick();
    } catch {
        window.localStorage.removeItem(key);
    }
}

// The stored timer belongs to a user: drop the in-memory one when the user changes or signs out,
// and restore theirs once their ID is known (it isn't at import time).
watch(
    currentUserId,
    () => {
        startedAt.value = null;
        timerExerciseName.value = "";
        switchTo.value = null;
        elapsedMs.value = 0;
        clearTick();
        restore();
    },
    { immediate: true },
);

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
        cancelSwitch,
        clear,
    };
}
