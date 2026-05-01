<script setup lang="ts">
import { computed } from "vue";
import MacroBar from "./MacroBar.vue";
import LogWaterDialog from "~/pages/diet/components/dialog/LogWaterDialog.vue";
import { useWaterLogs } from "~/api/tracking/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { ymdForDayOffset } from "~/utils/dateUtil";

const props = defineProps<{
    dateOffset: number;
}>();

const dateStr = computed(() => ymdForDayOffset(props.dateOffset));
const { goalOz, displayUnit, formatVolumeFromOz } = useWaterPrefs();
const { data: logs, isPending, isError, error } = useWaterLogs(dateStr);

const totalOz = computed(() =>
    (logs.value ?? []).reduce((s, row) => s + row.amount_oz, 0),
);

const totalDisplay = computed(
    () => formatVolumeFromOz(totalOz.value).value,
);
const plannedDisplay = computed(
    () => formatVolumeFromOz(goalOz.value).value,
);
const unitSuffix = computed(() => ` ${displayUnit.value}`);

async function openLog() {
    await dialogManager.custom<boolean>({
        title: "Log water",
        component: LogWaterDialog,
        componentProps: { dateStr: dateStr.value },
    });
}
</script>

<template>
    <section
        class="flex flex-col gap-2 rounded-md border border-zinc-700 bg-zinc-900/60 p-3"
    >
        <div v-if="isError" class="text-xs text-red-400">
            {{ error?.message ?? "Failed to load water" }}
        </div>
        <p v-else-if="isPending" class="m-0 text-xs text-zinc-500">
            Loading water…
        </p>
        <div v-else class="flex w-full flex-row">
            <MacroBar
                type="water"
                class="min-w-0 flex-1"
                :total="totalDisplay"
                :planned="plannedDisplay"
                :unit-suffix="unitSuffix"
            />
        </div>
        <div class="flex flex-wrap justify-end gap-2">
            <router-link
                :to="{ name: 'diet-water-presets' }"
                class="rounded-md border border-zinc-600 bg-zinc-800 px-3 py-1.5 text-xs font-semibold text-zinc-200 hover:bg-zinc-700"
            >
                Sizes & goal
            </router-link>
            <button
                type="button"
                class="rounded-md border border-sky-700/60 bg-sky-950/40 px-3 py-1.5 text-xs font-semibold text-sky-100 hover:bg-sky-950/70"
                @click="openLog"
            >
                Log water
            </button>
        </div>
    </section>
</template>
