import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import { activeGroup } from "./nav";

const STORAGE_KEY = "st:nav-last-path";

function load(): Record<string, string> {
    try {
        const raw = sessionStorage.getItem(STORAGE_KEY);
        const parsed = raw ? JSON.parse(raw) : null;
        return parsed && typeof parsed === "object" ? parsed : {};
    } catch {
        return {};
    }
}

/** Last full path visited per nav group, so tabs return where you left off. */
const lastPath = ref<Record<string, string>>(load());
let tracking = false;

export function useNavMemory() {
    if (!tracking) {
        tracking = true;
        const route = useRoute();
        watch(
            () => route.fullPath,
            () => {
                const group = activeGroup(route);
                if (!group) return;
                lastPath.value = { ...lastPath.value, [group.key]: route.fullPath };
                try {
                    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(lastPath.value));
                } catch {
                    /* storage unavailable: memory still works for this session */
                }
            },
            { immediate: true },
        );
    }
    return { lastPath };
}
