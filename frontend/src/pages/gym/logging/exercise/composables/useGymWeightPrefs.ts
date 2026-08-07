import { computed, ref } from "vue";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";


const STORAGE_KEY = "simpletracker:gym-weight-prefs";
const weightIncrement = ref(2.5);
let initialized = false;

export function useGymWeightPrefs() {
    if (!initialized) {
        const key = computed(() => STORAGE_KEY);
        const sync = useWebStorageJsonSync({
            key,
            watchSources: [weightIncrement],
            getSnapshot: () => ({ weightIncrement: weightIncrement.value }),
            tryRestore: (parsed) => {
                if (parsed.weightIncrement === 2.5) {
                    weightIncrement.value = parsed.weightIncrement;
                }
                return true;
            },
        });
        sync.restore();
        sync.setSaveEnabled(true);
        initialized = true;
    }

    const setWeightIncrement = (value: number) => {
        weightIncrement.value = value;
    };

    return { weightIncrement, setWeightIncrement };
}
