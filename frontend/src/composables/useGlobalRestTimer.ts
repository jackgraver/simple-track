import { computed, ref } from "vue";

type StoredGlobalTimer = {
    endsAt: number;
    exerciseName: string;
};

const STORAGE_KEY = "gym-rest-timer:global";
const TICK_MS = 250;

const endsAt = ref<number | null>(null);
const timerExerciseName = ref("");
const remainingMs = ref(0);

let intervalId: number | null = null;

function persist() {
    if (endsAt.value === null) {
        window.localStorage.removeItem(STORAGE_KEY);
        return;
    }
    const payload: StoredGlobalTimer = {
        endsAt: endsAt.value,
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
    if (endsAt.value === null) {
        remainingMs.value = 0;
        return;
    }
    const next = Math.max(0, endsAt.value - Date.now());
    remainingMs.value = next;
    if (next <= 0) {
        endsAt.value = null;
        timerExerciseName.value = "";
        clearTick();
        persist();
    }
}

function ensureTick() {
    if (intervalId !== null) return;
    intervalId = window.setInterval(tick, TICK_MS);
}

function start(durationMs: number, name: string) {
    endsAt.value = Date.now() + durationMs;
    timerExerciseName.value = name;
    persist();
    tick();
    ensureTick();
}

function clear() {
    endsAt.value = null;
    timerExerciseName.value = "";
    remainingMs.value = 0;
    clearTick();
    persist();
}

function restore() {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return;
    try {
        const parsed = JSON.parse(raw) as Partial<StoredGlobalTimer>;
        if (
            typeof parsed.endsAt !== "number" ||
            !Number.isFinite(parsed.endsAt)
        ) {
            window.localStorage.removeItem(STORAGE_KEY);
            return;
        }
        if (parsed.endsAt <= Date.now()) {
            window.localStorage.removeItem(STORAGE_KEY);
            return;
        }
        endsAt.value = parsed.endsAt;
        timerExerciseName.value = parsed.exerciseName ?? "";
        tick();
        ensureTick();
    } catch {
        window.localStorage.removeItem(STORAGE_KEY);
    }
}

restore();

const isActive = computed(() => remainingMs.value > 0);

const displayText = computed(() => {
    if (remainingMs.value <= 0) return "";
    const totalSeconds = Math.ceil(remainingMs.value / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
});

export function useGlobalRestTimer() {
    return {
        remainingMs,
        isActive,
        displayText,
        exerciseName: timerExerciseName,
        start,
        clear,
    };
}
