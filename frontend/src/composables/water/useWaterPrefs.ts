import { computed, ref } from "vue";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";

const STORAGE_KEY = "simple-track-water-prefs";

export type WaterDisplayUnit = "oz" | "ml";

export type WaterPrefsSnapshot = {
    goalOz: number;
    displayUnit: WaterDisplayUnit;
};

const DEFAULTS: WaterPrefsSnapshot = {
    goalOz: 64,
    displayUnit: "oz",
};

const ML_PER_OZ = 29.5735295625;

export function useWaterPrefs() {
    const goalOz = ref(DEFAULTS.goalOz);
    const displayUnit = ref<WaterDisplayUnit>(DEFAULTS.displayUnit);

    const key = computed(() => STORAGE_KEY);
    const sync = useWebStorageJsonSync({
        key,
        watchSources: [goalOz, displayUnit],
        getSnapshot: (): WaterPrefsSnapshot => ({
            goalOz: goalOz.value,
            displayUnit: displayUnit.value,
        }),
        tryRestore: (parsed) => {
            if (
                typeof parsed.goalOz === "number" &&
                Number.isFinite(parsed.goalOz) &&
                parsed.goalOz > 0
            ) {
                goalOz.value = parsed.goalOz;
            }
            if (parsed.displayUnit === "ml" || parsed.displayUnit === "oz") {
                displayUnit.value = parsed.displayUnit;
            }
            return true;
        },
    });

    sync.restore();
    sync.setSaveEnabled(true);

    function formatVolumeFromOz(oz: number): {
        value: number;
        unit: WaterDisplayUnit;
    } {
        if (displayUnit.value === "ml") {
            return {
                value: Math.round(oz * ML_PER_OZ * 10) / 10,
                unit: "ml",
            };
        }
        return { value: Math.round(oz * 10) / 10, unit: "oz" };
    }

    function ozFromDisplayAmount(
        amount: number,
        unit: WaterDisplayUnit,
    ): number {
        if (unit === "ml") return amount / ML_PER_OZ;
        return amount;
    }

    return {
        goalOz,
        displayUnit,
        formatVolumeFromOz,
        ozFromDisplayAmount,
    };
}
