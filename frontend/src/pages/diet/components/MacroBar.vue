<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from "vue";

function formatInt(n: number): string {
    return String(Math.round(n));
}

function calcWidth(total: number, planned: number): number {
    if (planned <= 0) {
        return total > 0 ? 100 : 0;
    }
    return Math.min(100, (total / planned) * 100);
}

function determineOverflow(total: number, planned: number): string {
    if (planned <= 0 || !props.indicateOverflow) return "";
    const overflow = (total / planned) * 100 - 100;

    if (overflow > 20) return "large-overflow";
    if (overflow > 10) return "medium-overflow";
    if (overflow > 0) return "small-overflow";
    return "";
}

const props = defineProps<{
    total: number;
    planned: number;
    type: "calories" | "protein" | "fiber" | "carbs";
    indicateOverflow?: boolean;
}>();

/** Ease-out cubic: quick start, settles gently at the end */
function easeOutCubic(t: number): number {
    return 1 - Math.pow(1 - t, 3);
}

const DURATION_MS = 380;
const displayTotal = ref(props.total ?? 0);
const displayPlanned = ref(props.planned ?? 0);

let rafId: number | null = null;

function animateTo(targetTotal: number, targetPlanned: number) {
    if (rafId != null) {
        cancelAnimationFrame(rafId);
        rafId = null;
    }
    const fromTotal = displayTotal.value;
    const fromPlanned = displayPlanned.value;
    if (fromTotal === targetTotal && fromPlanned === targetPlanned) return;

    const start = performance.now();
    const step = (now: number) => {
        const elapsed = now - start;
        const t = Math.min(1, elapsed / DURATION_MS);
        const e = easeOutCubic(t);
        displayTotal.value = fromTotal + (targetTotal - fromTotal) * e;
        displayPlanned.value = fromPlanned + (targetPlanned - fromPlanned) * e;
        if (t < 1) {
            rafId = requestAnimationFrame(step);
        } else {
            displayTotal.value = targetTotal;
            displayPlanned.value = targetPlanned;
            rafId = null;
        }
    };
    rafId = requestAnimationFrame(step);
}

watch(
    () => [props.total ?? 0, props.planned ?? 0] as const,
    ([t, p]) => {
        animateTo(t, p);
    },
);

onBeforeUnmount(() => {
    if (rafId != null) cancelAnimationFrame(rafId);
});
</script>

<template>
    <div class="fill-container">
        <div
            class="fill"
            :class="type"
            :style="{
                width: `${calcWidth(displayTotal, displayPlanned)}%`,
            }"
        >
            <span
                class="tabular-nums"
                :class="
                    determineOverflow(
                        Math.round(displayTotal),
                        Math.round(displayPlanned),
                    )
                "
                >{{
                    formatInt(displayTotal) + " / " + formatInt(displayPlanned)
                }}</span
            >
        </div>
    </div>
</template>

<style scoped>
.fill-container {
    flex: 1;
    border-radius: 4px;
    border: 1px solid;
    border-color: var(--color-secondBg);
}

.fill {
    height: 100%;
    text-align: center;
    line-height: 20px;
    border-radius: 4px;
}

.fill span {
    color: #ffffff;
    font-weight: bold;
    white-space: nowrap;
    padding: 0 8px;
}

.calories {
    background-color: orange;
}
.protein {
    background-color: #60a5fa;
}
.fiber {
    background-color: green;
}
.carbs {
    background-color: red;
}

.small-overflow {
    color: yellow;
}
.medium-overflow {
    color: orange;
}
.large-overflow {
    color: red;
}
</style>
