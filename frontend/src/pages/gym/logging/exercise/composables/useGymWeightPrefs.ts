import { computed, ref } from "vue";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";

export type GymWeightIncrement = 2.5 | 5;

const STORAGE_KEY = "simpletracker:gym-weight-prefs";
const weightIncrement = ref<GymWeightIncrement>(5);
let initialized = false;

export function useGymWeightPrefs() {
    if (!initialized) {
        const key = computed(() => STORAGE_KEY);
        const sync = useWebStorageJsonSync({
            key,
            watchSources: [weightIncrement],
            getSnapshot: () => ({ weightIncrement: weightIncrement.value }),
            tryRestore: (parsed) => {
                if (
                    parsed.weightIncrement === 2.5 ||
                    parsed.weightIncrement === 5
                ) {
                    weightIncrement.value = parsed.weightIncrement;
                }
                return true;
            },
        });
        sync.restore();
        sync.setSaveEnabled(true);
        initialized = true;
    }

    const setWeightIncrement = (value: GymWeightIncrement) => {
        weightIncrement.value = value;
    };

    return { weightIncrement, setWeightIncrement };
}
