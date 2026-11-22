import { onScopeDispose, ref, toValue, watch, type MaybeRefOrGetter, type Ref } from "vue";

/**
 * True only once `source` has been true for `delayMs`, false again the moment it clears.
 * Use it for loading UI so a fast load never flashes a skeleton (PRINCIPLES §6).
 */
export function useDelayedFlag(source: MaybeRefOrGetter<boolean>, delayMs = 150): Ref<boolean> {
    const flag = ref(false);
    let timer: ReturnType<typeof setTimeout> | undefined;
    const clear = () => {
        if (timer !== undefined) clearTimeout(timer);
        timer = undefined;
    };

    watch(
        () => toValue(source),
        (on) => {
            clear();
            if (!on) {
                flag.value = false;
                return;
            }
            timer = setTimeout(() => {
                flag.value = true;
                timer = undefined;
            }, delayMs);
        },
        { immediate: true },
    );
    onScopeDispose(clear);
    return flag;
}
