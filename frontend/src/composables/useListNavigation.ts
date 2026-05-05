import { ref, watch, onMounted, onUnmounted, type Ref } from "vue";

export function useListNavigation(
    itemCount: Ref<number>,
    globalActive: Ref<boolean>,
) {
    const activeIndex = ref(-1);

    watch(itemCount, () => {
        activeIndex.value = -1;
    });

    function handleKey(e: KeyboardEvent): boolean {
        const count = itemCount.value;
        if (count === 0) return false;

        if (e.key === "ArrowDown") {
            e.preventDefault();
            activeIndex.value =
                activeIndex.value < count - 1 ? activeIndex.value + 1 : 0;
            return true;
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            activeIndex.value =
                activeIndex.value > 0 ? activeIndex.value - 1 : count - 1;
            return true;
        }
        return false;
    }

    let onEnter: (() => void) | null = null;

    function setEnterHandler(fn: () => void) {
        onEnter = fn;
    }

    function onGlobalKeydown(e: KeyboardEvent) {
        if (!globalActive.value) return;
        if (e.key === "Enter") {
            e.preventDefault();
            onEnter?.();
            return;
        }
        handleKey(e);
    }

    onMounted(() => window.addEventListener("keydown", onGlobalKeydown));
    onUnmounted(() => window.removeEventListener("keydown", onGlobalKeydown));

    function reset() {
        activeIndex.value = -1;
    }

    return { activeIndex, handleKey, setEnterHandler, reset };
}
