<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";

type StoredRestTimer = {
    endsAt: number;
};

const props = withDefaults(
    defineProps<{
        storageKey: string;
        startToken: number;
        clearToken: number;
        durationMs?: number;
        fallbackText: string;
    }>(),
    {
        durationMs: 120000,
    },
);

const remainingMs = ref(0);
const restEndsAt = ref<number | null>(null);

let intervalId: number | null = null;

const clearTick = () => {
    if (intervalId !== null && typeof window !== "undefined") {
        window.clearInterval(intervalId);
    }
    intervalId = null;
};

const removeStoredTimer = (key: string) => {
    if (typeof window === "undefined" || !key) return;
    window.localStorage.removeItem(key);
};

const saveStoredTimer = (key: string, endsAt: number) => {
    if (typeof window === "undefined" || !key) return;
    const payload: StoredRestTimer = { endsAt };
    window.localStorage.setItem(key, JSON.stringify(payload));
};

const getStoredTimer = (key: string): StoredRestTimer | null => {
    if (typeof window === "undefined" || !key) return null;

    const raw = window.localStorage.getItem(key);
    if (!raw) return null;

    try {
        const parsed = JSON.parse(raw) as Partial<StoredRestTimer>;
        if (typeof parsed.endsAt !== "number" || !Number.isFinite(parsed.endsAt)) {
            removeStoredTimer(key);
            return null;
        }
        return { endsAt: parsed.endsAt };
    } catch {
        removeStoredTimer(key);
        return null;
    }
};

const refreshRemaining = () => {
    if (restEndsAt.value === null) {
        remainingMs.value = 0;
        return;
    }

    const nextRemaining = Math.max(0, restEndsAt.value - Date.now());
    remainingMs.value = nextRemaining;

    if (nextRemaining <= 0) {
        const key = props.storageKey;
        restEndsAt.value = null;
        clearTick();
        removeStoredTimer(key);
    }
};

const ensureTick = () => {
    if (intervalId !== null || typeof window === "undefined") return;
    intervalId = window.setInterval(() => {
        refreshRemaining();
    }, 250);
};

const clearTimer = () => {
    const key = props.storageKey;
    restEndsAt.value = null;
    remainingMs.value = 0;
    clearTick();
    removeStoredTimer(key);
};

const startTimer = () => {
    const key = props.storageKey;
    const endsAt = Date.now() + props.durationMs;
    restEndsAt.value = endsAt;
    saveStoredTimer(key, endsAt);
    refreshRemaining();
    ensureTick();
};

const restoreTimer = () => {
    clearTick();
    remainingMs.value = 0;
    restEndsAt.value = null;

    const key = props.storageKey;
    const stored = getStoredTimer(key);
    if (!stored) return;

    if (stored.endsAt <= Date.now()) {
        removeStoredTimer(key);
        return;
    }

    restEndsAt.value = stored.endsAt;
    refreshRemaining();
    ensureTick();
};

const displayText = computed(() => {
    if (remainingMs.value <= 0) return props.fallbackText;

    const totalSeconds = Math.ceil(remainingMs.value / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
});

onMounted(() => {
    restoreTimer();
});

onBeforeUnmount(() => {
    clearTick();
});

watch(
    () => props.startToken,
    () => {
        startTimer();
    },
);

watch(
    () => props.clearToken,
    () => {
        clearTimer();
    },
);

watch(
    () => props.storageKey,
    () => {
        restoreTimer();
    },
);
</script>

<template>
    <span>{{ displayText }}</span>
</template>
