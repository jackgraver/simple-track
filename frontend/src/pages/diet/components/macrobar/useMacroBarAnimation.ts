import { onBeforeUnmount, ref, watch } from "vue";

function easeOutCubic(t: number): number {
    return 1 - Math.pow(1 - t, 3);
}

const DURATION_MS = 380;

export function useMacroBarAnimation(getTotal: () => number) {
    const displayTotal = ref(getTotal() ?? 0);
    let rafId: number | null = null;

    function animateTo(targetTotal: number) {
        if (rafId != null) {
            cancelAnimationFrame(rafId);
            rafId = null;
        }
        const fromTotal = displayTotal.value;
        if (fromTotal === targetTotal) return;

        const start = performance.now();
        const step = (now: number) => {
            const elapsed = now - start;
            const t = Math.min(1, elapsed / DURATION_MS);
            const e = easeOutCubic(t);
            displayTotal.value = fromTotal + (targetTotal - fromTotal) * e;
            if (t < 1) {
                rafId = requestAnimationFrame(step);
            } else {
                displayTotal.value = targetTotal;
                rafId = null;
            }
        };
        rafId = requestAnimationFrame(step);
    }

    watch(
        () => getTotal() ?? 0,
        (t) => {
            animateTo(t);
        },
    );

    onBeforeUnmount(() => {
        if (rafId != null) cancelAnimationFrame(rafId);
    });

    return { displayTotal };
}
