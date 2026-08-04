import { computed, ref } from "vue";

type StoredGlobalTimer = {
    startedAt: number;
    exerciseName: string;
};

const STORAGE_KEY = "gym-rest-timer:global";
const TICK_MS = 250;
const MAX_REST_MS = 5 * 60 * 1000;

const startedAt = ref<number | null>(null);
const timerExerciseName = ref("");
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

function start(name: string) {
    startedAt.value = Date.now();
    timerExerciseName.value = name;
    persist();
    tick();
    ensureTick();
}

function clear() {
    startedAt.value = null;
    timerExerciseName.value = "";
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
        start,
        clear,
    };
}
